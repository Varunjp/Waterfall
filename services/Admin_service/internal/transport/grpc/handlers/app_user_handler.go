package handlers

import (
	"admin_service/internal/domain/entities"
	"admin_service/internal/domain/errors"
	"admin_service/internal/pkg/utils"
	pb "admin_service/internal/proto/admin"
	"admin_service/internal/usecase/interfaces"
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
)

type AppUserHandler struct {
	pb.UnimplementedAppUserServiceServer
	usecase interfaces.AppUserUsecase
}

func NewAppUserHandler(u interfaces.AppUserUsecase) *AppUserHandler {
	return &AppUserHandler{usecase: u}
}

func (h *AppUserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, h.usecase.Create(
		req.AppId,req.Email, req.Password, req.Role,
	)
}

func (h *AppUserHandler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {

	appID,err := utils.GetAppIDFromContext(ctx)
	
	if err != nil {
		return nil, errors.ErrUnauthenticated
	}

	users, err := h.usecase.List(appID)
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

func (h *AppUserHandler) RequestPasswordReset(ctx context.Context, req *pb.RequestResetPasswordRequest)(*pb.RequestResetPasswordResponse,error) {
	err := h.usecase.RequestPasswordReset(ctx,req.Email)
	if err != nil {
		return nil,err 
	}

	return &pb.RequestResetPasswordResponse{
		Message: "OTP sent to email",
	},nil 
}

func (h *AppUserHandler) VerifyPasswordResetOtp(ctx context.Context,req *pb.VerifyPasswordResetOtpRequest)(*pb.VerifyPasswordResetOtpResponse,error) {

	token,err := h.usecase.VerifyOtp(ctx,req.Email,req.Otp)
	if err != nil {
		return nil ,err 
	}

	return &pb.VerifyPasswordResetOtpResponse{
		ResetToken: token,
	},nil 
}

func (h *AppUserHandler) ResetPassword(ctx context.Context,req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse,error) {
	err := h.usecase.ResetPassword(ctx,req.ResetToken,req.NewPassword)

	if err != nil {
		return  nil,err 
	}

	return &pb.ResetPasswordResponse{
		Message: "password updated",
	},nil 
}

func (h *AppUserHandler) ListPlans(ctx context.Context, req *pb.ListPlansRequest)(*pb.ListPlansResponse,error) {

	plan,err := h.usecase.ListPlans(ctx)

	if err != nil {
		return nil,err 
	}

	return mapUPlans(plan),nil
}

func mapUPlans(plans []*entities.Plan) *pb.ListPlansResponse {
	resp := &pb.ListPlansResponse{}
	for _,p := range plans {
		resp.Plans = append(resp.Plans, &pb.UserPlan{
			PlanId: p.PlanID,
			PlanName: p.Name,
			MonthlyLimit: int32(p.MonthlyJobLimit),
			Planprice: p.Price,
		})
	}
	return resp 
}