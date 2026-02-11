package handlers

import (
	"admin_service/internal/domain/entities"
	pb "admin_service/internal/proto/admin"
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
)

type AppUserHandler struct {
	pb.UnimplementedAppUserServiceServer
	usecase interface {
		Create(string, string, string) error
		List(string) ([]*entities.AppUser, error)
		AppLogin(email,password string)(string,error)
	}
}

func NewAppUserHandler(u interface {
	Create(string, string, string) error
	List(string) ([]*entities.AppUser, error)
	AppLogin(email,password string)(string,error)
}) *AppUserHandler {
	return &AppUserHandler{usecase: u}
}

func (h *AppUserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, h.usecase.Create(
		req.Email, req.Password, req.Role,
	)
}

func (h *AppUserHandler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	users, err := h.usecase.List(req.AppId)
	if err != nil {
		return nil, err
	}

	var res []*pb.AppUser
	for _, u := range users {
		res = append(res, &pb.AppUser{
			Id:     u.ID,
			Email:  u.Email,
			Role:   u.Role,
			Status: u.Status,
		})
	}

	return &pb.ListUsersResponse{Users: res}, nil
}

func (h *AppUserHandler) AppLogin(ctx context.Context, req *pb.AppLoginRequest) (*pb.AppLoginResponse,error) {
	token,err := h.usecase.AppLogin(req.Email,req.Password)
	if err != nil {
		return nil,err 
	}
	return &pb.AppLoginResponse{AccessToken: token},nil 
}