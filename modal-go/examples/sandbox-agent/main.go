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

	app, err := mc.Apps.FromName(ctx, "libmodal-example", &modal.AppFromNameOptions{CreateIfMissing: true})
	if err != nil {
		log.Fatalf("Failed to get or create App: %v", err)
	}
	image := mc.Images.FromRegistry("alpine:3.21", nil).DockerfileCommands([]string{
		"RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep",
		"RUN curl -fsSL https://claude.ai/install.sh | bash",
		"ENV PATH=/root/.local/bin:$PATH USE_BUILTIN_RIPGREP=0",
	}, nil)

	sb, err := mc.Sandboxes.Create(ctx, app, image, nil)
	if err != nil {
		log.Fatalf("Failed to create Sandbox: %v", err)
	}
	fmt.Println("Started Sandbox:", sb.SandboxId)

	defer func() {
		if err := sb.Terminate(context.Background()); err != nil {
			log.Fatalf("Failed to terminate Sandbox %s: %v", sb.SandboxId, err)
		}
	}()

	repoUrl := "https://github.com/modal-labs/libmodal"
	git, err := sb.Exec(ctx, []string{"git", "clone", repoUrl, "/repo"}, modal.ExecOptions{})
	if err != nil {
		log.Fatalf("Failed to execute git clone: %v", err)
	}
	_, err = git.Wait(ctx)
	if err != nil {
		log.Fatalf("Git clone failed: %v", err)
	}
	fmt.Printf("Cloned '%s' into /repo.\n", repoUrl)

	claudeCmd := []string{
		"claude",
		"-p",
		"Summarize what this repository is about. Don't modify any code or files.",
	}
	fmt.Println("\nRunning command:", claudeCmd)

	secret, err := mc.Secrets.FromName(ctx, "libmodal-anthropic-secret", &modal.SecretFromNameOptions{
		RequiredKeys: []string{"ANTHROPIC_API_KEY"},
	})
	if err != nil {
		log.Fatalf("Failed to get secret: %v", err)
	}

	claude, err := sb.Exec(ctx, claudeCmd, modal.ExecOptions{
		PTY:     true, // Adding a PTY is important, since Claude requires it!
		Secrets: []*modal.Secret{secret},
		Workdir: "/repo",
		Stdout:  modal.Pipe,
		Stderr:  modal.Pipe,
	})
	if err != nil {
		log.Fatalf("Failed to execute claude command: %v", err)
	}
	_, err = claude.Wait(ctx)
	if err != nil {
		log.Fatalf("Claude command failed: %v", err)
	}

	fmt.Printf("\nAgent stdout:\n\n")
	stdout, err := io.ReadAll(claude.Stdout)
	if err != nil {
		log.Fatalf("Failed to read stdout: %v", err)
	}
	fmt.Print(string(stdout))

	stderr, err := io.ReadAll(claude.Stderr)
	if err != nil {
		log.Fatalf("Failed to read stderr: %v", err)
	}
	if len(stderr) > 0 {
		fmt.Println("Agent stderr:", string(stderr))
	}
}
