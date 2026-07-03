package handlers

import (
	"admin_service/internal/domain/entities"
	pb "admin_service/internal/proto/admin"
	"admin_service/internal/usecase/service"
	"context"
	"fmt"
	"time"
)

type AdminHandler struct {
	pb.UnimplementedAdminServiceServer
	usecase interface {
		Login(string, string) (string, error)
		ListPayment(ctx context.Context, appID, status string, limit, offset int, startDate, endDate *time.Time) ([]entities.Payment, int, error)
		GetInvoice(ctx context.Context, invoice_id string) ([]byte, error)
		ListSubcribers(ctx context.Context, limit, offset int, startDate, endDate *time.Time)([]entities.Subscriber,int,error)
		GetOverview(ctx context.Context) (*entities.DashboardOverview,error)
	}
	plan *service.PlanService
}

func NewAdminHandler(u interface {
	Login(string, string) (string, error)
	ListPayment(ctx context.Context, appID, status string, limit, offset int, startDate, endDate *time.Time) ([]entities.Payment, int, error)
	GetInvoice(ctx context.Context, invoice_id string) ([]byte, error)
	ListSubcribers(ctx context.Context, limit, offset int, startDate, endDate *time.Time)([]entities.Subscriber,int,error)
	GetOverview(ctx context.Context) (*entities.DashboardOverview,error)
}, p *service.PlanService) *AdminHandler {
	return &AdminHandler{usecase: u, plan: p}
}

func (h *AdminHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	token, err := h.usecase.Login(req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	return &pb.LoginResponse{AccessToken: token}, nil
}

func (h *AdminHandler) CreatePlan(ctx context.Context, req *pb.CreatePlanRequest) (*pb.CreatePlanResponse, error) {
	err := h.plan.CreatePlan(req.Name, int(req.Joblimit), req.Price, req.StripePriceID)
	if err != nil {
		return nil, err
	}
	return &pb.CreatePlanResponse{Message: "Plan created"}, nil
}

func (h *AdminHandler) ListPlans(ctx context.Context, req *pb.ListPlanRequest) (*pb.ListPlanResponse, error) {
	plans, err := h.plan.ListPlan()
	if err != nil {
		return nil, err
	}

	return mapPlans(plans), nil
}

func (h *AdminHandler) UpdatePlan(ctx context.Context, req *pb.UpdatePlanRequest) (*pb.UpdatePlanResponse, error) {
	plan, err := h.plan.UpdatePlans(req.PlanID, req.Name, int(req.JobLimit), req.Price, req.StripePriceID)
	if err != nil {
		return nil, err
	}

	resp := &pb.Plan{
		PlanID:   plan.PlanID,
		Name:     plan.Name,
		JobLimit: int32(plan.MonthlyJobLimit),
		Price:    plan.Price,
	}

	return &pb.UpdatePlanResponse{
		UpdatedPlan: resp,
	}, nil
}

func (h *AdminHandler) UpdatePlanStatus(ctx context.Context, req *pb.UpdatePlanStatusRequest) (*pb.UpdatePlanStatusResponse, error) {

	plan, err := h.plan.UpdateStatusPlan(req.PlanID, req.Status)

	if err != nil {
		return nil, err
	}

	resp := &pb.Plan{
		PlanID:        plan.PlanID,
		Name:          plan.Name,
		JobLimit:      int32(plan.MonthlyJobLimit),
		Price:         plan.Price,
		StripePriceID: plan.StripeID,
		Status:        plan.Status,
	}

	return &pb.UpdatePlanStatusResponse{
		Plan: resp,
	}, nil
}

func (h *AdminHandler) ListPayments(ctx context.Context, req *pb.ListPaymentAdminRequest) (*pb.ListPaymentAdminResponse, error) {

	payments, total, err := h.usecase.ListPayment(ctx, req.AppId, req.Status, int(req.Limit), int(req.Offset), optionalTimestamp(req.StartDate), optionalTimestamp(req.EndDate))

	if err != nil {
		return nil, err
	}

	return mapAdminPayments(payments, total, int(req.Limit), int(req.Offset)), nil
}

func (h *AdminHandler) GetSubscribers(ctx context.Context, req *pb.GetSubscriberRequest)(*pb.GetSubscriberResponse,error) {

	subscribers,total,err := h.usecase.ListSubcribers(ctx,int(req.Limit),int(req.Offset),optionalTimestamp(req.StartDate),optionalTimestamp(req.EndDate))

	if err != nil {
		return nil,err 
	}

	return mapSubcribers(subscribers,total,int(req.Limit),int(req.Offset)),nil 
}

func (h *AdminHandler) GetAdminInvoice(ctx context.Context, req *pb.GetAdminInvoiceRequest) (*pb.GetAdminInvoiceResponse, error) {

	pdfBytes, err := h.usecase.GetInvoice(ctx, req.InvoiceId)
	if err != nil {
		return nil, err
	}
	fileName := fmt.Sprintf("invoice-%s.pdf", req.InvoiceId)
	return &pb.GetAdminInvoiceResponse{
		Pdf:      pdfBytes,
		Filename: fileName,
	}, nil
}

func (h *AdminHandler) GetDashboardOverview(ctx context.Context, req *pb.GetDashboardOverviewRequest)(*pb.GetDashboardOverviewResponse,error) {

	overview,err := h.usecase.GetOverview(ctx)

	if err != nil {
		return nil,err 
	}

	return &pb.GetDashboardOverviewResponse{
		TotalUsers: overview.TotalUsers,
		TotalApps: overview.TotalApps,
		ActiveSubscribers: overview.ActiveSubscribers,
		RevenueMonth: overview.RevenueMonth,
		RevenueLastMonth: overview.RevenueLastMonth,
		JobsToday: overview.JobsToday,
		FailedJobsToday: overview.FailedJobsToday,
	},nil
}

func mapPlans(plans []*entities.Plan) *pb.ListPlanResponse {
	resp := &pb.ListPlanResponse{}
	for _, p := range plans {
		resp.Plans = append(resp.Plans, &pb.Plan{
			PlanID:   p.PlanID,
			Name:     p.Name,
			Status:   p.Status,
			JobLimit: int32(p.MonthlyJobLimit),
			Price:    p.Price,
		})
	}
	return resp
}

func mapAdminPayments(payments []entities.Payment, total, limit, offset int) *pb.ListPaymentAdminResponse {
	resp := &pb.ListPaymentAdminResponse{}
	for _, p := range payments {
		resp.Payments = append(resp.Payments, &pb.PaymentAdmin{
			InvoiceId:  p.InvoiceID,
			PlanName:   p.PlanName,
			Amount:     float64(p.Amount),
			Status:     p.Status,
			PaidAt:     formatUTC(p.PaidAt),
			AppName:    p.AppName,
			AppEmail:   p.CustomerEmail,
			PlanAmount: float64(p.PlanAmount),
		})
	}

	resp.Total = int32(total)
	resp.Limit = int32(limit)
	resp.Offset = int32(offset)

	return resp
}

func mapSubcribers(subcribers []entities.Subscriber, total, limit, offset int) *pb.GetSubscriberResponse {

	resp := &pb.GetSubscriberResponse{}

	for _,s := range subcribers {
		resp.Subscribers = append(resp.Subscribers, &pb.Subscriber{
			AppId: s.AppID,
			AppName: s.AppName,
			PlanName: s.PlanName,
			Status: s.Status,
			StartDate: formatUTC(s.StartDate),
			EndDate: formatUTC(s.EndDate),
		})
	}

	resp.Total = int32(total)
	resp.Limit = int32(limit)
	resp.Offset = int32(offset)

	return resp
}