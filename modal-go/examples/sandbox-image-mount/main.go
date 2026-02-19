// This example shows how to mount Images in the Sandbox filesystem and take snapshots
// of them.
//
// The feature is still experimental in the sense that the API is subject to change.
//
// High level, it allows you to:
// - Mount any Modal Image at a specific directory within the Sandbox filesystem.
// - Take a snapshot of that directory, which will create a new Modal Image with
//   the updated contents of the directory.
//
// You can only snapshot directories that have previously been mounted using
// `Sandbox.ExperimentalMountImage`. If you want to mount an empty directory,
// you can pass nil as the image parameter.
//
// For example, you can use this to mount user specific dependencies into a running
// Sandbox, that is started with a base Image with shared system dependencies. This
// way, you can update system dependencies and user projects independently.

package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/modal-labs/libmodal/modal-go"
)

func main() {
	ctx := context.Background()
	mc, err := modal.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	app, err := mc.Apps.FromName(ctx, "libmodal-example", &modal.AppFromNameParams{CreateIfMissing: true})
	if err != nil {
		log.Fatalf("Failed to get or create App: %v", err)
	}

	// The base Image you use for the Sandbox must have a /usr/bin/mount binary.
	baseImage := mc.Images.FromRegistry("debian:12-slim", nil).DockerfileCommands([]string{
		"RUN apt-get update && apt-get install -y git",
	}, nil)

	sb, err := mc.Sandboxes.Create(ctx, app, baseImage, nil)
	if err != nil {
		log.Fatalf("Failed to create Sandbox: %v", err)
	}
	sbFromID, err := mc.Sandboxes.FromID(ctx, sb.SandboxID)
	if err != nil {
		log.Fatalf("Failed to create Sandbox: %v", err)
	}
	defer func() {
		if err := sbFromID.Terminate(context.Background()); err != nil {
			log.Fatalf("Failed to terminate Sandbox %s: %v", sb.SandboxID, err)
		}
	}()
	fmt.Printf("Started first Sandbox: %s\n", sb.SandboxID)

	// You must mount an Image at a directory in the Sandbox filesystem before you
	// can snapshot it. You can pass nil as the image parameter to mount an
	// empty directory.
	//
	// The target directory must exist before you can mount it:
	mkdirProc, err := sb.Exec(ctx, []string{"mkdir", "-p", "/repo"}, nil)
	if err != nil {
		log.Fatalf("Failed to exec mkdir: %v", err)
	}
	if exitCode, err := mkdirProc.Wait(ctx); err != nil || exitCode != 0 {
		log.Fatalf("Failed to wait for mkdir: exit code: %d, err: %v", exitCode, err)
	}
	if err := sb.ExperimentalMountImage(ctx, "/repo", nil); err != nil {
		log.Fatalf("Failed to mount image: %v", err)
	}

	gitClone, err := sb.Exec(ctx, []string{
		"git",
		"clone",
		"https://github.com/modal-labs/libmodal.git",
		"/repo",
	}, nil)
	if err != nil {
		log.Fatalf("Failed to exec git clone: %v", err)
	}
	if exitCode, err := gitClone.Wait(ctx); err != nil || exitCode != 0 {
		log.Fatalf("Failed to wait for git clone: exit code: %d, err: %v", exitCode, err)
	}

	repoSnapshot, err := sb.ExperimentalSnapshotDirectory(ctx, "/repo")
	if err != nil {
		log.Fatalf("Failed to snapshot directory: %v", err)
	}
	fmt.Printf("Took a snapshot of the /repo directory, Image ID: %s\n", repoSnapshot.ImageID)

	if err := sb.Terminate(ctx); err != nil {
		log.Fatalf("Failed to terminate Sandbox: %v", err)
	}
	if err := sb.Detach(); err != nil {
		log.Fatalf("Failed to detach Sandbox %s: %v", sb.SandboxID, err)
	}

	// Start a new Sandbox, and mount the repo directory:
	sb2, err := mc.Sandboxes.Create(ctx, app, baseImage, nil)
	if err != nil {
		log.Fatalf("Failed to create second Sandbox: %v", err)
	}
	defer func() {
		if err := sb2.Terminate(context.Background()); err != nil {
			log.Fatalf("Failed to terminate Sandbox %s: %v", sb2.SandboxID, err)
		}
	}()
	fmt.Printf("Started second Sandbox: %s\n", sb2.SandboxID)

	mkdirProc2, err := sb2.Exec(ctx, []string{"mkdir", "-p", "/repo"}, nil)
	if err != nil {
		log.Fatalf("Failed to exec mkdir in sb2: %v", err)
	}
	if exitCode, err := mkdirProc2.Wait(ctx); err != nil || exitCode != 0 {
		log.Fatalf("Failed to wait for mkdir in sb2: exit code: %d, err: %v", exitCode, err)
	}
	if err := sb2.ExperimentalMountImage(ctx, "/repo", repoSnapshot); err != nil {
		log.Fatalf("Failed to mount snapshot in sb2: %v", err)
	}

	repoLs, err := sb2.Exec(ctx, []string{"ls", "/repo"}, nil)
	if err != nil {
		log.Fatalf("Failed to exec ls: %v", err)
	}
	if exitCode, err := repoLs.Wait(ctx); err != nil || exitCode != 0 {
		log.Fatalf("Failed to wait for ls: exit code: %d, err: %v", exitCode, err)
	}
	output, err := io.ReadAll(repoLs.Stdout)
	if err != nil {
		log.Fatalf("Failed to read stdout: %v", err)
	}
	fmt.Printf("Contents of /repo directory in new Sandbox sb2:\n%s", output)

	if err := mc.Images.Delete(ctx, repoSnapshot.ImageID, nil); err != nil {
		log.Fatalf("Failed to delete snapshot image: %v", err)
	}
}
