package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/badhouseplants/envspotting-apps/tools/logger"
	"github.com/badhouseplants/envspotting-go-proto/models/apps/contours"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ContourStore represents methods to store contour
type ContourStore interface {
	Create(context.Context, *contours.ContourInfoWithoutServices) error
	Get(context.Context, *contours.ContourId) (*contours.ContourInfo, error)
	Update(context.Context, *contours.ContourInfoWithoutServices) error
	List(context.Context, contours.Contours_ListServer, *contours.ContoursListOption) error
	Delete(context.Context, *contours.ContourIdAndName) error
	AddServices(context.Context, *contours.RepeatedServiceWithId) error
	RemoveService(context.Context, *contours.ServiceIdAndContourId) error
	GetAppIDByContourID(context.Context, string) (string, error) 
}

// ContourRepo implements ContoueRepo
type ContourRepo struct {
	Pool      *pgxpool.Conn
	CreatedAt time.Time
}

// Create a contour (add to db)
func (store ContourRepo) Create(ctx context.Context, contour *contours.ContourInfoWithoutServices) error {
	const sqlAddContour = "INSERT INTO contours (id, application_id, name, description) VALUES ($1, $2, $3, $4)"
	const sqlPairContourAndApp = "UPDATE applications SET contours = array_append(contours, $1) WHERE id=$2;"
	var log = logger.GetGrpcLogger(ctx)
	// Begin transaction
	tx, err := store.Pool.Begin(ctx)
	if err != nil {
		log.Error(err)
		return status.Error(codes.Internal, err.Error())
	}
	defer tx.Rollback(ctx)
	// Add contour
	_, err = store.Pool.Exec(ctx, sqlAddContour, contour.GetId(), contour.GetAppId(), contour.GetName(), contour.GetDescription())
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
	// Pair contour and app
	_, err = store.Pool.Exec(ctx, sqlPairContourAndApp, contour.Id, contour.AppId)
	if err != nil {
		log.Error(err)
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

// Get a contour (from db)
func (store ContourRepo) Get(ctx context.Context, contourIn *contours.ContourId) (*contours.ContourInfo, error) {
	const sql = "SELECT id, name, description, services FROM contours WHERE id = $1"
	var (
		err        error
		contourOut = &contours.ContourInfo{}
		log        = logger.GetGrpcLogger(ctx)
	)
	err = store.Pool.QueryRow(ctx, sql, contourIn.GetId()).Scan(&contourOut.Id, &contourOut.Name, &contourOut.Description, &contourOut.Services)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("contour with this id can't be found: %s", contourIn.Id))
		} else {
			log.Error(err)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return contourOut, nil
}

func (store ContourRepo) Update(ctx context.Context, contour *contours.ContourInfoWithoutServices) error {
	const sql = "UPDATE contours SET name=$2, description=$3 WHERE id=$1 RETURNING *"
	var log = logger.GetGrpcLogger(ctx)
	tag, err := store.Pool.Exec(ctx, sql, contour.GetId(), contour.GetName(), contour.GetDescription())
	if tag.RowsAffected() == 0 {
		return status.Error(codes.NotFound, fmt.Sprintf("contour with this id can't be found: %s", contour.Id))
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			default:
				log.Error(err)
				return status.Error(codes.Internal, err.Error())
			}
		}
	}
	return nil
}

func (store ContourRepo) List(ctx context.Context, stream contours.Contours_ListServer, options *contours.ContoursListOption) error {
	const sql = "SELECT id, name, description, services FROM contours WHERE contours.id=ANY(ARRAY(SELECT contours FROM applications WHERE applications.id=$1))"
	var (
		log     = logger.GetGrpcLogger(ctx)
		contour = &contours.ContourInfo{}
	)
	// Get contours
	rows, err := store.Pool.Query(ctx, sql, options.AppId)
	if err != nil {
		log.Error(err)
		return status.Error(codes.Internal, err.Error())
	}
	// Stream applications
	for rows.Next() {
		// Scan apps into struct
		err = rows.Scan(&contour.Id, &contour.Name, &contour.Description, &contour.Services)
		if err != nil {
			log.Error(err)
			return status.Error(codes.Internal, err.Error())
		}
		// Send stream
		if err := stream.Send(contour); err != nil {
			log.Error(err)
			return status.Error(codes.Internal, err.Error())
		}
	}
	return nil
}

// Delete a contour
func (store ContourRepo) Delete(ctx context.Context, contour *contours.ContourIdAndName) (err error) {
	const sql = `DELETE FROM contours 
	WHERE id=(
		SELECT id FROM contours WHERE contours.id=ANY(ARRAY(
			SELECT contours FROM applications WHERE applications.id=$2
			)
		) 
	  AND contours.id=$1);
`
	var log = logger.GetGrpcLogger(ctx)
	tag, err := store.Pool.Exec(ctx, sql, contour.Id, contour.AppId)
	if tag.RowsAffected() == 0 {
		return status.Error(codes.NotFound, fmt.Sprintf("contour with this id (%s) doesn't belong to the application %s", contour.Id, contour.AppId))
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return status.Error(codes.NotFound, fmt.Sprintf("contour with this id can't be found: %s", contour.Id))
		} else {
			log.Error(err)
			return status.Error(codes.Internal, err.Error())
		}
	}
	return nil
}

func (store ContourRepo) AddServices(ctx context.Context, contour *contours.RepeatedServiceWithId) (err error) {
	const sql = `UPDATE contours SET services = (
    CASE
        WHEN services IS NULL THEN '[]'::JSONB
        ELSE services
    END
) || $2::JSONB WHERE id = $1;`

	var log = logger.GetGrpcLogger(ctx)
	tag, err := store.Pool.Exec(ctx, sql, contour.GetContourId(), contour.GetServices())
	if tag.RowsAffected() == 0 {
		return status.Error(codes.NotFound, fmt.Sprintf("contour with this id can't be found: %s", contour.GetContourId()))
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			default:
				log.Error(err)
				return status.Error(codes.Internal, err.Error())
			}
		}
	}
	return nil
}

func (store ContourRepo) RemoveService(ctx context.Context, in *contours.ServiceIdAndContourId) error {
	const sql = `
UPDATE contours c SET services = array_to_json(
ARRAY(SELECT obj.val AS count
	FROM   contours c
	JOIN   LATERAL jsonb_array_elements(c.services) obj(val) ON obj.val->>'id' != $2
	WHERE  c.id = $1 )
) WHERE c.id = $1;
`
	var log = logger.GetGrpcLogger(ctx)
	tag, err := store.Pool.Exec(ctx, sql, in.GetContourId(), in.GetServiceId())
	if tag.RowsAffected() == 0 {
		return status.Error(codes.NotFound, "no rows affected")
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			default:
				log.Error(err)
				return status.Error(codes.Internal, err.Error())
			}
		}
	}
	return nil
}

func (store ContourRepo) GetAppIDByContourID(ctx context.Context, contourID string) (string, error) {
	const sql = "SELECT application_id FROM contours WHERE id = $1"
	var appID string
	var log = logger.GetGrpcLogger(ctx)
	err := store.Pool.QueryRow(ctx, sql, contourID).Scan(&appID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			default:
				log.Error(err)
				return "", status.Error(codes.Internal, err.Error())
			}
		}
	}
	return appID, nil
}