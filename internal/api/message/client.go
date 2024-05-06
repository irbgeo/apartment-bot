package message

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	api "github.com/irbgeo/apartment-bot/internal/api/message/proto"
	"github.com/irbgeo/apartment-bot/internal/api/middleware"
	tgbot "github.com/irbgeo/apartment-bot/internal/client/tg"
)

type client struct {
	cli api.MessageServerClient

	messageCh chan tgbot.Message
	errCh     chan error
}

func NewClient(
	addr string,
	authToken string,
) (*client, error) {
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(middleware.AddMetadataUnaryInterceptor(authToken, 0)),
		grpc.WithStreamInterceptor(middleware.AddMetadataStreamInterceptor(authToken, 0)),
	)
	if err != nil {
		err = fmt.Errorf(status.Convert(err).Message())
		return nil, err
	}

	cli := &client{
		cli: api.NewMessageServerClient(conn),

		messageCh: make(chan tgbot.Message),
		errCh:     make(chan error),
	}

	return cli, nil
}

func (s *client) StartWatcher(ctx context.Context) (<-chan tgbot.Message, <-chan error, error) {
	stream, err := s.cli.Watch(ctx, &emptypb.Empty{})
	if err != nil {
		err = fmt.Errorf(status.Convert(err).Message())
		return nil, nil, err
	}

	go s.messagePipeline(stream.Context(), s.messageCh, s.errCh, stream.Recv)

	return s.messageCh, s.errCh, nil
}

func (s *client) messagePipeline(ctx context.Context, messageCh chan tgbot.Message, errCh chan error, receiveMessage func() (*api.Message, error)) {
	for {
		select {
		case <-ctx.Done():
			if ctx.Err() != nil && ctx.Err() != context.Canceled {
				slog.Error("pipeline finish", "err", ctx.Err())
			}
			return
		default:
			resp, err := receiveMessage()
			if err != nil {
				err = fmt.Errorf(status.Convert(err).Message())

				if strings.Contains(err.Error(), io.EOF.Error()) {
					return
				}

				errCh <- err
			}

			messageCh <- messageFromAPIToBot(resp)
		}
	}
}
