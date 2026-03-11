package handlers

import (
	"admin_service/internal/domain/entities"
	pb "admin_service/internal/proto/admin"
	"admin_service/internal/usecase/service"
	"context"
)

type AdminHandler struct {
	pb.UnimplementedAdminServiceServer
	usecase interface {
		Login(string,string)(string,error)
	}
	plan *service.PlanService
}

func NewAdminHandler(u interface{
	Login(string,string)(string,error)
},p *service.PlanService)*AdminHandler {
	return &AdminHandler{usecase: u,plan: p}
}

func (h *AdminHandler) Login(ctx context.Context,req *pb.LoginRequest)(*pb.LoginResponse,error) {
	token,err := h.usecase.Login(req.Email,req.Password)
	if err != nil {
		return nil,err 
	}
	return &pb.LoginResponse{AccessToken: token},nil 
}

func (h *AdminHandler) CreatePlan(ctx context.Context,req *pb.CreatePlanRequest)(*pb.CreatePlanResponse,error) {
	err := h.plan.CreatePlan(req.Name,int(req.Joblimit),req.Price)
	if err != nil {
		return nil,err 
	}
	return &pb.CreatePlanResponse{Message: "Plan created"},nil 
}

func (h *AdminHandler) ListPlans(ctx context.Context,req *pb.ListPlanRequest)(*pb.ListPlanResponse,error) {
	plans,err := h.plan.ListPlan()
	if err != nil {
		return nil,err 
	}

	return mapPlans(plans),nil 
}

func (h *AdminHandler) UpdatePlan(ctx context.Context,req *pb.UpdatePlanRequest)(*pb.UpdatePlanResponse,error) {
	plan,err := h.plan.UpdatePlans(req.PlanID,req.Name,int(req.JobLimt),req.Price)
	if err != nil {
		return nil,err 
	}

	resp := &pb.Plan{
		PlanID: plan.PlanID,
		Name: plan.Name,
		JobLimit: int32(plan.MonthlyJobLimit),
		Price: plan.Price,
	}

	return &pb.UpdatePlanResponse{
		UpdatedPlan: resp,
	},nil 
}

func mapPlans(plans []*entities.Plan) *pb.ListPlanResponse {
	resp := &pb.ListPlanResponse{}
	for _,p := range plans {
		resp.Plans = append(resp.Plans, &pb.Plan{
			PlanID: p.PlanID,
			Name: p.Name,
			JobLimit: int32(p.MonthlyJobLimit),
			Price: p.Price,
		})
	}
	return resp 
}