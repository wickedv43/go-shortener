package grpcserver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-shortener/internal/config"
	pb "github.com/wickedv43/go-shortener/internal/grpcapi"
	"github.com/wickedv43/go-shortener/internal/mocks"
	"github.com/wickedv43/go-shortener/internal/storage"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestServer_Create(t *testing.T) {
	// используем mocks.MockShortener
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockShortener := mocks.NewMockShortener(ctrl)
	mockShortener.EXPECT().Save(gomock.Any(), "http://example.com", 123).
		Return(storage.Data{ShortURL: "abc123"}, nil)

	s := &Server{
		URLService: mockShortener,
	}

	ctx := context.WithValue(context.Background(), userIDKey, 123)

	resp, err := s.Create(ctx, &pb.CreateRequest{
		Url: "http://example.com",
	})

	require.NoError(t, err)
	require.Equal(t, "abc123", resp.ShortUrl)
}

func TestServer_CreateBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockShortener := mocks.NewMockShortener(ctrl)
	mockShortener.EXPECT().Save(gomock.Any(), "http://example.com/1", 123).
		Return(storage.Data{ShortURL: "short1"}, nil)
	mockShortener.EXPECT().Save(gomock.Any(), "http://example.com/2", 123).
		Return(storage.Data{ShortURL: "short2"}, nil)

	s := &Server{
		URLService: mockShortener,
		cfg:        &config.Config{Server: config.Server{FlagSuffixAddr: "http://localhost:8080"}},
	}

	ctx := context.WithValue(context.Background(), userIDKey, 123)

	resp, err := s.CreateBatch(ctx, &pb.BatchRequest{
		Urls: []*pb.BatchRequest_Item{
			{CorrelationId: "1", OriginalUrl: "http://example.com/1"},
			{CorrelationId: "2", OriginalUrl: "http://example.com/2"},
		},
	})

	require.NoError(t, err)
	require.Len(t, resp.Urls, 2)
	require.Equal(t, "http://localhost:8080/short1", resp.Urls[0].ShortUrl)
	require.Equal(t, "http://localhost:8080/short2", resp.Urls[1].ShortUrl)
}

func TestServer_GetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockShortener := mocks.NewMockShortener(ctrl)
	mockShortener.EXPECT().GetAll(gomock.Any(), 123).Return([]storage.Data{
		{ShortURL: "short1", OriginalURL: "http://example.com/1"},
		{ShortURL: "short2", OriginalURL: "http://example.com/2"},
	}, nil)

	s := &Server{
		URLService: mockShortener,
	}

	ctx := context.WithValue(context.Background(), userIDKey, 123)

	resp, err := s.GetAll(ctx, &emptypb.Empty{})

	require.NoError(t, err)
	require.Len(t, resp.Urls, 2)
	require.Equal(t, "short1", resp.Urls[0].ShortUrl)
	require.Equal(t, "http://example.com/1", resp.Urls[0].OriginalUrl)
	require.Equal(t, "short2", resp.Urls[1].ShortUrl)
	require.Equal(t, "http://example.com/2", resp.Urls[1].OriginalUrl)
}

func TestServer_DeleteUserURLs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockShortener := mocks.NewMockShortener(ctrl)
	mockShortener.EXPECT().DeleteUserURLS(gomock.Any(), 123, []string{"short1", "short2"}).
		Return(nil)

	s := &Server{
		URLService: mockShortener,
	}

	ctx := context.WithValue(context.Background(), userIDKey, 123)

	_, err := s.DeleteUserURLs(ctx, &pb.DeleteUserURLsRequest{
		ShortUrls: []string{"short1", "short2"},
	})

	require.NoError(t, err)
}

func TestServer_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockShortener := mocks.NewMockShortener(ctrl)
	mockShortener.EXPECT().HealthCheck().Return(nil)

	s := &Server{
		URLService: mockShortener,
	}

	_, err := s.Ping(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
}

func TestServer_Stats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockShortener := mocks.NewMockShortener(ctrl)
	mockShortener.EXPECT().Stats().Return(42, 7, nil)

	s := &Server{
		URLService: mockShortener,
	}

	resp, err := s.Stats(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	require.Equal(t, int32(42), resp.Urls)
	require.Equal(t, int32(7), resp.Users)
}
