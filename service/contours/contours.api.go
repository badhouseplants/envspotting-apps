package service

import (
	"context"

	"github.com/badhouseplants/envspotting-go-proto/models/apps/applications"
	"github.com/badhouseplants/envspotting-go-proto/models/apps/contours"
	"github.com/badhouseplants/envspotting-go-proto/models/common"
	"github.com/badhouseplants/envspotting-go-proto/models/users/rights"

	// "github.com/badhouseplants/envspotting-go-proto/models/users/rights"

	grpcusers "github.com/badhouseplants/envspotting-apps/internal/grpc-users"
	"github.com/badhouseplants/envspotting-apps/tools/logger"
	"google.golang.org/grpc"
)

type contoursGrpcServer struct {
	contours.UnimplementedContoursServer
}

func Register(grpcServer *grpc.Server) {
	contours.RegisterContoursServer(grpcServer, &contoursGrpcServer{})
}

func (s *contoursGrpcServer) Create(ctx context.Context, in *contours.ContourNameAndDescription) (*contours.ContourInfoWithoutServices, error) {
	logger.EnpointHit(ctx)
	_, err := grpcusers.AuthorizationClient.ValidateToken(ctx, &common.EmptyMessage{})
	if err != nil {
		return nil, err
	}
	_, err = grpcusers.RightsClient.CheckRight(ctx, &rights.AccessRightRequest{
		ApplicationId: &applications.AppId{Id: in.AppId},
		AccessRight:   rights.AccessRights_ACCESS_RIGHTS_WRITE,
	})
	if err != nil {
		return nil, err
	}
	return Create(ctx, in)
}

func (s *contoursGrpcServer) Get(ctx context.Context, in *contours.ContourId) (*contours.ContourInfo, error) {
	logger.EnpointHit(ctx)
	_, err := grpcusers.AuthorizationClient.ValidateToken(ctx, &common.EmptyMessage{})
	if err != nil {
		return nil, err
	}

	appID, err := GetAppIDByContourID(ctx, in.Id)
	if err != nil {
		return nil, err
	}

	_, err = grpcusers.RightsClient.CheckRight(ctx, &rights.AccessRightRequest{
		ApplicationId: &applications.AppId{Id: appID.GetId()},
		AccessRight:   rights.AccessRights_ACCESS_RIGHTS_READ_UNSPECIFIED,
	})
	if err != nil {
		return nil, err
	}
	return Get(ctx, in)
}

func (s *contoursGrpcServer) Update(ctx context.Context, in *contours.ContourInfoWithoutServices) (*contours.ContourInfoWithoutServices, error) {
	logger.EnpointHit(ctx)
	_, err := grpcusers.AuthorizationClient.ValidateToken(ctx, &common.EmptyMessage{})
	if err != nil {
		return nil, err
	}
	appID, err := GetAppIDByContourID(ctx, in.Id)
	if err != nil {
		return nil, err
	}

	_, err = grpcusers.RightsClient.CheckRight(ctx, &rights.AccessRightRequest{
		ApplicationId: &applications.AppId{Id: appID.GetId()},
		AccessRight:   rights.AccessRights_ACCESS_RIGHTS_WRITE,
	})
	if err != nil {
		return nil, err
	}

	return Update(ctx, in)
}

func (s *contoursGrpcServer) List(in *contours.ContoursListOption, stream contours.Contours_ListServer) error {
	logger.EnpointHit(stream.Context())
	_, err := grpcusers.AuthorizationClient.ValidateToken(stream.Context(), &common.EmptyMessage{})
	if err != nil {
		return err
	}
	_, err = grpcusers.RightsClient.CheckRight(stream.Context(), &rights.AccessRightRequest{
		ApplicationId: &applications.AppId{Id: in.AppId},
		AccessRight:   rights.AccessRights_ACCESS_RIGHTS_READ_UNSPECIFIED,
	})
	if err != nil {
		return err
	}
	err = List(stream.Context(), stream, in)
	if err != nil {
		return err
	}
	return nil
}

func (s *contoursGrpcServer) Delete(ctx context.Context, in *contours.ContourIdAndName) (*common.EmptyMessage, error) {
	logger.EnpointHit(ctx)
	_, err := grpcusers.AuthorizationClient.ValidateToken(ctx, &common.EmptyMessage{})
	if err != nil {
		return nil, err
	}

	_, err = grpcusers.RightsClient.CheckRight(ctx, &rights.AccessRightRequest{
		ApplicationId: &applications.AppId{Id: in.AppId},
		AccessRight:   rights.AccessRights_ACCESS_RIGHTS_WRITE,
	})
	if err != nil {
		return nil, err
	}

	return Delete(ctx, in)
}

func (s *contoursGrpcServer) AddServices(ctx context.Context, in *contours.RepeatedServiceWithoutId) (*common.EmptyMessage, error) {
	logger.EnpointHit(ctx)
	_, err := grpcusers.AuthorizationClient.ValidateToken(ctx, &common.EmptyMessage{})
	if err != nil {
		return nil, err
	}

	_, err = GetAppIDByContourID(ctx, in.ContourId)
	if err != nil {
		return nil, err
	}

	_, err = grpcusers.RightsClient.CheckRight(ctx, &rights.AccessRightRequest{
		ApplicationId: &applications.AppId{Id: in.AppId},
		AccessRight:   rights.AccessRights_ACCESS_RIGHTS_DELETE,
	})

	if err != nil {
		return nil, err
	}

	return AddServices(ctx, in)
}

func (s *contoursGrpcServer) RemoveService(ctx context.Context, in *contours.ServiceIdAndContourId) (*common.EmptyMessage, error) {
	logger.EnpointHit(ctx)
	_, err := grpcusers.AuthorizationClient.ValidateToken(ctx, &common.EmptyMessage{})
	if err != nil {
		return nil, err
	}

	_, err = GetAppIDByContourID(ctx, in.ContourId)
	if err != nil {
		return nil, err
	}

	_, err = grpcusers.RightsClient.CheckRight(ctx, &rights.AccessRightRequest{
		ApplicationId: &applications.AppId{Id: in.AppId},
		AccessRight:   rights.AccessRights_ACCESS_RIGHTS_WRITE,
	})

	return RemoveService(ctx, in)
}
