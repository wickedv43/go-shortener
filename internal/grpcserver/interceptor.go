package grpcserver

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-shortener/internal/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const bearerPrefix = "Bearer "

// ChainUnaryInterceptors chains multiple grpc.UnaryServerInterceptor functions into a single interceptor.
//
// The interceptors are executed in the order they are provided (first to last).
// Useful for composing logging, authentication, and other middleware layers.
func ChainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		chain := handler

		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]

			next := chain

			chain = func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return interceptor(currentCtx, currentReq, info, next)
			}
		}

		return chain(ctx, req)
	}
}

// LogUnaryInterceptor returns a grpc.UnaryServerInterceptor that logs gRPC requests.
//
// The log includes method name, latency, and any returned error.
func (s *Server) LogUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()

		resp, err = handler(ctx, req)

		latency := time.Since(start)

		s.log.WithFields(logrus.Fields{
			"method":  info.FullMethod,
			"latency": latency,
			"error":   err,
		}).Infoln("grpc request")

		return resp, err
	}
}

// TrustedSubnetInterceptor returns a grpc.UnaryServerInterceptor that restricts access to the Stats method
// based on a configured trusted subnet.
//
// If no trusted subnet is configured or the client IP is not within the subnet, access is denied.
// This is used for protecting internal endpoints like /Stats.
func (s *Server) TrustedSubnetInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		// Apply only to Stats method
		if strings.HasSuffix(info.FullMethod, "/Stats") {
			if s.cfg.Server.FlagTrustedSubnet == "" {
				return nil, status.Error(codes.PermissionDenied, "trusted subnet not configured")
			}

			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, status.Error(codes.PermissionDenied, "missing metadata")
			}

			ips := md.Get("x-real-ip")
			if len(ips) == 0 {
				return nil, status.Error(codes.PermissionDenied, "missing x-real-ip")
			}

			realIP := ips[0]
			_, subnet, err := net.ParseCIDR(s.cfg.Server.FlagTrustedSubnet)
			if err != nil {
				s.log.Error("invalid trusted subnet", err)
				return nil, status.Error(codes.Internal, "invalid server config")
			}

			ip := net.ParseIP(realIP)
			if ip == nil || !subnet.Contains(ip) {
				return nil, status.Error(codes.PermissionDenied, "access forbidden")
			}
		}

		return handler(ctx, req)
	}
}

// AuthInterceptor returns a grpc.UnaryServerInterceptor that handles JWT-based authentication.
//
// If an Authorization header with a valid Bearer token is present, it parses and validates the token.
// If no Authorization header is present, it generates a new token for the client.
// The user ID is extracted from the token and injected into the request context as "userID".
func (s *Server) AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		authHeaders := md.Get("authorization")

		var userID int
		var jwtToken string
		var err error

		if len(authHeaders) == 0 {
			jwtToken, err = auth.CreateJWT()
			if err != nil {
				return nil, status.Error(codes.Internal, "could not generate token")
			}

			userID, err = auth.ParseJWT(jwtToken)
			if err != nil {
				return nil, status.Error(codes.Internal, "could not parse generated token")
			}
		} else {
			token := authHeaders[0]

			if !strings.HasPrefix(token, bearerPrefix) {
				return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
			}

			jwtToken = strings.TrimPrefix(token, bearerPrefix)

			userID, err = auth.ParseJWT(jwtToken)
			if err != nil {
				return nil, status.Error(codes.Unauthenticated, "invalid token")
			}
		}

		ctx = context.WithValue(ctx, "userID", userID)

		return handler(ctx, req)
	}
}
