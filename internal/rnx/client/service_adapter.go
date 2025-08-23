package client

import (
	"context"
	"time"

	pb "joblet/api/gen"
	"joblet/internal/rnx/common"
	"joblet/pkg/config"
)

// ServiceAdapter provides a clean interface between RNX CLI and internal services
type ServiceAdapter struct {
	jobClient pb.JobServiceClient
	config    *config.Config
}

// NewServiceAdapter creates a service adapter for RNX CLI to server communication
func NewServiceAdapter(config *config.Config) (*ServiceAdapter, error) {
	jobletClient, err := common.NewJobClient()
	if err != nil {
		return nil, err
	}

	return &ServiceAdapter{
		jobClient: pb.NewJobServiceClient(jobletClient.GetConn()),
		config:    config,
	}, nil
}

// JobRequest represents a simplified job request for CLI operations
type JobRequest struct {
	Command           string
	Args              []string
	MaxCPU            int32
	CpuCores          string
	MaxMemory         int32
	MaxIOBPS          int32
	Uploads           []*pb.FileUpload
	Schedule          string
	Network           string
	Volumes           []string
	Runtime           string
	Environment       map[string]string
	SecretEnvironment map[string]string
}

// JobResponse represents a simplified job response for CLI operations
type JobResponse struct {
	JobUUID       string
	Command       string
	Args          []string
	Status        string
	StartTime     string
	ScheduledTime string
}

// SubmitJob sends a job execution request to the server and returns the response
func (sa *ServiceAdapter) SubmitJob(ctx context.Context, req *JobRequest) (*JobResponse, error) {
	pbRequest := &pb.RunJobRequest{
		Command:           req.Command,
		Args:              req.Args,
		MaxCpu:            req.MaxCPU,
		CpuCores:          req.CpuCores,
		MaxMemory:         req.MaxMemory,
		MaxIobps:          req.MaxIOBPS,
		Uploads:           req.Uploads,
		Schedule:          req.Schedule,
		Network:           req.Network,
		Volumes:           req.Volumes,
		Runtime:           req.Runtime,
		Environment:       req.Environment,
		SecretEnvironment: req.SecretEnvironment,
	}

	response, err := sa.jobClient.RunJob(ctx, pbRequest)
	if err != nil {
		return nil, err
	}
	return &JobResponse{
		JobUUID:       response.JobUuid,
		Command:       response.Command,
		Args:          response.Args,
		Status:        response.Status,
		StartTime:     response.StartTime,
		ScheduledTime: response.ScheduledTime,
	}, nil
}

// Close terminates the underlying gRPC connection to the server
func (sa *ServiceAdapter) Close() error {
	if closer, ok := sa.jobClient.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}

type WorkflowAdapter struct {
	workflowClient pb.JobServiceClient
}

// NewWorkflowAdapter creates an adapter for workflow execution operations
func NewWorkflowAdapter() (*WorkflowAdapter, error) {
	jobletClient, err := common.NewJobClient()
	if err != nil {
		return nil, err
	}

	return &WorkflowAdapter{
		workflowClient: pb.NewJobServiceClient(jobletClient.GetConn()),
	}, nil
}

// WorkflowRequest represents a workflow execution request
type WorkflowRequest struct {
	WorkflowFile  string
	YamlContent   string
	WorkflowFiles []*pb.FileUpload
	TotalJobs     int32
}

// WorkflowResponse represents a workflow execution response
type WorkflowResponse struct {
	WorkflowUUID string
	Message      string
}

// SubmitWorkflow sends a workflow execution request to the server for multi-job orchestration
func (wa *WorkflowAdapter) SubmitWorkflow(ctx context.Context, req *WorkflowRequest) (*WorkflowResponse, error) {
	pbRequest := &pb.RunWorkflowRequest{
		Workflow:      req.WorkflowFile,
		YamlContent:   req.YamlContent,
		WorkflowFiles: req.WorkflowFiles,
		TotalJobs:     req.TotalJobs,
	}

	response, err := wa.workflowClient.RunWorkflow(ctx, pbRequest)
	if err != nil {
		return nil, err
	}

	return &WorkflowResponse{
		WorkflowUUID: response.WorkflowUuid,
		Message:      "Workflow submitted successfully",
	}, nil
}

// NetworkAdapter handles network operations
type NetworkAdapter struct {
	networkClient pb.NetworkServiceClient
}

// NewNetworkAdapter creates an adapter for network management operations
func NewNetworkAdapter() (*NetworkAdapter, error) {
	jobletClient, err := common.NewJobClient()
	if err != nil {
		return nil, err
	}

	return &NetworkAdapter{
		networkClient: pb.NewNetworkServiceClient(jobletClient.GetConn()),
	}, nil
}

// ListNetworks retrieves all available networks from the server
func (na *NetworkAdapter) ListNetworks(ctx context.Context) (*pb.Networks, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return na.networkClient.ListNetworks(ctx, &pb.EmptyRequest{})
}

// VolumeAdapter handles volume operations
type VolumeAdapter struct {
	volumeClient pb.VolumeServiceClient
}

// NewVolumeAdapter creates an adapter for volume management operations
func NewVolumeAdapter() (*VolumeAdapter, error) {
	client, err := common.NewJobClient()
	if err != nil {
		return nil, err
	}

	return &VolumeAdapter{
		volumeClient: pb.NewVolumeServiceClient(client.GetConn()),
	}, nil
}

// ListVolumes retrieves all available persistent volumes from the server
func (va *VolumeAdapter) ListVolumes(ctx context.Context) (*pb.Volumes, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return va.volumeClient.ListVolumes(ctx, &pb.EmptyRequest{})
}

// RuntimeAdapter handles runtime operations
type RuntimeAdapter struct {
	runtimeClient pb.RuntimeServiceClient
}

// NewRuntimeAdapter creates an adapter for runtime environment management
func NewRuntimeAdapter() (*RuntimeAdapter, error) {
	client, err := common.NewJobClient()
	if err != nil {
		return nil, err
	}

	return &RuntimeAdapter{
		runtimeClient: pb.NewRuntimeServiceClient(client.GetConn()),
	}, nil
}

// ListRuntimes retrieves all available runtime environments from the server
func (ra *RuntimeAdapter) ListRuntimes(ctx context.Context) (*pb.RuntimesRes, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return ra.runtimeClient.ListRuntimes(ctx, &pb.EmptyRequest{})
}
