package metadata

import (
	"context"
	"errors"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"google.golang.org/grpc/codes"
)

const (
	errAuthTokenNotProvided = "authorization token not provided via grpc metadata"
)

const (
	internalToken = "x-internal-token"
)

// Get auth token from metadata
func GetAuthToken(ctx context.Context, header string) (token string, code codes.Code, err error){
	tknStr := metautils.ExtractIncoming(ctx).Get("authorization")
	if tknStr == "" {
		return "", codes.PermissionDenied, errors.New(errAuthTokenNotProvided)
	}
	return tknStr, codes.OK, nil
} 
// Set authtoken to metadata

// Get interanl token from metadata
// Set internal token to metadata

func MetadataInternalProxy(ctx context.Context) context.Context{
	md := metautils.ExtractIncoming(ctx)
	md.Set(internalToken, "temp")
	return md.ToOutgoing(ctx)
}