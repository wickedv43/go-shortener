package grpcserver

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-shortener/internal/auth"
	"github.com/wickedv43/go-shortener/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestLogUnaryInterceptor(t *testing.T) {
	s := &Server{
		log: logrus.NewEntry(logrus.New()), // пустой логгер, можно оборачивать если хочешь
	}

	interceptor := s.LogUnaryInterceptor()

	ctx := context.Background()

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.TestService/TestMethod",
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	resp, err := interceptor(ctx, "request", info, handler)

	require.NoError(t, err)
	require.Equal(t, "ok", resp)
}

func TestTrustedSubnetInterceptor(t *testing.T) {
	s := &Server{
		cfg: &config.Config{
			Server: config.Server{
				FlagTrustedSubnet: "192.168.1.0/24",
			},
		},
		log: logrus.NewEntry(logrus.New()),
	}

	interceptor := s.TrustedSubnetInterceptor()

	info := &grpc.UnaryServerInfo{
		FullMethod: "/shortener.Shortener/Stats",
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "allowed", nil
	}

	// allowed IP
	md := metadata.New(map[string]string{"x-real-ip": "192.168.1.42"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	resp, err := interceptor(ctx, "req", info, handler)
	require.NoError(t, err)
	require.Equal(t, "allowed", resp)

	// forbidden IP
	md = metadata.New(map[string]string{"x-real-ip": "10.0.0.1"})
	ctx = metadata.NewIncomingContext(context.Background(), md)

	resp, err = interceptor(ctx, "req", info, handler)
	require.Nil(t, resp)
	require.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestAuthInterceptor(t *testing.T) {
	s := &Server{
		log: logrus.NewEntry(logrus.New()),
	}

	interceptor := s.AuthInterceptor()

	info := &grpc.UnaryServerInfo{
		FullMethod: "/shortener.Shortener/Create",
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		userID, ok := ctx.Value(userIDKey).(int)
		require.True(t, ok)
		require.NotZero(t, userID)

		return "ok", nil
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(nil))

	resp, err := interceptor(ctx, "req", info, handler)
	require.NoError(t, err)
	require.Equal(t, "ok", resp)

	// valid token
	token, err := auth.CreateJWT()
	require.NoError(t, err)

	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx = metadata.NewIncomingContext(context.Background(), md)

	resp, err = interceptor(ctx, "req", info, handler)
	require.NoError(t, err)
	require.Equal(t, "ok", resp)

	// non-valid token
	md = metadata.New(map[string]string{"authorization": "Bearer invalid.token.here"})
	ctx = metadata.NewIncomingContext(context.Background(), md)

	resp, err = interceptor(ctx, "req", info, handler)
	require.Nil(t, resp)
	require.Equal(t, codes.Unauthenticated, status.Code(err))
}
