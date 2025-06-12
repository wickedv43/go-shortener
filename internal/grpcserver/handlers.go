package grpcserver

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	pb "github.com/wickedv43/go-shortener/internal/grpcapi"
	"github.com/wickedv43/go-shortener/internal/storage"
	"github.com/wickedv43/go-shortener/internal/url"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Create creates a short URL from the given original URL.
func (s *Server) Create(ctx context.Context, in *pb.CreateRequest) (*pb.CreateResponse, error) {
	var resp pb.CreateResponse

	ctxUserID := ctx.Value(userIDKey)

	userID, ok := ctxUserID.(int)
	if !ok {
		return &resp, status.Error(codes.Unauthenticated, "invalid userID")
	}

	data, err := s.URLService.Save(ctx, in.Url, userID)
	if err != nil {
		if errors.Is(err, url.ErrEmptyURL) {
			return nil, status.Errorf(codes.Aborted, url.ErrEmptyURL.Error())
		}
		if errors.Is(err, url.ErrConflict) {
			resp.ShortUrl = data.ShortURL

			return &resp, nil
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	resp.ShortUrl = data.ShortURL

	return &resp, nil
}

// CreateBatch creates multiple short URLs in a batch operation.
func (s *Server) CreateBatch(ctx context.Context, in *pb.BatchRequest) (*pb.BatchResponse, error) {
	var resp pb.BatchResponse

	var (
		err  error
		data storage.Data
	)

	ctxUserID := ctx.Value(userIDKey)

	userID, ok := ctxUserID.(int)
	if !ok {
		return &resp, status.Error(codes.Unauthenticated, "invalid userID")
	}

	for _, req := range in.Urls {
		data, err = s.URLService.Save(ctx, req.OriginalUrl, userID)
		if err != nil {
			return nil, status.Error(codes.Internal, url.ErrInternal.Error())
		}

		res := fmt.Sprintf("%s/%s", s.cfg.Server.FlagSuffixAddr, data.ShortURL)
		r := &pb.BatchResponse_Item{
			CorrelationId: req.CorrelationId,
			ShortUrl:      res,
		}
		resp.Urls = append(resp.Urls, r)
	}

	return &resp, nil
}

// GetShort retrieves the original URL corresponding to the given short URL.
func (s *Server) GetShort(ctx context.Context, in *pb.GetShortRequest) (*pb.GetShortResponse, error) {
	var resp pb.GetShortResponse

	data, err := s.URLService.Get(ctx, in.ShortUrl)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if data.DeletedFlag {
		return nil, status.Error(codes.Canceled, url.ErrGone.Error())
	}

	return &resp, nil
}

// GetAll retrieves all URLs created by the given user.
func (s *Server) GetAll(ctx context.Context, _ *emptypb.Empty) (*pb.GetAllResponse, error) {
	var resp pb.GetAllResponse

	ctxUserID := ctx.Value(userIDKey)

	userID, ok := ctxUserID.(int)
	if !ok {
		return &resp, status.Error(codes.Unauthenticated, "invalid userID")
	}

	urls, err := s.URLService.GetAll(ctx, userID)
	if err != nil {
		if errors.Is(err, url.ErrNoContent) {
			return nil, status.Errorf(codes.NotFound, url.ErrNoContent.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	for _, url := range urls {
		d := &pb.URLData{
			ShortUrl:    url.ShortURL,
			OriginalUrl: url.OriginalURL,
		}

		resp.Urls = append(resp.Urls, d)
	}

	return &resp, nil
}

// DeleteUserURLs deletes a list of short URLs for the given user.
func (s *Server) DeleteUserURLs(ctx context.Context, in *pb.DeleteUserURLsRequest) (*emptypb.Empty, error) {
	ctxUserID := ctx.Value(userIDKey)

	userID, ok := ctxUserID.(int)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid userID")
	}

	err := s.URLService.DeleteUserURLS(ctx, userID, in.ShortUrls)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return nil, nil
}

// Stats returns the total number of URLs and users in the system.
func (s *Server) Stats(_ context.Context, _ *emptypb.Empty) (*pb.StatsResponse, error) {
	urls, users, err := s.URLService.Stats()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.StatsResponse{
		Urls:  int32(urls),
		Users: int32(users),
	}, nil
}

// Ping checks if the service is healthy and the storage is reachable.
func (s *Server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	err := s.URLService.Ping()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
