package grpcusers

import (
	"fmt"

	"github.com/badhouseplants/envspotting-apps/tools/logger"
	"github.com/badhouseplants/envspotting-go-proto/models/users/accounts"
	"github.com/badhouseplants/envspotting-go-proto/models/users/authorization"
	"github.com/badhouseplants/envspotting-go-proto/models/users/rights"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	defaultName = "world"
)

// Clients
var (
	AccountClient accounts.AccountsClient
	RightsClient rights.RightsClient
	AuthorizationClient authorization.AuthorizationClient
)

// var Client *
func Connect() {
	log := logger.GetServerLogger()
	// Set up a connection to the server.
	grpcopt := grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time: 20,
		Timeout: 10,
		PermitWithoutStream: true,
	})
	conn, err := grpc.Dial(getHost(), grpc.WithInsecure(), grpc.WithBlock(), grpcopt)
	if err != nil {
		log.Error("did not connect: %v", err)
	}
	AccountClient = accounts.NewAccountsClient(conn)
	RightsClient = rights.NewRightsClient(conn)
	AuthorizationClient = authorization.NewAuthorizationClient(conn)
}

func getHost() string {
	host := fmt.Sprintf("%s:%s", viper.GetString("envspotting_users_host"), viper.GetString("envspotting_users_port"))
	return host
}


