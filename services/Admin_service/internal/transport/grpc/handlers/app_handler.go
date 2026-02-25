package handlers

import (
	"admin_service/internal/domain/entities"
	pb "admin_service/internal/proto/admin"
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
)

type AppHandler struct {
	pb.UnimplementedAppServiceServer 
	usecase interface {
		Register(string,string)(string,error) 
		List()([]*entities.App,error)
		Block(string)error 
		Unblock(string)error 
	}
}

func NewAppHandler(u interface{
	Register(string,string)(string,error) 
	List() ([]*entities.App, error)
	Block(string) error
	Unblock(string) error
}) *AppHandler {
	return &AppHandler{usecase: u}
}

func (h *AppHandler) RegisterApp(ctx context.Context, req *pb.RegisterAppRequest) (*pb.RegisterAppResponse,error) {
	app_id,err := h.usecase.Register(req.AppName,req.AppEmail)
	return &pb.RegisterAppResponse{AppId: app_id},err 
}

func (h *AppHandler) ListApps(ctx context.Context,_ *pb.ListAppsRequest)(*pb.ListAppsResponse,error) {
	apps,err := h.usecase.List()
	if err != nil {
		return nil ,err 
	}

	var result []*pb.App
	for _,a := range apps {
		result = append(result, &pb.App{
			AppId: a.AppID,
			AppName: a.AppName,
			AppEmail: a.AppEmail,
			Status: a.Status,
			Tier: a.Tier,
		})
	}

	return &pb.ListAppsResponse{Apps: result},nil 
}

func (h *AppHandler) BlockApp(ctx context.Context, req *pb.BlockAppRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, h.usecase.Block(req.AppId)
}

func (h *AppHandler) UnblockApp(ctx context.Context, req *pb.BlockAppRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, h.usecase.Unblock(req.AppId)
}