package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"

	pb "github.com/red-hat-storage/ocs-operator/services/provider/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
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
	token, err := c.authManager.GetAccessToken()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate access token: %v", err)
	}
	return &pb.GenereateTokenResponse{AccessToken: token}, nil
}

// OnBoardConsumer RPC call to onboard a new OCS consumer cluster.
func (c *ocsProviderServer) OnBoardConsumer(ctx context.Context, req *pb.OnBoardConsumerRequest) (*pb.OnBoardConsumerResponse, error) {
	/* TODO:
	- Create Storage Consumer CR
	- Return encrypted ID of the consumer CR
	*/
	uid, err := c.authManager.GetEncryptedID("testUUID")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to encrypt consumer resource ID: %v", err)
	}

	base64.StdEncoding.EncodeToString(uid)
	return &pb.OnBoardConsumerResponse{StorageConsumerUUID: base64.StdEncoding.EncodeToString(uid), GrantedCapacity: "2Gb"}, nil
}

// GetStorageConfig RPC call to onboard a new OCS consumer cluster.
func (c *ocsProviderServer) GetStorageConfig(ctx context.Context, req *pb.StorageConfigRequest) (*pb.StorageConfigResponse, error) {
	/* TODO:
	- Verify Status of the StorageConsumer CR.
	- Return connection string
	*/
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

func Start(port int, providerServer *ocsProviderServer, opts []grpc.ServerOption) {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		klog.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterOCSProviderServer(grpcServer, providerServer)
	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)
	err = grpcServer.Serve(lis)
	if err != nil {
		klog.Fatalf("failed to start gRPC server: %v", err)
	}
}
