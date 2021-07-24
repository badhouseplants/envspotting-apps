package service

import (
	"context"
	"time"

	repo "github.com/badhouseplants/envspotting-apps/repo/contours"
	"github.com/badhouseplants/envspotting-apps/third_party/postgres"
	"github.com/badhouseplants/envspotting-go-proto/models/apps/applications"
	"github.com/badhouseplants/envspotting-go-proto/models/apps/contours"
	"github.com/badhouseplants/envspotting-go-proto/models/common"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var apprepo repo.ContourStore

var initRepo = func(ctx context.Context) repo.ContourStore {
	if apprepo == nil {
		apprepo = repo.ContourRepo{
			Pool:      postgres.Pool(ctx),
			CreatedAt: time.Now(),
		}
	}
	return apprepo
}

// Create a new contour
func Create(ctx context.Context, in *contours.ContourNameAndDescription) (*contours.ContourInfoWithoutServices, error) {
	repo := initRepo(ctx)
	// Init contour struct
	contour := &contours.ContourInfoWithoutServices{
		Id:          uuid.NewString(),
		Name:        in.Name,
		Description: in.Description,
		AppId:       in.AppId,
	}
	// Create contour
	if err := repo.Create(ctx, contour); err != nil {
		return nil, err
	}
	return contour, nil
}

// Get a contour
func Get(ctx context.Context, in *contours.ContourId) (*contours.ContourInfo, error) {
	repo := initRepo(ctx)
	app, err := repo.Get(ctx, in)
	if err != nil {
		return nil, err
	}
	return app, nil
}

// Update a contour
func Update(ctx context.Context, contour *contours.ContourInfoWithoutServices) (*contours.ContourInfoWithoutServices, error) {
	repo := initRepo(ctx)
	err := repo.Update(ctx, contour)
	if err != nil {
		return nil, err
	}
	return contour, nil
}

// List contours
func List(ctx context.Context, stream contours.Contours_ListServer, options *contours.ContoursListOption) error {
	repo := initRepo(ctx)
	err := repo.List(ctx, stream, options)
	if err != nil {
		return err
	}
	return nil
}

// Delete a contour
func Delete(ctx context.Context, in *contours.ContourIdAndName) (out *common.EmptyMessage, err error) {
	repo := initRepo(ctx)
	contourID := &contours.ContourId{
		Id: in.Id,
	}
	appGotten, err := repo.Get(ctx, contourID)
	if err != nil {
		return nil, err
	}
	if appGotten.Name == in.Name {
		err := repo.Delete(ctx, in)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, status.Error(codes.Aborted, "to delete a contour you should provide a correct name")
	}
	return &common.EmptyMessage{}, nil
}

// AddServices to a contour
func AddServices(ctx context.Context, in *contours.RepeatedServiceWithoutId) (*common.EmptyMessage, error) {
	repo := initRepo(ctx)
	servicesWithID := &contours.RepeatedServiceWithId{
		ContourId: in.GetContourId(),
	}
	for _, service := range in.Services {
		serviceInfo := &contours.ServiceInfo{
			Id:          uuid.NewString(),
			Project:     service.Project,
			Environment: service.Environment,
		}
		servicesWithID.Services = append(servicesWithID.Services, serviceInfo)
	}
	err := repo.AddServices(ctx, servicesWithID)
	if err != nil {
		return nil, err
	}
	return &common.EmptyMessage{}, nil
}

// RemoveService from contour
func RemoveService(ctx context.Context, in *contours.ServiceIdAndContourId) (*common.EmptyMessage, error) {
	repo := initRepo(ctx)
	err := repo.RemoveService(ctx, in)
	if err != nil {
		return nil, err
	}
	return &common.EmptyMessage{}, nil
}

func GetAppIDByContourID(ctx context.Context, contourID string) (*applications.AppId, error) {
	repo := initRepo(ctx)
	appId, err := repo.GetAppIDByContourID(ctx, contourID)
	if err != nil {
		return nil, err
	}
	return &applications.AppId{
		Id: appId,
	}, err
}
