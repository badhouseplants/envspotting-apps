package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/badhouseplants/envspotting-apps/tools/logger"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/viper"
)

// ConnectionParams to open database pool
type ConnectionParams struct {
	Username string
	Password string
	Database string
	Host     string
	Port     string
}

func NewConnectionParams() *ConnectionParams {
	return &ConnectionParams{
		Username: viper.GetString("database_username"),
		Password: viper.GetString("database_password"),
		Database: viper.GetString("database_name"),
		Host:     viper.GetString("database_host"),
		Port:     viper.GetString("database_port"),
	}
}

var (
	pool       *pgxpool.Pool
	maxRetries = 5
)

var errCantConnect = "unable to connect to database"

func Pool(ctx context.Context) *pgxpool.Conn {
	log := logger.GetServerLogger()
	if pool == nil {
		err := openConnectionPool(ctx)
		if err != nil {
			return nil
		}
	}
	conn, err := pool.Acquire(ctx)

	if err == pgx.ErrDeadConn {
		if maxRetries > 0 {
			log.Infof("%v, try acquiring a non-dead connection", err)
			pool.Close()
			openConnectionPool(ctx)
			return Pool(ctx)
		}
	}
	log.Info(pool.Stat().AcquiredConns())

	return conn
}

func openConnectionPool(ctx context.Context) error {
	var err error
	log := logger.GetServerLogger()
	params := NewConnectionParams()
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", params.Username, params.Password, params.Host, params.Port, params.Database)
	for i := 0; i < maxRetries; i++ {
		pool, err = pgxpool.Connect(ctx, connectionString)
		if err != nil {
			log.Error(err)
			time.Sleep(5 * time.Second)
			continue
		}
		return nil
	}
	log.Error(errCantConnect)
	return err
}
