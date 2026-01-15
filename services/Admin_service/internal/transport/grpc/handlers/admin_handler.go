package handlers

import (
	pb "admin_service/internal/proto/admin"
	"context"
)

type AdminHandler struct {
	pb.UnimplementedAdminServiceServer
	usecase interface {
		Login(string,string)(string,error)
	}
}

func NewAdminHandler(u interface{
	Login(string,string)(string,error)
})*AdminHandler {
	return &AdminHandler{usecase: u}
}

func (h *AdminHandler) Login(ctx context.Context,req *pb.LoginRequest)(*pb.LoginResponse,error) {
	token,err := h.usecase.Login(req.Email,req.Password)
	if err != nil {
		return nil,err 
	}
	return &pb.LoginResponse{AccessToken: token},nil 
}