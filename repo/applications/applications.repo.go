package repo

import (
	"errors"
	"fmt"
	"time"

	"github.com/badhouseplants/envspotting-apps/tools/logger"
	"github.com/badhouseplants/envspotting-go-proto/models/apps/applications"
	"github.com/badhouseplants/envspotting-go-proto/models/users/accounts"
	"github.com/badhouseplants/envspotting-go-proto/models/users/rights"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/badhouseplants/envspotting-apps/third_party/postgres"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/net/context"
)

// ApplicationStore represents methods to store applications
type ApplicationStore interface {
	Create(context.Context, *applications.AppWithoutContours) error
	Get(context.Context, *applications.AppId) (*applications.AppFullInfo, error)
	Delete(context.Context, *applications.AppId) (err error) 
	Update(context.Context, *applications.AppWithoutContours) error
	ListAdded(context.Context, applications.Applications_ListServer, *accounts.AccountsApps) error
	ListAvailable(context.Context, applications.Applications_ListServer, *rights.Applications) error
}

// ApplicationRepo implements ApplicationRepo
type ApplicationRepo struct {
	Pool      *pgxpool.Conn
	CreatedAt time.Time
}

// Create application (add to database)
func (store ApplicationRepo) Create(ctx context.Context, app *applications.AppWithoutContours) (err error) {
	const sql = "INSERT INTO applications (id, name, description) VALUES ($1, $2, $3);"
	var log = logger.GetGrpcLogger(ctx)
	_, err = store.Pool.Exec(ctx, sql, app.GetId(), app.GetName(), app.GetDescription())
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return status.Error(codes.AlreadyExists, err.Error())
			default:
				log.Error(err)
				return status.Error(codes.Internal, err.Error())
			}
		}
	}
	return nil
}

// Get application (from database)
func (store ApplicationRepo) Get(ctx context.Context, appIn *applications.AppId) (*applications.AppFullInfo, error) {
	const sql = "SELECT id, name, description, contours FROM applications WHERE id = $1"
	var (
		err    error
		appOut = &applications.AppFullInfo{}
		log    = logger.GetGrpcLogger(ctx)
	)
	err = store.Pool.QueryRow(ctx, sql, appIn.GetId()).Scan(&appOut.Id, &appOut.Name, &appOut.Description, &appOut.Contours)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("application with this id can't be found: %s", appIn.Id))
		} else {
			log.Error(err)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return appOut, nil
}

// Update applications (database update)
func (store ApplicationRepo) Update(ctx context.Context, app *applications.AppWithoutContours) (err error) {
	const sql = "UPDATE applications SET name=$2, description=$3 WHERE id=$1 RETURNING *"
	var log = logger.GetGrpcLogger(ctx)
	tag, err := store.Pool.Exec(ctx, sql, app.GetId(), app.GetName(), app.GetDescription())
	if tag.RowsAffected() == 0 {
		return status.Error(codes.NotFound, fmt.Sprintf("application with this id can't be found: %s", app.Id))
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return status.Error(codes.AlreadyExists, fmt.Sprintf("application with this name already exists: %s", app.Name))
			default:
				log.Error(err)
				return status.Error(codes.Internal, err.Error())
			}
		}
	}
	return nil
}


// List applications (streaming from database)
func (store ApplicationRepo) ListAvailable(ctx context.Context, stream applications.Applications_ListServer, apps *rights.Applications) error {
	// TODO: Add pagination @allanger
	const sql = "SELECT id, name, description FROM applications WHERE id = $1"
	var (
		db  = postgres.Pool(ctx)
		log = logger.GetGrpcLogger(ctx)
		app = &applications.AppWithoutContours{}
	)
	// Get applications
	rows, err := db.Query(ctx, sql, apps.GetApplicationId().GetId())
	if err != nil {
		log.Error(err)
		return status.Error(codes.Internal, err.Error())
	}
	// Stream applications
	for rows.Next() {
		// Scan apps into struct
		err = rows.Scan(&app.Id, &app.Name, &app.Description)
		if err != nil {
			log.Error(err)
			return status.Error(codes.Internal, err.Error())
		}
		// Send stream
		if err := stream.Send(app); err != nil {
			log.Error(err)
			return status.Error(codes.Internal, err.Error())
		}
	}
	return nil
}

// List applications (streaming from database)
func (store ApplicationRepo) ListAdded(ctx context.Context, stream applications.Applications_ListServer, apps *accounts.AccountsApps) error {
	// TODO: Add pagination @allanger
	const sql = "SELECT id, name, description FROM applications WHERE id=ANY(ARRAY($1));"
	var (
		db  = postgres.Pool(ctx)
		log = logger.GetGrpcLogger(ctx)
		app = &applications.AppWithoutContours{}
	)
	// Get applications
	rows, err := db.Query(ctx, sql, apps.Apps)
	if err != nil {
		log.Error(err)
		return status.Error(codes.Internal, err.Error())
	}
	// Stream applications
	for rows.Next() {
		// Scan apps into struct
		err = rows.Scan(&app.Id, &app.Name, &app.Description)
		if err != nil {
			log.Error(err)
			return status.Error(codes.Internal, err.Error())
		}
		// Send stream
		if err := stream.Send(app); err != nil {
			log.Error(err)
			return status.Error(codes.Internal, err.Error())
		}
	}
	return nil
}

// Get application (from database)
func (store ApplicationRepo) Delete(ctx context.Context, appIn *applications.AppId) (err error) {
	const sql = "DELETE FROM applications WHERE id = $1"
	var (
		log    = logger.GetGrpcLogger(ctx)
	)
	tag, err := store.Pool.Exec(ctx, sql, appIn.Id)
	if tag.RowsAffected() == 0 {
		return status.Error(codes.NotFound, fmt.Sprintf("application with this id can't be found: %s", appIn.Id))
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return status.Error(codes.NotFound, fmt.Sprintf("application with this id can't be found: %s", appIn.Id))
		} else {
			log.Error(err)
			return status.Error(codes.Internal, err.Error())
		}
	}
	return nil
}
