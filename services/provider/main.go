package main

import (
	"flag"
	"time"

	"github.com/red-hat-storage/ocs-operator/services/provider/server"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

func main() {
	flag.Parse()

	authManager := server.NewAuthManager("mysecretkey", 4*time.Minute)
	interceptor := server.NewAuthInterceptor(authManager)
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.Unary()),
	}
	server.Start(*port, opts)
}
