package server

import (
	"context"
	"fmt"
	"net"

	pb "github.com/red-hat-storage/ocs-operator/services/provider/pb"
	"google.golang.org/grpc"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ocsProviderServer struct {
	pb.UnimplementedOCSProviderServer
	client.Client
	authManager *AuthManager
}

func NewOCSProviderServer(client client.Client, auth *AuthManager) *ocsProviderServer {
	return &ocsProviderServer{
		Client:      client,
		authManager: auth}
}

// GenerateToken RPC call to generate a new jwt token for the consumer cluster
func (c *ocsProviderServer) GenerateToken(ctx context.Context, req *pb.GenerateTokenRequest) (*pb.GenereateTokenResponse, error) {
	return &pb.GenereateTokenResponse{}, nil
}

// OnBoardConsumer RPC call to onboard a new OCS consumer cluster.
func (c *ocsProviderServer) OnBoardConsumer(ctx context.Context, req *pb.OnBoardConsumerRequest) (*pb.OnBoardConsumerResponse, error) {
	return &pb.OnBoardConsumerResponse{}, nil
}

// GetStorageConfig RPC call to onboard a new OCS consumer cluster.
func (c *ocsProviderServer) GetStorageConfig(ctx context.Context, req *pb.StorageConfigRequest) (*pb.StorageConfigResponse, error) {
	return &pb.StorageConfigResponse{}, nil
}

// OffBoardConsumer RPC call to delete the StorageConsumer CR
func (c *ocsProviderServer) OffBoardConsumer(ctx context.Context, req *pb.OffBoardConsumerRequest) (*pb.OffBoardConsumerResponse, error) {
	return &pb.OffBoardConsumerResponse{}, nil
}

// UpdateCapacity PRC call to increase or decrease the storage pool size
func (c *ocsProviderServer) UpdateCapacity(ctx context.Context, req *pb.UpdateCapacityRequest) (*pb.UpdateCapacityResponse, error) {
	return &pb.UpdateCapacityResponse{}, nil
}

func Start(port int, opts []grpc.ServerOption) {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		klog.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterOCSProviderServer(grpcServer, &ocsProviderServer{})
	err = grpcServer.Serve(lis)
	if err != nil {
		klog.Fatalf("failed to start gRPC server: %v", err)
	}
}
