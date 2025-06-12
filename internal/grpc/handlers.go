package grpc

import "context"

func (s *Server) Create(ctx context.Context, in *CreateRequest) (*CreateResponse, error) {
	var resp CreateResponse

	data, err := s.URLService.Save(ctx, in.Url, int(in.UserId))
	if err != nil {
		return nil, err
	}

	resp.ShortUrl = data.ShortURL

	return &resp, nil
}
