package modal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"iter"
	"strings"
	"sync"
	"time"

	"github.com/djherbis/buffer"
	"github.com/djherbis/nio/v3"
	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SandboxService provides Sandbox related operations.
type SandboxService interface {
	Create(ctx context.Context, app *App, image *Image, params *SandboxCreateParams) (*Sandbox, error)
	FromID(ctx context.Context, sandboxID string) (*Sandbox, error)
	FromName(ctx context.Context, appName, name string, params *SandboxFromNameParams) (*Sandbox, error)
	List(ctx context.Context, params *SandboxListParams) (iter.Seq2[*Sandbox, error], error)
}

type sandboxServiceImpl struct{ client *Client }

// SandboxCreateParams are options for creating a Modal Sandbox.
type SandboxCreateParams struct {
	CPU               float64                      // CPU request in fractional, physical cores.
	CPULimit          float64                      // Hard limit in fractional, physical CPU cores. Zero means no limit.
	MemoryMib         int                          // Memory request in MiB.
	MemoryLimitMib    int                          // Hard memory limit in MiB. Zero means no limit.
	GPU               string                       // GPU reservation for the Sandbox (e.g. "A100", "T4:2", "A100-80GB:4").
	Timeout           time.Duration                // Maximum lifetime for the Sandbox.
	IdleTimeout       time.Duration                // The amount of time that a Sandbox can be idle before being terminated.
	Workdir           string                       // Working directory of the Sandbox.
	Command           []string                     // Command to run in the Sandbox on startup.
	Env               map[string]string            // Environment variables to set in the Sandbox.
	Secrets           []*Secret                    // Secrets to inject into the Sandbox as environment variables.
	Volumes           map[string]*Volume           // Mount points for Volumes.
	CloudBucketMounts map[string]*CloudBucketMount // Mount points for cloud buckets.
	PTY               bool                         // Enable a PTY for the Sandbox.
	EncryptedPorts    []int                        // List of encrypted ports to tunnel into the Sandbox, with TLS encryption.
	H2Ports           []int                        // List of encrypted ports to tunnel into the Sandbox, using HTTP/2.
	UnencryptedPorts  []int                        // List of ports to tunnel into the Sandbox without encryption.
	BlockNetwork      bool                         // Whether to block all network access from the Sandbox.
	CIDRAllowlist     []string                     // List of CIDRs the Sandbox is allowed to access. Cannot be used with BlockNetwork.
	Cloud             string                       // Cloud provider to run the Sandbox on.
	Regions           []string                     // Region(s) to run the Sandbox on.
	Verbose           bool                         // Enable verbose logging.
	Proxy             *Proxy                       // Reference to a Modal Proxy to use in front of this Sandbox.
	Name              string                       // Optional name for the Sandbox. Unique within an App.
}

// buildSandboxCreateRequestProto builds a SandboxCreateRequest proto from options.
func buildSandboxCreateRequestProto(appID, imageID string, params SandboxCreateParams) (*pb.SandboxCreateRequest, error) {
	gpuConfig, err := parseGPUConfig(params.GPU)
	if err != nil {
		return nil, err
	}

	if params.Workdir != "" && !strings.HasPrefix(params.Workdir, "/") {
		return nil, fmt.Errorf("the Workdir value must be an absolute path, got: %s", params.Workdir)
	}

	var volumeMounts []*pb.VolumeMount
	if params.Volumes != nil {
		volumeMounts = make([]*pb.VolumeMount, 0, len(params.Volumes))
		for mountPath, volume := range params.Volumes {
			volumeMounts = append(volumeMounts, pb.VolumeMount_builder{
				VolumeId:               volume.VolumeID,
				MountPath:              mountPath,
				AllowBackgroundCommits: true,
				ReadOnly:               volume.IsReadOnly(),
			}.Build())
		}
	}

	var cloudBucketMounts []*pb.CloudBucketMount
	if params.CloudBucketMounts != nil {
		cloudBucketMounts = make([]*pb.CloudBucketMount, 0, len(params.CloudBucketMounts))
		for mountPath, mount := range params.CloudBucketMounts {
			proto, err := mount.toProto(mountPath)
			if err != nil {
				return nil, err
			}
			cloudBucketMounts = append(cloudBucketMounts, proto)
		}
	}

	var ptyInfo *pb.PTYInfo
	if params.PTY {
		ptyInfo = defaultSandboxPTYInfo()
	}

	var openPorts []*pb.PortSpec
	for _, port := range params.EncryptedPorts {
		openPorts = append(openPorts, pb.PortSpec_builder{
			Port:        uint32(port),
			Unencrypted: false,
		}.Build())
	}
	for _, port := range params.H2Ports {
		openPorts = append(openPorts, pb.PortSpec_builder{
			Port:        uint32(port),
			Unencrypted: false,
			TunnelType:  pb.TunnelType_TUNNEL_TYPE_H2.Enum(),
		}.Build())
	}
	for _, port := range params.UnencryptedPorts {
		openPorts = append(openPorts, pb.PortSpec_builder{
			Port:        uint32(port),
			Unencrypted: true,
		}.Build())
	}

	var portSpecs *pb.PortSpecs
	if len(openPorts) > 0 {
		portSpecs = pb.PortSpecs_builder{
			Ports: openPorts,
		}.Build()
	}

	secretIds := []string{}
	for _, secret := range params.Secrets {
		if secret != nil {
			secretIds = append(secretIds, secret.SecretID)
		}
	}

	var networkAccess *pb.NetworkAccess
	if params.BlockNetwork {
		if len(params.CIDRAllowlist) > 0 {
			return nil, fmt.Errorf("CIDRAllowlist cannot be used when BlockNetwork is enabled")
		}
		networkAccess = pb.NetworkAccess_builder{
			NetworkAccessType: pb.NetworkAccess_BLOCKED,
			AllowedCidrs:      []string{},
		}.Build()
	} else if len(params.CIDRAllowlist) > 0 {
		networkAccess = pb.NetworkAccess_builder{
			NetworkAccessType: pb.NetworkAccess_ALLOWLIST,
			AllowedCidrs:      params.CIDRAllowlist,
		}.Build()
	} else {
		networkAccess = pb.NetworkAccess_builder{
			NetworkAccessType: pb.NetworkAccess_OPEN,
			AllowedCidrs:      []string{},
		}.Build()
	}

	schedulerPlacement := pb.SchedulerPlacement_builder{Regions: params.Regions}.Build()

	var proxyID *string
	if params.Proxy != nil {
		proxyID = &params.Proxy.ProxyID
	}

	var workdir *string
	if params.Workdir != "" {
		workdir = &params.Workdir
	}

	var idleTimeoutSecs *uint32
	if params.IdleTimeout != 0 {
		v := uint32(params.IdleTimeout.Seconds())
		idleTimeoutSecs = &v
	}

	var milliCPU, milliCPUMax *uint32
	if params.CPU == 0 && params.CPULimit > 0 {
		return nil, fmt.Errorf("must also specify non-zero CPU request when CPULimit is specified")
	}
	if params.CPU != 0 {
		if params.CPU <= 0 {
			return nil, fmt.Errorf("the CPU request (%f) must be a positive number", params.CPU)
		}
		v := uint32(1000 * params.CPU)
		milliCPU = &v
		if params.CPULimit > 0 {
			if params.CPULimit < params.CPU {
				return nil, fmt.Errorf("the CPU request (%f) cannot be higher than CPULimit (%f)", params.CPU, params.CPULimit)
			}
			vMax := uint32(1000 * params.CPULimit)
			milliCPUMax = &vMax
		}
	}

	var memoryMb, memoryMbMax uint32
	if params.MemoryMib == 0 && params.MemoryLimitMib > 0 {
		return nil, fmt.Errorf("must also specify non-zero MemoryMib request when MemoryLimitMib is specified")
	}
	if params.MemoryMib != 0 {
		if params.MemoryMib <= 0 {
			return nil, fmt.Errorf("the MemoryMib request (%d) must be a positive number", params.MemoryMib)
		}
		memoryMb = uint32(params.MemoryMib)
		if params.MemoryLimitMib > 0 {
			if params.MemoryLimitMib < params.MemoryMib {
				return nil, fmt.Errorf("the MemoryMib request (%d) cannot be higher than MemoryLimitMib (%d)", params.MemoryMib, params.MemoryLimitMib)
			}
			memoryMbMax = uint32(params.MemoryLimitMib)
		}
	}

	resourcesBuilder := pb.Resources_builder{
		GpuConfig: gpuConfig,
	}
	if milliCPU != nil {
		resourcesBuilder.MilliCpu = *milliCPU
	}
	if milliCPUMax != nil {
		resourcesBuilder.MilliCpuMax = *milliCPUMax
	}
	if memoryMb > 0 {
		resourcesBuilder.MemoryMb = memoryMb
	}
	if memoryMbMax > 0 {
		resourcesBuilder.MemoryMbMax = memoryMbMax
	}

	return pb.SandboxCreateRequest_builder{
		AppId: appID,
		Definition: pb.Sandbox_builder{
			EntrypointArgs:     params.Command,
			ImageId:            imageID,
			SecretIds:          secretIds,
			TimeoutSecs:        uint32(params.Timeout.Seconds()),
			IdleTimeoutSecs:    idleTimeoutSecs,
			Workdir:            workdir,
			NetworkAccess:      networkAccess,
			Resources:          resourcesBuilder.Build(),
			VolumeMounts:       volumeMounts,
			CloudBucketMounts:  cloudBucketMounts,
			PtyInfo:            ptyInfo,
			OpenPorts:          portSpecs,
			CloudProviderStr:   params.Cloud,
			SchedulerPlacement: schedulerPlacement,
			Verbose:            params.Verbose,
			ProxyId:            proxyID,
			Name:               &params.Name,
		}.Build(),
	}.Build(), nil
}

// Create creates a new Sandbox in the App with the specified Image and options.
func (s *sandboxServiceImpl) Create(ctx context.Context, app *App, image *Image, params *SandboxCreateParams) (*Sandbox, error) {
	if params == nil {
		params = &SandboxCreateParams{}
	}

	image, err := image.Build(ctx, app)
	if err != nil {
		return nil, err
	}

	mergedSecrets, err := mergeEnvIntoSecrets(ctx, s.client, &params.Env, &params.Secrets)
	if err != nil {
		return nil, err
	}

	mergedParams := *params
	mergedParams.Secrets = mergedSecrets
	mergedParams.Env = nil // nil'ing Env just to clarify it's not needed anymore

	req, err := buildSandboxCreateRequestProto(app.AppID, image.ImageID, mergedParams)
	if err != nil {
		return nil, err
	}

	createResp, err := s.client.cpClient.SandboxCreate(ctx, req)
	if err != nil {
		if status, ok := status.FromError(err); ok && status.Code() == codes.AlreadyExists {
			return nil, AlreadyExistsError{Exception: status.Message()}
		}
		return nil, err
	}

	return newSandbox(s.client, createResp.GetSandboxId()), nil
}

// StdioBehavior defines how the standard input/output/error streams should behave.
type StdioBehavior string

const (
	// Pipe allows the Sandbox to pipe the streams.
	Pipe StdioBehavior = "pipe"
	// Ignore ignores the streams, meaning they will not be available.
	Ignore StdioBehavior = "ignore"
)

// Tunnel represents a port forwarded from within a running Modal Sandbox.
type Tunnel struct {
	Host            string // The public hostname for the tunnel
	Port            int    // The public port for the tunnel
	UnencryptedHost string // The unencrypted hostname (if applicable)
	UnencryptedPort int    // The unencrypted port (if applicable)
}

// URL gets the public HTTPS URL of the forwarded port.
func (t *Tunnel) URL() string {
	if t.Port == 443 {
		return fmt.Sprintf("https://%s", t.Host)
	}
	return fmt.Sprintf("https://%s:%d", t.Host, t.Port)
}

// TLSSocket gets the public TLS socket as a (host, port) tuple.
func (t *Tunnel) TLSSocket() (string, int) {
	return t.Host, t.Port
}

// TCPSocket gets the public TCP socket as a (host, port) tuple.
func (t *Tunnel) TCPSocket() (string, int, error) {
	if t.UnencryptedHost == "" || t.UnencryptedPort == 0 {
		return "", 0, InvalidError{Exception: "This tunnel is not configured for unencrypted TCP."}
	}
	return t.UnencryptedHost, t.UnencryptedPort, nil
}

// Sandbox represents a Modal Sandbox, which can run commands and manage
// input/output streams for a remote process.
type Sandbox struct {
	SandboxID string
	Stdin     io.WriteCloser
	Stdout    io.ReadCloser
	Stderr    io.ReadCloser

	taskID  string
	tunnels map[int]*Tunnel

	client *Client
}

func defaultSandboxPTYInfo() *pb.PTYInfo {
	return pb.PTYInfo_builder{
		Enabled:                true,
		WinszRows:              24,
		WinszCols:              80,
		EnvTerm:                "xterm-256color",
		EnvColorterm:           "truecolor",
		PtyType:                pb.PTYInfo_PTY_TYPE_SHELL,
		NoTerminateOnIdleStdin: true,
	}.Build()
}

// newSandbox creates a new Sandbox object from ID.
func newSandbox(client *Client, sandboxID string) *Sandbox {
	sb := &Sandbox{SandboxID: sandboxID, client: client}
	sb.Stdin = inputStreamSb(client.cpClient, sandboxID)
	sb.Stdout = outputStreamSb(client.cpClient, sandboxID, pb.FileDescriptor_FILE_DESCRIPTOR_STDOUT)
	sb.Stderr = outputStreamSb(client.cpClient, sandboxID, pb.FileDescriptor_FILE_DESCRIPTOR_STDERR)
	return sb
}

// FromID returns a running Sandbox object from an ID.
func (s *sandboxServiceImpl) FromID(ctx context.Context, sandboxID string) (*Sandbox, error) {
	_, err := s.client.cpClient.SandboxWait(ctx, pb.SandboxWaitRequest_builder{
		SandboxId: sandboxID,
		Timeout:   0,
	}.Build())
	if status, ok := status.FromError(err); ok && status.Code() == codes.NotFound {
		return nil, NotFoundError{fmt.Sprintf("Sandbox with id: '%s' not found", sandboxID)}
	}
	if err != nil {
		return nil, err
	}
	return newSandbox(s.client, sandboxID), nil
}

// SandboxFromNameParams are options for finding deployed Sandbox objects by name.
type SandboxFromNameParams struct {
	Environment string
}

// FromName gets a running Sandbox by name from a deployed App.
//
// Raises a NotFoundError if no running Sandbox is found with the given name.
// A Sandbox's name is the `Name` argument passed to `App.CreateSandbox`.
func (s *sandboxServiceImpl) FromName(ctx context.Context, appName, name string, params *SandboxFromNameParams) (*Sandbox, error) {
	if params == nil {
		params = &SandboxFromNameParams{}
	}

	resp, err := s.client.cpClient.SandboxGetFromName(ctx, pb.SandboxGetFromNameRequest_builder{
		SandboxName:     name,
		AppName:         appName,
		EnvironmentName: environmentName(params.Environment, s.client.profile),
	}.Build())
	if err != nil {
		if status, ok := status.FromError(err); ok && status.Code() == codes.NotFound {
			return nil, NotFoundError{Exception: fmt.Sprintf("Sandbox with name '%s' not found in pp '%s'", name, appName)}
		}
		return nil, err
	}

	return newSandbox(s.client, resp.GetSandboxId()), nil
}

// SandboxExecParams defines options for executing commands in a Sandbox.
type SandboxExecParams struct {
	// Stdout defines whether to pipe or ignore standard output.
	Stdout StdioBehavior
	// Stderr defines whether to pipe or ignore standard error.
	Stderr StdioBehavior
	// Workdir is the working directory to run the command in.
	Workdir string
	// Timeout is the timeout for command execution. Defaults to 0 (no timeout).
	Timeout time.Duration
	// Environment variables to set for the command.
	Env map[string]string
	// Secrets to inject as environment variables for the command.
	Secrets []*Secret
	// PTY defines whether to enable a PTY for the command.
	PTY bool
}

// buildContainerExecRequestProto builds a ContainerExecRequest proto from command and options.
func buildContainerExecRequestProto(taskID string, command []string, params SandboxExecParams) (*pb.ContainerExecRequest, error) {
	var workdir *string
	if params.Workdir != "" {
		workdir = &params.Workdir
	}
	secretIds := []string{}
	for _, secret := range params.Secrets {
		if secret != nil {
			secretIds = append(secretIds, secret.SecretID)
		}
	}

	var ptyInfo *pb.PTYInfo
	if params.PTY {
		ptyInfo = defaultSandboxPTYInfo()
	}

	return pb.ContainerExecRequest_builder{
		TaskId:      taskID,
		Command:     command,
		Workdir:     workdir,
		TimeoutSecs: uint32(params.Timeout.Seconds()),
		SecretIds:   secretIds,
		PtyInfo:     ptyInfo,
	}.Build(), nil
}

// Exec runs a command in the Sandbox and returns text streams.
func (sb *Sandbox) Exec(ctx context.Context, command []string, params *SandboxExecParams) (*ContainerProcess, error) {
	if params == nil {
		params = &SandboxExecParams{}
	}

	if err := sb.ensureTaskID(ctx); err != nil {
		return nil, err
	}

	mergedSecrets, err := mergeEnvIntoSecrets(ctx, sb.client, &params.Env, &params.Secrets)
	if err != nil {
		return nil, err
	}

	mergedParams := *params
	mergedParams.Secrets = mergedSecrets
	mergedParams.Env = nil // nil'ing Env just to clarify it's not needed anymore

	req, err := buildContainerExecRequestProto(sb.taskID, command, mergedParams)
	if err != nil {
		return nil, err
	}
	resp, err := sb.client.cpClient.ContainerExec(ctx, req)
	if err != nil {
		return nil, err
	}
	return newContainerProcess(sb.client.cpClient, resp.GetExecId(), *params), nil
}

// Open opens a file in the Sandbox filesystem.
// The mode parameter follows the same conventions as os.OpenFile:
// "r" for read-only, "w" for write-only (truncates), "a" for append, etc.
func (sb *Sandbox) Open(ctx context.Context, path, mode string) (*SandboxFile, error) {
	if err := sb.ensureTaskID(ctx); err != nil {
		return nil, err
	}

	_, resp, err := runFilesystemExec(ctx, sb.client.cpClient, pb.ContainerFilesystemExecRequest_builder{
		FileOpenRequest: pb.ContainerFileOpenRequest_builder{
			Path: path,
			Mode: mode,
		}.Build(),
		TaskId: sb.taskID,
	}.Build(), nil)

	if err != nil {
		return nil, err
	}

	return &SandboxFile{
		fileDescriptor: resp.GetFileDescriptor(),
		taskID:         sb.taskID,
		cpClient:       sb.client.cpClient,
	}, nil
}

func (sb *Sandbox) ensureTaskID(ctx context.Context) error {
	if sb.taskID == "" {
		resp, err := sb.client.cpClient.SandboxGetTaskId(ctx, pb.SandboxGetTaskIdRequest_builder{
			SandboxId: sb.SandboxID,
		}.Build())
		if err != nil {
			return err
		}
		if resp.GetTaskId() == "" {
			return fmt.Errorf("Sandbox %s does not have a task ID, it may not be running", sb.SandboxID)
		}
		if resp.GetTaskResult() != nil {
			return fmt.Errorf("Sandbox %s has already completed with result: %v", sb.SandboxID, resp.GetTaskResult())
		}
		sb.taskID = resp.GetTaskId()
	}
	return nil
}

// Terminate stops the Sandbox.
func (sb *Sandbox) Terminate(ctx context.Context) error {
	_, err := sb.client.cpClient.SandboxTerminate(ctx, pb.SandboxTerminateRequest_builder{
		SandboxId: sb.SandboxID,
	}.Build())
	if err != nil {
		return err
	}
	sb.taskID = ""
	return nil
}

// Wait blocks until the Sandbox exits.
func (sb *Sandbox) Wait(ctx context.Context) (int, error) {
	for {
		resp, err := sb.client.cpClient.SandboxWait(ctx, pb.SandboxWaitRequest_builder{
			SandboxId: sb.SandboxID,
			Timeout:   10,
		}.Build())
		if err != nil {
			return 0, err
		}
		if resp.GetResult() != nil {
			returnCode := getReturnCode(resp.GetResult())
			if returnCode != nil {
				return *returnCode, nil
			}
			return 0, nil
		}
	}
}

// Tunnels gets Tunnel metadata for the Sandbox.
// Returns SandboxTimeoutError if the tunnels are not available after the timeout.
// Returns a map of Tunnel objects keyed by the container port.
func (sb *Sandbox) Tunnels(ctx context.Context, timeout time.Duration) (map[int]*Tunnel, error) {
	if sb.tunnels != nil {
		return sb.tunnels, nil
	}

	resp, err := sb.client.cpClient.SandboxGetTunnels(ctx, pb.SandboxGetTunnelsRequest_builder{
		SandboxId: sb.SandboxID,
		Timeout:   float32(timeout.Seconds()),
	}.Build())
	if err != nil {
		return nil, err
	}

	if resp.GetResult() != nil && resp.GetResult().GetStatus() == pb.GenericResult_GENERIC_STATUS_TIMEOUT {
		return nil, SandboxTimeoutError{Exception: "Sandbox operation timed out"}
	}

	sb.tunnels = make(map[int]*Tunnel)
	for _, t := range resp.GetTunnels() {
		sb.tunnels[int(t.GetContainerPort())] = &Tunnel{
			Host:            t.GetHost(),
			Port:            int(t.GetPort()),
			UnencryptedHost: t.GetUnencryptedHost(),
			UnencryptedPort: int(t.GetUnencryptedPort()),
		}
	}

	return sb.tunnels, nil
}

// SnapshotFilesystem takes a snapshot of the Sandbox's filesystem.
// Returns an Image object which can be used to spawn a new Sandbox with the same filesystem.
func (sb *Sandbox) SnapshotFilesystem(ctx context.Context, timeout time.Duration) (*Image, error) {
	resp, err := sb.client.cpClient.SandboxSnapshotFs(ctx, pb.SandboxSnapshotFsRequest_builder{
		SandboxId: sb.SandboxID,
		Timeout:   float32(timeout.Seconds()),
	}.Build())
	if err != nil {
		return nil, err
	}

	if resp.GetResult() != nil && resp.GetResult().GetStatus() != pb.GenericResult_GENERIC_STATUS_SUCCESS {
		return nil, ExecutionError{Exception: fmt.Sprintf("Sandbox snapshot failed: %s", resp.GetResult().GetException())}
	}

	if resp.GetImageId() == "" {
		return nil, ExecutionError{Exception: "Sandbox snapshot response missing image ID"}
	}

	return &Image{ImageID: resp.GetImageId(), client: sb.client}, nil
}

// Poll checks if the Sandbox has finished running.
// Returns nil if the Sandbox is still running, else returns the exit code.
func (sb *Sandbox) Poll(ctx context.Context) (*int, error) {
	resp, err := sb.client.cpClient.SandboxWait(ctx, pb.SandboxWaitRequest_builder{
		SandboxId: sb.SandboxID,
		Timeout:   0,
	}.Build())
	if err != nil {
		return nil, err
	}

	return getReturnCode(resp.GetResult()), nil
}

// SetTags sets key-value tags on the Sandbox. Tags can be used to filter results in SandboxList.
func (sb *Sandbox) SetTags(ctx context.Context, tags map[string]string) error {
	tagsList := make([]*pb.SandboxTag, 0, len(tags))
	for k, v := range tags {
		tagsList = append(tagsList, pb.SandboxTag_builder{TagName: k, TagValue: v}.Build())
	}
	_, err := sb.client.cpClient.SandboxTagsSet(ctx, pb.SandboxTagsSetRequest_builder{
		EnvironmentName: environmentName("", sb.client.profile),
		SandboxId:       sb.SandboxID,
		Tags:            tagsList,
	}.Build())
	return err
}

// GetTags fetches any tags (key-value pairs) currently attached to this Sandbox from the server.
func (sb *Sandbox) GetTags(ctx context.Context) (map[string]string, error) {
	resp, err := sb.client.cpClient.SandboxTagsGet(ctx, pb.SandboxTagsGetRequest_builder{
		SandboxId: sb.SandboxID,
	}.Build())
	if err != nil {
		if status, ok := status.FromError(err); ok && status.Code() == codes.InvalidArgument {
			return nil, InvalidError{Exception: status.Message()}
		}
		return nil, err
	}

	tags := make(map[string]string, len(resp.GetTags()))
	for _, tag := range resp.GetTags() {
		tags[tag.GetTagName()] = tag.GetTagValue()
	}
	return tags, nil
}

// SandboxListParams are options for listing Sandboxes.
type SandboxListParams struct {
	AppID       string            // Filter by App ID
	Tags        map[string]string // Only include Sandboxes that have all these tags
	Environment string            // Override environment for this request
}

// List lists Sandboxes for the current environment (or provided App ID), optionally filtered by tags.
func (s *sandboxServiceImpl) List(ctx context.Context, params *SandboxListParams) (iter.Seq2[*Sandbox, error], error) {
	if params == nil {
		params = &SandboxListParams{}
	}

	tagsList := make([]*pb.SandboxTag, 0, len(params.Tags))
	for k, v := range params.Tags {
		tagsList = append(tagsList, pb.SandboxTag_builder{TagName: k, TagValue: v}.Build())
	}

	return func(yield func(*Sandbox, error) bool) {
		var before float64
		for {
			resp, err := s.client.cpClient.SandboxList(ctx, pb.SandboxListRequest_builder{
				AppId:           params.AppID,
				BeforeTimestamp: before,
				EnvironmentName: environmentName(params.Environment, s.client.profile),
				IncludeFinished: false,
				Tags:            tagsList,
			}.Build())
			if err != nil {
				yield(nil, err)
				return
			}
			sandboxes := resp.GetSandboxes()
			if len(sandboxes) == 0 {
				return
			}
			for _, info := range sandboxes {
				if !yield(newSandbox(s.client, info.GetId()), nil) {
					return
				}
			}
			before = sandboxes[len(sandboxes)-1].GetCreatedAt()
		}
	}, nil
}

func getReturnCode(result *pb.GenericResult) *int {
	if result == nil || result.GetStatus() == pb.GenericResult_GENERIC_STATUS_UNSPECIFIED {
		return nil
	}

	// Statuses are converted to exitcodes so we can conform to subprocess API.
	var exitCode int
	switch result.GetStatus() {
	case pb.GenericResult_GENERIC_STATUS_TIMEOUT:
		exitCode = 124
	case pb.GenericResult_GENERIC_STATUS_TERMINATED:
		exitCode = 137
	default:
		exitCode = int(result.GetExitcode())
	}

	return &exitCode
}

// ContainerProcess represents a process running in a Modal container, allowing
// interaction with its standard input/output/error streams.
//
// It is created by executing a command in a Sandbox.
type ContainerProcess struct {
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
	Stderr io.ReadCloser

	execID   string
	cpClient pb.ModalClientClient
}

func newContainerProcess(cpClient pb.ModalClientClient, execID string, params SandboxExecParams) *ContainerProcess {
	stdoutBehavior := Pipe
	stderrBehavior := Pipe
	if params.Stdout != "" {
		stdoutBehavior = params.Stdout
	}
	if params.Stderr != "" {
		stderrBehavior = params.Stderr
	}

	cp := &ContainerProcess{execID: execID, cpClient: cpClient}
	cp.Stdin = inputStreamCp(cpClient, execID)

	cp.Stdout = outputStreamCp(cpClient, execID, pb.FileDescriptor_FILE_DESCRIPTOR_STDOUT)
	if stdoutBehavior == Ignore {
		cp.Stdout.Close()
		cp.Stdout = io.NopCloser(bytes.NewReader(nil))
	}
	cp.Stderr = outputStreamCp(cpClient, execID, pb.FileDescriptor_FILE_DESCRIPTOR_STDERR)
	if stderrBehavior == Ignore {
		cp.Stderr.Close()
		cp.Stderr = io.NopCloser(bytes.NewReader(nil))
	}

	return cp
}

// Wait blocks until the container process exits and returns its exit code.
func (cp *ContainerProcess) Wait(ctx context.Context) (int, error) {
	for {
		resp, err := cp.cpClient.ContainerExecWait(ctx, pb.ContainerExecWaitRequest_builder{
			ExecId:  cp.execID,
			Timeout: 55,
		}.Build())
		if err != nil {
			return 0, err
		}
		if resp.GetCompleted() {
			return int(resp.GetExitCode()), nil
		}
	}
}

func inputStreamSb(cpClient pb.ModalClientClient, sandboxID string) io.WriteCloser {
	return &sbStdin{sandboxID: sandboxID, index: 1, cpClient: cpClient}
}

type sbStdin struct {
	sandboxID string
	cpClient  pb.ModalClientClient

	mu    sync.Mutex // protects index
	index uint32
}

func (sbs *sbStdin) Write(p []byte) (n int, err error) {
	sbs.mu.Lock()
	defer sbs.mu.Unlock()
	index := sbs.index
	sbs.index++
	_, err = sbs.cpClient.SandboxStdinWrite(context.Background(), pb.SandboxStdinWriteRequest_builder{
		SandboxId: sbs.sandboxID,
		Input:     p,
		Index:     index,
	}.Build())
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (sbs *sbStdin) Close() error {
	sbs.mu.Lock()
	defer sbs.mu.Unlock()
	_, err := sbs.cpClient.SandboxStdinWrite(context.Background(), pb.SandboxStdinWriteRequest_builder{
		SandboxId: sbs.sandboxID,
		Index:     sbs.index,
		Eof:       true,
	}.Build())
	return err
}

func inputStreamCp(cpClient pb.ModalClientClient, execID string) io.WriteCloser {
	return &cpStdin{execID: execID, messageIndex: 1, cpClient: cpClient}
}

type cpStdin struct {
	execID       string
	messageIndex uint64
	cpClient     pb.ModalClientClient
}

func (cps *cpStdin) Write(p []byte) (n int, err error) {
	_, err = cps.cpClient.ContainerExecPutInput(context.Background(), pb.ContainerExecPutInputRequest_builder{
		ExecId: cps.execID,
		Input: pb.RuntimeInputMessage_builder{
			Message:      p,
			MessageIndex: cps.messageIndex,
		}.Build(),
	}.Build())
	if err != nil {
		return 0, err
	}
	cps.messageIndex++
	return len(p), nil
}

func (cps *cpStdin) Close() error {
	_, err := cps.cpClient.ContainerExecPutInput(context.Background(), pb.ContainerExecPutInputRequest_builder{
		ExecId: cps.execID,
		Input: pb.RuntimeInputMessage_builder{
			MessageIndex: cps.messageIndex,
			Eof:          true,
		}.Build(),
	}.Build())
	return err
}

func outputStreamSb(cpClient pb.ModalClientClient, sandboxID string, fd pb.FileDescriptor) io.ReadCloser {
	pr, pw := nio.Pipe(buffer.New(64 * 1024))
	go func() {
		defer pw.Close()
		lastIndex := "0-0"
		completed := false
		retries := 10
		for !completed {
			stream, err := cpClient.SandboxGetLogs(context.Background(), pb.SandboxGetLogsRequest_builder{
				SandboxId:      sandboxID,
				FileDescriptor: fd,
				Timeout:        55,
				LastEntryId:    lastIndex,
			}.Build())
			if err != nil {
				if isRetryableGrpc(err) && retries > 0 {
					retries--
					continue
				}
				pw.CloseWithError(fmt.Errorf("error getting output stream: %w", err))
				return
			}
			for {
				batch, err := stream.Recv()
				if err != nil {
					if err != io.EOF {
						if isRetryableGrpc(err) && retries > 0 {
							retries--
						} else {
							pw.CloseWithError(fmt.Errorf("error getting output stream: %w", err))
							return
						}
					}
					break // we need to retry, either from an EOF or gRPC error
				}
				lastIndex = batch.GetEntryId()
				for _, item := range batch.GetItems() {
					// On error, writer has been closed. Still consume the rest of the channel.
					pw.Write([]byte(item.GetData()))
				}
				if batch.GetEof() {
					completed = true
					break
				}
			}
		}
	}()
	return pr
}

func outputStreamCp(cpClient pb.ModalClientClient, execID string, fd pb.FileDescriptor) io.ReadCloser {
	pr, pw := nio.Pipe(buffer.New(64 * 1024))
	go func() {
		defer pw.Close()
		var lastIndex uint64
		completed := false
		retries := 10
		for !completed {
			stream, err := cpClient.ContainerExecGetOutput(context.Background(), pb.ContainerExecGetOutputRequest_builder{
				ExecId:         execID,
				FileDescriptor: fd,
				Timeout:        55,
				GetRawBytes:    true,
				LastBatchIndex: lastIndex,
			}.Build())
			if err != nil {
				if isRetryableGrpc(err) && retries > 0 {
					retries--
					continue
				}
				pw.CloseWithError(fmt.Errorf("error getting output stream: %w", err))
				return
			}
			for {
				batch, err := stream.Recv()
				if err != nil {
					if err != io.EOF {
						if isRetryableGrpc(err) && retries > 0 {
							retries--
						} else {
							pw.CloseWithError(fmt.Errorf("error getting output stream: %w", err))
							return
						}
					}
					break // we need to retry, either from an EOF or gRPC error
				}
				lastIndex = batch.GetBatchIndex()
				for _, item := range batch.GetItems() {
					// On error, writer has been closed. Still consume the rest of the channel.
					pw.Write(item.GetMessageBytes())
				}
				if batch.HasExitCode() {
					completed = true
					break
				}
			}
		}
	}()
	return pr
}
