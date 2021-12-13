package client

import (
	"context"

	"github.com/red-hat-storage/ocs-operator/services/provider/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"k8s.io/klog"
)

type AuthInterceptor struct {
	accessToken string
	consumerID  string
}

func NewAuthInterceptor(accessToken, consumerID string) *AuthInterceptor {
	return &AuthInterceptor{accessToken, consumerID}
}

func (interceptor *AuthInterceptor) Unary() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		klog.Infof("--> unary interceptor: %q", method)

		if common.TokenAuthMethods()[method] {
			return invoker(interceptor.attachToken(ctx), method, req, reply, cc, opts...)
		} else if common.UIDAuthMethods()[method] {
			return invoker(interceptor.attachUID(ctx), method, req, reply, cc, opts...)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (interceptor *AuthInterceptor) attachToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", interceptor.accessToken)
}

func (interceptor *AuthInterceptor) attachUID(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", interceptor.consumerID)
}
