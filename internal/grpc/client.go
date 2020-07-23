package grpc

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"net/url"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/resolver"
)

// ErrInvalidScheme is returned when provided URL has an unsupported scheme.
// Supported schemes are "grpc" or "grpc+tls".
var ErrInvalidScheme = errors.New("unsupported URL scheme")

// Client is a wrapper around gRPC connection that implements StarterStopper and Healthcheck interfaces.
type Client struct {
	conn   *grpc.ClientConn
	name   string
	host   string
	secure bool
}

// NewClient returns a new instance of GRPCClient.
func NewClient(name string, u url.URL) (*Client, error) {
	var secure bool
	switch u.Scheme {
	case "grpc+tls":
		secure = true
	case "grpc":
		secure = false
	default:
		return nil, ErrInvalidScheme
	}

	return &Client{
		name:   name,
		host:   u.Host,
		secure: secure,
	}, nil
}

// Invoke invokes a unary gRPC call through the underlying connection.
func (c *Client) Invoke(ctx context.Context, method string, args, res interface{}, opts ...grpc.CallOption) error {
	conn, err := c.Conn()
	if err != nil {
		return err
	}

	return conn.Invoke(ctx, method, args, res, opts...)
}

// Conn checks the underlying gRPC client connection for readiness and returns it.
// If connection is not established or is not ready, an error is returned instead.
func (c *Client) Conn() (*grpc.ClientConn, error) {
	if c.conn == nil || c.conn.GetState() != connectivity.Ready {
		return nil, fmt.Errorf("gRPC client %s is not connected", c.name)
	}

	return c.conn, nil
}

// Dial starts the client by establishing gRPC connection.
func (c *Client) Dial(ctx context.Context) error {
	var tlsOpt grpc.DialOption

	if c.secure {
		cp, err := x509.SystemCertPool()
		if err != nil {
			return fmt.Errorf("failed to start gRPC client %s: %w", c.name, err)
		}

		tlsOpt = grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(cp, ""))
	} else {
		tlsOpt = grpc.WithInsecure()
	}

	resolver.SetDefaultScheme("dns")

	conn, err := grpc.DialContext(ctx, c.host, grpc.WithBlock(), tlsOpt)
	if err != nil {
		return fmt.Errorf("failed to start gRPC client %s: %w", c.name, err)
	}

	c.conn = conn

	return nil
}

// Close closes the client by closing gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}
