package main

import (
	"fmt"
	"net"

	applications "github.com/badhouseplants/envspotting-apps/service/applications"
	contours "github.com/badhouseplants/envspotting-apps/service/contours"

	grpcusers "github.com/badhouseplants/envspotting-apps/internal/grpc-users"
	"github.com/badhouseplants/envspotting-apps/migrations"
	"github.com/badhouseplants/envspotting-apps/tools/logger"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
)

var (
	host string
)

func init() {
	// app variables
	viper.SetDefault("envspotting_apps_host", "0.0.0.0")
	viper.SetDefault("envspotting_apps_port", "9090")
	viper.SetDefault("envspotting_users_host", "0.0.0.0")
	viper.SetDefault("envspotting_users_port", "9090")
	viper.SetDefault("database_username", "docker_user")
	viper.SetDefault("database_password", "qwertyu9")
	viper.SetDefault("database_name", "applications")
	viper.SetDefault("database_host", "localhost")
	viper.SetDefault("database_port", "5432")
	viper.AutomaticEnv() // read in environment variables that match)
}

func main() {
	log := logger.GetServerLogger()
	// migrations
	err := migrations.Migrate()
	if err != nil {
		panic(err)
	}
	// seting up grpc server
	listener, err := net.Listen("tcp", getHost())
	if err != nil {
		log.Fatal(err)
	}
	grpcusers.Connect()
	grpcServer := grpc.NewServer(
		setupGrpcStreamOpts(),
		setupGrpcUnaryOpts(),
	)

	registerServices(grpcServer)

	log.Infof("starting to serve on %s", getHost())
	grpcServer.Serve(listener)
}

func getHost() string {
	host = fmt.Sprintf("%s:%s", viper.GetString("envspotting_apps_host"), viper.GetString("envspotting_apps_port"))
	return host
}

func setupGrpcUnaryOpts() grpc.ServerOption {
	return grpc_middleware.WithUnaryServerChain(
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_logrus.UnaryServerInterceptor(logger.GrpcLogrusEntry, logger.GrpcLogrusOpts...),
	)
}

func setupGrpcStreamOpts() grpc.ServerOption {
	return grpc_middleware.WithStreamServerChain(
		grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_logrus.StreamServerInterceptor(logger.GrpcLogrusEntry, logger.GrpcLogrusOpts...),
	)
}

func registerServices(grpcServer *grpc.Server) {
	applications.Register(grpcServer)
	contours.Register(grpcServer)
	// Disable on prod env
	reflection.Register(grpcServer)
}
