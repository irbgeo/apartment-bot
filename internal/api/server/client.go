package server

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

	"github.com/irbgeo/apartment-bot/internal/api/middleware"
	api "github.com/irbgeo/apartment-bot/internal/api/server/proto"
	"github.com/irbgeo/apartment-bot/internal/server"
)

type client struct {
	cli api.ServerClient

	apartmentCh chan server.Apartment
	errCh       chan error
}

func NewClient(
	addr, authToken string,
	id int64,
) (*client, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(middleware.AddMetadataUnaryInterceptor(authToken, id)),
		grpc.WithStreamInterceptor(middleware.AddMetadataStreamInterceptor(authToken, id)),
	)

	if err != nil {
		err = fmt.Errorf(status.Convert(err).Message())
		return nil, err
	}

	cli := &client{
		cli: api.NewServerClient(conn),

		apartmentCh: make(chan server.Apartment),
		errCh:       make(chan error),
	}

	return cli, nil
}

func (s *client) SaveFilter(ctx context.Context, filter server.Filter) (int64, error) {
	res, err := s.cli.SaveFilter(ctx, filterToAPI(filter))
	if err != nil {
		err = fmt.Errorf(status.Convert(err).Message())
		return 0, err
	}

	return res.Count, nil
}

func (s *client) Filters(ctx context.Context, u server.User) ([]server.Filter, error) {
	req := &api.FilterListReq{
		UserId: u.ID,
	}

	resp, err := s.cli.Filters(ctx, req)
	if err != nil {
		err = fmt.Errorf(status.Convert(err).Message())
		return nil, err
	}

	filters := make([]server.Filter, 0, len(resp.Filters))

	for _, f := range resp.Filters {
		filters = append(filters, filterFromAPI(f))
	}

	return filters, nil
}

func (s *client) Filter(ctx context.Context, f server.Filter) (*server.Filter, error) {
	resp, err := s.cli.FilterInfo(ctx, filterToAPI(f))
	if err != nil {
		err = fmt.Errorf(status.Convert(err).Message())
		return nil, err
	}

	filter := filterFromAPI(resp)
	return &filter, nil
}

func (s *client) DeleteFilter(ctx context.Context, f server.Filter) error {
	_, err := s.cli.DeleteFilter(ctx, filterToAPI(f))
	if err != nil {
		err = fmt.Errorf(status.Convert(err).Message())
		return err
	}

	return nil
}

func (s *client) ConnectUser(ctx context.Context, u server.User) error {
	_, err := s.cli.ConnectUser(ctx, &api.User{Id: u.ID})
	if err != nil {
		err = fmt.Errorf(status.Convert(err).Message())
		return err
	}

	return nil
}

func (s *client) DisconnectUser(ctx context.Context, u server.User) error {
	_, err := s.cli.DisconnectUser(ctx, &api.User{Id: u.ID})
	if err != nil {
		err = fmt.Errorf(status.Convert(err).Message())
		return err
	}

	return nil
}

func (s *client) Cities(ctx context.Context) (map[string][]string, error) {
	cities, err := s.cli.Cities(ctx, &emptypb.Empty{})
	if err != nil {
		err = fmt.Errorf(status.Convert(err).Message())
		return nil, err
	}

	result := make(map[string][]string)
	for name, district := range cities.Name {
		result[name] = district.Names
	}

	return result, nil
}

func (s *client) Apartments(ctx context.Context, f server.Filter) (<-chan server.Apartment, <-chan error, error) {
	stream, err := s.cli.Apartments(ctx, filterToAPI(f))
	if err != nil {
		err = fmt.Errorf(status.Convert(err).Message())
		return nil, nil, err
	}

	apartmentCh := make(chan server.Apartment)
	errCh := make(chan error)

	go s.apartmentPipeline(ctx, stream.CloseSend, apartmentCh, errCh, stream.Recv)

	return apartmentCh, errCh, nil
}

func (s *client) StartApartmentWatcher(ctx context.Context) (<-chan server.Apartment, <-chan error, error) {
	stream, err := s.cli.Connect(ctx, &emptypb.Empty{})
	if err != nil {
		err = fmt.Errorf(status.Convert(err).Message())
		return nil, nil, err
	}

	go s.apartmentPipeline(ctx, stream.CloseSend, s.apartmentCh, s.errCh, stream.Recv)

	return s.apartmentCh, s.errCh, nil
}

func (s *client) apartmentPipeline(ctx context.Context, closePipeline func() error, apartmentCh chan server.Apartment, errCh chan error, receiveApartment func() (*api.Apartment, error)) {
	for {
		select {
		case <-ctx.Done():
			err := closePipeline()
			if err != nil {
				slog.Error("pipeline_close", "err", err)
			}
			return
		default:
			resp, err := receiveApartment()
			if err != nil {
				err = fmt.Errorf(status.Convert(err).Message())

				if strings.Contains(err.Error(), io.EOF.Error()) {
					return
				}

				errCh <- err
			}

			if resp != nil {
				apartmentCh <- apartmentFromAPI(resp)
			}
		}
	}
}
