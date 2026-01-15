package handlers

import (
	pb "admin_service/internal/proto/admin"
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
)

type ApiKeyHandler struct {
	pb.UnimplementedApiKeyServiceServer
	usecase interface {
		Create(string, []string) (string, error)
		Revoke(string) error
	}
}

func NewApiKeyHandler(u interface {
	Create(string, []string) (string, error)
	Revoke(string) error
}) *ApiKeyHandler {
	return &ApiKeyHandler{usecase: u}
}

func (h *ApiKeyHandler) Create(ctx context.Context, req *pb.CreateApiKeyRequest) (*pb.CreateApiKeyResponse, error) {
	key, err := h.usecase.Create(req.AppId, req.Scopes)
	return &pb.CreateApiKeyResponse{ApiKey: key}, err
}

func (h *ApiKeyHandler) Revoke(ctx context.Context, req *pb.RevokeApiKeyRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, h.usecase.Revoke(req.KeyId)
}
