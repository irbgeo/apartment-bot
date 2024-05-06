package server

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/irbgeo/apartment-bot/internal/api/middleware"
	api "github.com/irbgeo/apartment-bot/internal/api/server/proto"
	"github.com/irbgeo/apartment-bot/internal/server"
	"github.com/irbgeo/apartment-bot/internal/utils"
)

type srv struct {
	svc serverSvc
}

type serverSvc interface {
	SaveFilter(ctx context.Context, f server.Filter) (int64, error)
	Filter(ctx context.Context, f server.Filter) (*server.Filter, error)
	Filters(ctx context.Context, u server.User) ([]server.Filter, error)
	DeleteFilter(ctx context.Context, f server.Filter) error
	ConnectUser(ctx context.Context, u server.User) error
	DisconnectUser(ctx context.Context, u server.User) error
	Cities(ctx context.Context) ([]server.City, error)
	Apartments(ctx context.Context, f server.Filter) (<-chan server.Apartment, error)
	Subscribe(ctx context.Context) <-chan server.Apartment
	Unsubscribe(ctx context.Context)
}

func ListenAndServe(
	addr string,
	authToken string,
	svc serverSvc,
) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	srv := &srv{
		svc: svc,
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.CheckMetadataUnaryInterceptor(authToken)),
		grpc.StreamInterceptor(middleware.CheckMetadataStreamInterceptor(authToken)),
	)

	api.RegisterServerServer(s, srv)

	reflection.Register(s)

	return s.Serve(l)
}

func (s *srv) SaveFilter(ctx context.Context, req *api.Filter) (*api.SaveFilterResult, error) {
	res := &api.SaveFilterResult{}

	count, err := s.svc.SaveFilter(ctx, filterFromAPI(req))
	if err != nil {
		return res, err
	}

	res.Count = count

	return res, nil
}

func (s *srv) FilterInfo(ctx context.Context, in *api.Filter) (*api.Filter, error) {
	f, err := s.svc.Filter(ctx, filterFromAPI(in))
	if err != nil {
		return nil, err
	}

	return filterToAPI(*f), nil
}

func (s *srv) Filters(ctx context.Context, in *api.FilterListReq) (*api.FilterListRes, error) {
	filters, err := s.svc.Filters(ctx, server.User{ID: in.UserId})
	if err != nil {
		return nil, err
	}

	res := &api.FilterListRes{
		Filters: make([]*api.Filter, 0, len(filters)),
	}

	for _, f := range filters {
		res.Filters = append(res.Filters, filterToAPI(f))
	}
	return res, nil
}

func (s *srv) DeleteFilter(ctx context.Context, in *api.Filter) (*emptypb.Empty, error) {
	err := s.svc.DeleteFilter(ctx, filterFromAPI(in))
	return &emptypb.Empty{}, err
}

func (s *srv) ConnectUser(ctx context.Context, in *api.User) (*emptypb.Empty, error) {
	err := s.svc.ConnectUser(ctx, server.User{ID: in.Id})
	return &emptypb.Empty{}, err
}

func (s *srv) DisconnectUser(ctx context.Context, in *api.User) (*emptypb.Empty, error) {
	err := s.svc.DisconnectUser(ctx, server.User{ID: in.Id})
	return &emptypb.Empty{}, err
}

func (s *srv) Cities(ctx context.Context, _ *emptypb.Empty) (*api.City, error) {
	cities, err := s.svc.Cities(ctx)
	if err != nil {
		return &api.City{}, err
	}

	r := &api.City{
		Name: make(map[string]*api.District),
	}

	for _, city := range cities {
		districtNames := make([]string, 0, len(city.District))
		for name := range city.District {
			districtNames = append(districtNames, name)
		}

		districts, ok := r.Name[city.Name]
		if ok {
			districts.Names = append(districts.Names, districtNames...)
			continue
		}

		r.Name[city.Name] = &api.District{
			Names: districtNames,
		}
	}

	return r, nil
}

func (s *srv) Apartments(req *api.Filter, srv api.Server_ApartmentsServer) error {
	apartmentCh, err := s.svc.Apartments(srv.Context(), filterFromAPI(req))
	if err != nil {
		return err
	}

	for {
		select {
		case <-srv.Context().Done():
			return nil
		case a, ok := <-apartmentCh:
			if !ok {
				return nil
			}

			if err := srv.Send(apartmentToAPI(a)); err != nil {
				return err
			}
		}
	}
}

func (s *srv) Connect(req *emptypb.Empty, srv api.Server_ConnectServer) error {
	ctx := utils.PackVar(srv.Context(), utils.IDKey, middleware.GetID(srv.Context()))

	apartmentCh := s.svc.Subscribe(ctx)
	defer s.svc.Unsubscribe(ctx)
	for {
		select {
		case <-srv.Context().Done():
			return nil
		case a, ok := <-apartmentCh:
			if !ok {
				return nil
			}

			if err := srv.Send(apartmentToAPI(a)); err != nil {
				return err
			}
		}
	}
}
