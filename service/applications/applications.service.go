package service

import (
	context "context"
	"io"
	"time"

	grpcusers "github.com/badhouseplants/envspotting-apps/internal/grpc-users"
	repo "github.com/badhouseplants/envspotting-apps/repo/applications"
	"github.com/badhouseplants/envspotting-apps/tools/logger"
	"github.com/badhouseplants/envspotting-apps/tools/metadata"
	"github.com/badhouseplants/envspotting-go-proto/models/users/rights"

	// rightsService "github.com/badhouseplants/envspotting-apps/service/users/rights"
	"github.com/badhouseplants/envspotting-apps/third_party/postgres"
	"github.com/badhouseplants/envspotting-go-proto/models/apps/applications"
	"github.com/badhouseplants/envspotting-go-proto/models/common"
	"github.com/google/uuid"
)

// var apprepo repo.ApplicationStore

var initRepo = func(ctx context.Context) repo.ApplicationStore {
	conn := postgres.Pool(ctx)
	apprepo := repo.ApplicationRepo{
		Pool:      conn,
		CreatedAt: time.Now(),
	}
	return apprepo
}

// Create a new application
func Create(ctx context.Context, in *applications.AppNameAndDescription) (*applications.AppWithoutContours, error) {
	log := logger.GetGrpcLogger(ctx)
	repo := initRepo(ctx)
	// Application info struct
	app := &applications.AppWithoutContours{
		Id:          uuid.New().String(),
		Name:        in.GetName(),
		Description: in.GetDescription(),
	}
	// Create application
	if err := repo.Create(ctx, app); err != nil {
		return nil, err
	}
	userID, err := grpcusers.AuthorizationClient.ParseIdFromToken(ctx, &common.EmptyMessage{})
	if err != nil {
		appToDelete := &applications.AppIdAndName{
			Id:   app.Id,
			Name: app.Name,
		}
		Delete(ctx, appToDelete)
		log.Error(err)
		return nil, err
	}
	accessRight := &rights.AccessRuleWithoutId{
		UserId:        userID.GetId(),
		ApplicationId: app.GetId(),
		AccessRight:   rights.AccessRights_ACCESS_RIGHTS_DELETE,
	}
	_, err = grpcusers.RightsClient.Init(ctx, accessRight)
	if err != nil {
		appToDelete := &applications.AppIdAndName{
			Id:   app.Id,
			Name: app.Name,
		}
		Delete(ctx, appToDelete)
		log.Error(err)
		return nil, err
	}

	// Add application to the accounts
	appId := &applications.AppId{
		Id: app.Id,
	}
	_, err = grpcusers.AccountClient.AddAppToUser(ctx, appId)
	if err != nil {
		return nil, err
	}

	return app, nil
}

// Get application by name
func Get(ctx context.Context, appId *applications.AppId) (*applications.AppFullInfo, error) {
	repo := initRepo(ctx)
	app, err := repo.Get(ctx, appId)
	if err != nil {
		return nil, err
	}
	return app, nil
}

// Update application
func Update(ctx context.Context, app *applications.AppWithoutContours) (*applications.AppWithoutContours, error) {
	repo := initRepo(ctx)
	err := repo.Update(ctx, app)
	if err != nil {
		return nil, err
	}
	return app, nil
}

// Update application
func Delete(ctx context.Context, app *applications.AppIdAndName) (*common.EmptyMessage, error) {
	repo := initRepo(ctx)
	appId := &applications.AppId{
		Id: app.Id,
	}
	appGotten, err := repo.Get(ctx, appId)
	if err != nil {
		return nil, err
	}
	if appGotten.Name == app.Name {
		err := repo.Delete(ctx, appId)
		if err != nil {
			return nil, err
		}
	}
	return &common.EmptyMessage{}, nil
}

// List applications
// TODO: add paging @allanger
func List(ctx context.Context, stream applications.Applications_ListServer, options *applications.ListOptions) error {
	repo := initRepo(ctx)
	var (
		log     = logger.GetGrpcLogger(ctx)
		appsArr []string
	)
	userID, err := grpcusers.AuthorizationClient.ParseIdFromToken(
		metadata.MetadataInternalProxy(stream.Context()),
		&common.EmptyMessage{},
	)
	if err != nil {
		return err
	}

	if options.Added {
		apps, err := grpcusers.AccountClient.GetAppsFromUser(
			metadata.MetadataInternalProxy(stream.Context()),
			userID,
		)
		if err != nil {
			return err
		}
		err = repo.ListAdded(ctx, stream, apps)
		if err != nil {
			return err
		}
	} else if !options.Added {
		apps, err := grpcusers.RightsClient.ListAvailableApps(
			metadata.MetadataInternalProxy(stream.Context()),
			&rights.AvailableAppsListOptions{AccountId: userID},
		)
		if err != nil {
			log.Errorf("open stream error %v", err)
			return err
		}

		for {
			availableApps, err := apps.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Errorf("cannot receive %v", err)
			}
			appsArr = append(appsArr, availableApps.GetApplicationId().GetId())
		}
		repo.ListAvailable(stream.Context(), stream, appsArr)
		log.Printf("finished")
		if err != nil {
			return err
		}
		// err = repo.ListAvailable(ctx, stream, apps)
	}
	if err != nil {
		return err
	}
	return nil
}
