package xmlrpc

import (
	"fmt"
	"net/http"
	"net/rpc"
	"net/url"
)

// Client is responsible for making calls to RPC services with help of underlying rpc.Client.
type Client struct {
	*rpc.Client
	codec *Codec
}

// NewClient creates a Client with http.DefaultClient.
// If provided endpoint is not valid, an error is returned.
func NewClient(endpoint string, opts ...Option) (*Client, error) {

	// Parse Endpoint URL
	endpointUrl, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint url: %w", err)
	}

	codec := NewCodec(endpointUrl, http.DefaultClient)

	c := &Client{
		codec:  codec,
		Client: rpc.NewClientWithCodec(codec),
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// NewCustomClient allows customization of http.Client used to make RPC calls.
// If provided endpoint is not valid, an error is returned.
// Deprecated: prefer using NewClient with HttpClient Option
func NewCustomClient(endpoint string, httpClient *http.Client) (*Client, error) {

	return NewClient(endpoint, HttpClient(httpClient))
}
