package handlers

import (
	pb "admin_service/internal/proto/admin"
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
)

type EmailHandler struct {
	pb.UnimplementedEmailServiceServer
	usecase interface {
		Configure(string, string, string) error
	}
}

func NewEmailHandler(u interface {
	Configure(string, string, string) error
}) *EmailHandler {
	return &EmailHandler{usecase: u}
}

func (h *EmailHandler) Configure(ctx context.Context, req *pb.ConfigureEmailRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, h.usecase.Configure(
		req.AppId,
		req.Provider,
		req.FromEmail,
	)
}
