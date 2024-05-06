package middleware

import (
	"context"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/irbgeo/apartment-bot/internal/utils"
)

const (
	authKey = "auth_key"
	idKey   = "id_key"
)

func AddMetadataUnaryInterceptor(tokenAuth string, id int64) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		md := metadata.New(map[string]string{
			authKey: tokenAuth,
			idKey:   strconv.FormatInt(id, 10),
		})
		ctx = metadata.NewOutgoingContext(ctx, md)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func AddMetadataStreamInterceptor(tokenAuth string, id int64) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		md := metadata.New(map[string]string{
			authKey: tokenAuth,
			idKey:   strconv.FormatInt(id, 10),
		})
		ctx = metadata.NewOutgoingContext(ctx, md)
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func CheckMetadataUnaryInterceptor(tokenAuth string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if strings.Contains(info.FullMethod, "Health") {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}

		token, ok := md[authKey]
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing token")
		}

		if token[0] != tokenAuth {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = utils.PackVar(ctx, utils.IDKey, GetID(ctx))

		return handler(ctx, req)
	}
}

func CheckMetadataStreamInterceptor(tokenAuth string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if strings.Contains(info.FullMethod, "Health") {
			return handler(srv, ss)
		}

		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Error(codes.InvalidArgument, "missing metadata")
		}

		token, ok := md[authKey]
		if !ok {
			return status.Error(codes.Unauthenticated, "missing token")
		}

		if token[0] != tokenAuth {
			return status.Error(codes.Unauthenticated, "invalid token")
		}

		return handler(srv, ss)
	}
}

func GetID(ctx context.Context) int64 {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		ids, ok := md[idKey]
		if ok {
			id, _ := strconv.ParseInt(ids[0], 10, 64)
			return id
		}
	}
	return 0
}
