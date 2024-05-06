package message

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"

	api "github.com/irbgeo/apartment-bot/internal/api/message/proto"
	"github.com/irbgeo/apartment-bot/internal/api/middleware"
	"github.com/irbgeo/apartment-bot/internal/message"
)

type server struct {
	svc messageService
}

type messageService interface {
	Publish(ctx context.Context, msg message.Message) error
	Watcher() <-chan message.Message
}

func ListenAndServe(
	addr string,
	authToken string,
	svc messageService,
) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	srv := &server{
		svc: svc,
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.CheckMetadataUnaryInterceptor(authToken)),
		grpc.StreamInterceptor(middleware.CheckMetadataStreamInterceptor(authToken)),
	)

	api.RegisterMessageServerServer(s, srv)

	reflection.Register(s)

	return s.Serve(l)
}

func (s *server) Publish(ctx context.Context, r *api.Message) (*emptypb.Empty, error) {
	err := s.svc.Publish(ctx, messageFromBotToSvc(r))
	return &emptypb.Empty{}, err
}

func (s *server) Watch(req *emptypb.Empty, srv api.MessageServer_WatchServer) error {
	for {
		select {
		case <-srv.Context().Done():
			return nil
		case msg, ok := <-s.svc.Watcher():
			if !ok {
				return nil
			}
			if err := srv.Send(messageFromSvcToBot(msg)); err != nil {
				return err
			}
		}
	}
}
