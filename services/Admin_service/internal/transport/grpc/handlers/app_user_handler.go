package handlers

import (
	"admin_service/internal/domain/entities"
	"admin_service/internal/domain/errors"
	"admin_service/internal/pkg/utils"
	pb "admin_service/internal/proto/admin"
	"admin_service/internal/usecase/interfaces"
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AppUserHandler struct {
	pb.UnimplementedAppUserServiceServer
	usecase interfaces.AppUserUsecase
}

func NewAppUserHandler(u interfaces.AppUserUsecase) *AppUserHandler {
	return &AppUserHandler{usecase: u}
}

func (h *AppUserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*emptypb.Empty, error) {
	appID, err := utils.GetAppIDFromContext(ctx)
	if err != nil {
		return nil, errors.ErrUnauthenticated
	}

	return &emptypb.Empty{}, h.usecase.Create(
		appID, req.Email, req.Password, req.Role,
	)
}

func (h *AppUserHandler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {

	appID, err := utils.GetAppIDFromContext(ctx)

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

func (h *AppUserHandler) ListAppUsers(ctx context.Context, req *pb.ListAppUserRequest) (*pb.ListAppUserResponse, error) {

	appID := req.AppId

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

	return &pb.ListAppUserResponse{Users: res}, nil
}

func (h *AppUserHandler) AppLogin(ctx context.Context, req *pb.AppLoginRequest) (*pb.AppLoginResponse, error) {
	token, err := h.usecase.AppLogin(req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	return &pb.AppLoginResponse{AccessToken: token}, nil
}

func (h *AppUserHandler) RequestResetPassword(ctx context.Context, req *pb.RequestResetPasswordRequest) (*pb.RequestResetPasswordResponse, error) {

	err := h.usecase.RequestPasswordReset(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	return &pb.RequestResetPasswordResponse{
		Message: "OTP sent to email",
	}, nil
}

func (h *AppUserHandler) VerifyPasswordResetOtp(ctx context.Context, req *pb.VerifyPasswordResetOtpRequest) (*pb.VerifyPasswordResetOtpResponse, error) {

	token, err := h.usecase.VerifyOtp(ctx, req.Email, req.Otp)
	if err != nil {
		return nil, err
	}

	return &pb.VerifyPasswordResetOtpResponse{
		ResetToken: token,
	}, nil
}

func (h *AppUserHandler) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	err := h.usecase.ResetPassword(ctx, req.ResetToken, req.NewPassword)

	if err != nil {
		return nil, err
	}

	return &pb.ResetPasswordResponse{
		Message: "password updated",
	}, nil
}

func (h *AppUserHandler) ListPlans(ctx context.Context, req *pb.ListPlansRequest) (*pb.ListPlansResponse, error) {

	plan, err := h.usecase.ListPlans(ctx)

	if err != nil {
		return nil, err
	}

	return mapUPlans(plan), nil
}

func (h *AppUserHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*emptypb.Empty, error) {
	userID, err := utils.GetUserIDFromContext(ctx)

	if err != nil {
		return &emptypb.Empty{}, errors.ErrUnauthenticated
	}

	if userID == req.UserId {
		return &emptypb.Empty{}, errors.ErrInvalidOperation
	}

	err = h.usecase.UpdateUser(ctx, req.UserId, req.Role, req.Password)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

func (h *AppUserHandler) UpdateUserStatus(ctx context.Context, req *pb.UpdateUserStatusRequest) (*emptypb.Empty, error) {

	userID, err := utils.GetUserIDFromContext(ctx)

	if err != nil {
		return &emptypb.Empty{}, errors.ErrUnauthenticated
	}

	if userID == req.UserId {
		return &emptypb.Empty{}, errors.ErrInvalidOperation
	}

	err = h.usecase.BlockUser(ctx, req.UserId, req.Status)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

func (h *AppUserHandler) ListPayments(ctx context.Context, req *pb.ListPaymentRequest) (*pb.ListPaymentResponse, error) {

	appID, err := utils.GetAppIDFromContext(ctx)

	if err != nil {
		return nil, err
	}

	payments, total, err := h.usecase.ListPayments(ctx, appID, req.Status, int(req.Limit), int(req.Offset), optionalTimestamp(req.StartDate), optionalTimestamp(req.EndDate))

	if err != nil {
		return nil, err
	}

	return mapPayments(payments, total, int(req.Limit), int(req.Offset)), nil
}

func (h *AppUserHandler) GetInvoice(ctx context.Context, req *pb.GetInvoiceRequest) (*pb.GetInvoiceResponse, error) {

	appID, err := utils.GetAppIDFromContext(ctx)

	if err != nil {
		return nil, err
	}

	pdfBytes, err := h.usecase.GetInvoice(ctx, appID, req.InvoiceId)
	if err != nil {
		return nil, err
	}

	return &pb.GetInvoiceResponse{
		Pdf:      pdfBytes,
		Filename: fmt.Sprintf("invoice-%s.pdf", req.InvoiceId),
	}, nil
}

func mapUPlans(plans []*entities.Plan) *pb.ListPlansResponse {
	resp := &pb.ListPlansResponse{}
	for _, p := range plans {
		resp.Plans = append(resp.Plans, &pb.UserPlan{
			PlanId:       p.PlanID,
			PlanName:     p.Name,
			MonthlyLimit: int32(p.MonthlyJobLimit),
			Planprice:    p.Price,
		})
	}
	return resp
}

func optionalTimestamp(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}

	value := ts.AsTime().UTC()
	return &value
}

func mapPayments(payment []entities.Payment, total, limit, offset int) *pb.ListPaymentResponse {
	resp := &pb.ListPaymentResponse{}
	for _, p := range payment {
		resp.Payments = append(resp.Payments, &pb.Payment{
			InvoiceId: p.InvoiceID,
			PlanName:  p.PlanName,
			Amount:    float64(p.Amount),
			Status:    p.Status,
			PaidAt:    formatUTC(p.PaidAt),
		})
	}

	resp.Total = int32(total)
	resp.Limit = int32(limit)
	resp.Offset = int32(offset)

	return resp
}
