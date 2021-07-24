package service

import (
	context "context"

	"github.com/badhouseplants/envspotting-go-proto/models/apps/applications"
	"github.com/badhouseplants/envspotting-go-proto/models/common"
	"github.com/badhouseplants/envspotting-go-proto/models/users/rights"

	// rs "github.com/badhouseplants/envspotting-apps/service/users/rights"

	grpcusers "github.com/badhouseplants/envspotting-apps/internal/grpc-users"
	"github.com/badhouseplants/envspotting-apps/tools/logger"
	"github.com/badhouseplants/envspotting-apps/tools/metadata"
	"google.golang.org/grpc"
)

type applicationsGrpcImpl struct {
	applications.UnimplementedApplicationsServer
}

func Register(grpcServer *grpc.Server) {
	applications.RegisterApplicationsServer(grpcServer, &applicationsGrpcImpl{})
}

func (s *applicationsGrpcImpl) Create(ctx context.Context, in *applications.AppNameAndDescription) (*applications.AppWithoutContours, error) {
	logger.EnpointHit(ctx)
	ctx = metadata.MetadataInternalProxy(ctx)
	_, err := grpcusers.AuthorizationClient.ValidateToken(ctx, &common.EmptyMessage{})
	if err != nil {
		return nil, err
	}

	return Create(ctx, in)
}

func (s *applicationsGrpcImpl) Get(ctx context.Context, in *applications.AppId) (*applications.AppFullInfo, error) {
	logger.EnpointHit(ctx)
	ctx = metadata.MetadataInternalProxy(ctx)
	_, err := grpcusers.AuthorizationClient.ValidateToken(ctx, &common.EmptyMessage{})
	if err != nil {
		return nil, err
	}

	rightRequest := &rights.AccessRightRequest{
		ApplicationId: in,
		AccessRight:   rights.AccessRights_ACCESS_RIGHTS_READ_UNSPECIFIED,
	}
	_, err = grpcusers.RightsClient.CheckRight(ctx, rightRequest)
	if err != nil {
		return nil, err
	}
	return Get(ctx, in)
}

func (s *applicationsGrpcImpl) Update(ctx context.Context, in *applications.AppWithoutContours) (*applications.AppWithoutContours, error) {
	logger.EnpointHit(ctx)
	ctx = metadata.MetadataInternalProxy(ctx)
	_, err := grpcusers.AuthorizationClient.ValidateToken(ctx, &common.EmptyMessage{}) 
	if err != nil {
		return nil, err
	}

	_, err = grpcusers.RightsClient.CheckRight(ctx, &rights.AccessRightRequest{
		ApplicationId: &applications.AppId{Id: in.Id},
		AccessRight:   rights.AccessRights_ACCESS_RIGHTS_WRITE,
	})
	if err != nil {
		return nil, err
	}

	return Update(ctx, in)
}

func (s *applicationsGrpcImpl) Delete(ctx context.Context, in *applications.AppIdAndName) (*common.EmptyMessage, error) {
	logger.EnpointHit(ctx)
	ctx = metadata.MetadataInternalProxy(ctx)
	_, err := grpcusers.AuthorizationClient.ValidateToken(ctx, &common.EmptyMessage{}) 
	if err != nil {
		return nil, err
	}

	rightRequest := &rights.AccessRightRequest{
		ApplicationId: &applications.AppId{Id: in.GetId()},
		AccessRight:   rights.AccessRights_ACCESS_RIGHTS_DELETE,
	}
	_, err = grpcusers.RightsClient.CheckRight(ctx, rightRequest)
	if err != nil {
		return nil, err
	}

	return Delete(ctx, in)
}

func (s *applicationsGrpcImpl) List(in *applications.ListOptions, stream applications.Applications_ListServer) error {
	logger.EnpointHit(stream.Context())
	ctx := metadata.MetadataInternalProxy(stream.Context())
	_, err := grpcusers.AuthorizationClient.ValidateToken(ctx, &common.EmptyMessage{}) 
	if err != nil {
		return err
	}

	err = List(stream.Context(), stream, in)
	if err != nil {
		return err
	}
	return nil
}
