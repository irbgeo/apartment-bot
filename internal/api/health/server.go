package health

import (
	"context"
	"net"

	"google.golang.org/grpc"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type server struct {
}

func ListenAndServe(
	addr string,
) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	srv := &server{}

	s := grpc.NewServer()

	health.RegisterHealthServer(s, srv)

	reflection.Register(s)

	return s.Serve(l)
}

func (s *server) Check(ctx context.Context, in *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	return &health.HealthCheckResponse{Status: health.HealthCheckResponse_SERVING}, nil
}

func (s *server) Watch(in *health.HealthCheckRequest, _ health.Health_WatchServer) error {
	return nil
}
