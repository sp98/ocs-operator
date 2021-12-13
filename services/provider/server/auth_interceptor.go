package server

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor is a server interceptor for authenticating access token and StorageConsumer UID
type AuthInterceptor struct {
	authManager *AuthManager
}

// NewAuthInterceptor returns a new auth interceptor
func NewAuthInterceptor(authManager *AuthManager) *AuthInterceptor {
	return &AuthInterceptor{authManager}
}

// Unary returns a server interceptor function to authenticate and authorize unary RPC
func (interceptor *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		log.Println("--> unary interceptor: ", info.FullMethod)

		err := interceptor.authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func (a *AuthInterceptor) authorize(ctx context.Context, method string) error {

	switch method {
	case "providerpb.OnBoardConsumer":
		return a.authorizeToken(ctx)
	default:
		return a.authorizeConsumerID(ctx)
	}

}

// authorizeConsumerID validates the storageConsumer UID
func (a *AuthInterceptor) authorizeConsumerID(ctx context.Context) error {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	accessToken := values[0]
	_, err := a.authManager.VerifyAccessToken(accessToken)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}

	return nil
}

// authorizeToken validates the access token
func (a *AuthInterceptor) authorizeToken(ctx context.Context) error {
	return nil
}
