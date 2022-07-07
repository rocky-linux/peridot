package authzed

import (
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/jzelinskie/stringz"
	"google.golang.org/grpc"
)

// Client represents an open connection to Authzed.
//
// Clients are backed by a gRPC client and as such are thread-safe.
type Client struct {
	v1.SchemaServiceClient
	v1.PermissionsServiceClient
}

// NewClient initializes a brand new client for interacting with Authzed.
func NewClient(endpoint string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := grpc.Dial(
		stringz.DefaultEmpty(endpoint, "grpc.authzed.com:443"),
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		v1.NewSchemaServiceClient(conn),
		v1.NewPermissionsServiceClient(conn),
	}, nil
}
