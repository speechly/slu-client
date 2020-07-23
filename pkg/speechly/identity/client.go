package identity

import (
	"context"
	"net/url"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	pgrpc "speechly/slu-client/internal/grpc"
	"speechly/slu-client/pkg/speechly"
)

const (
	name   = "speechly.identity.v1"
	method = "/v1.Identity/Login"
)

// Client is a client for Speechly Identity API.
type Client struct {
	*pgrpc.Client
}

// NewClient returns a new Client configured to use provided URL.
func NewClient(u url.URL) (*Client, error) {
	c, err := pgrpc.NewClient(name, u)
	if err != nil {
		return nil, err
	}

	return &Client{c}, nil
}

// Login calls Identity.Login method with provided identifiers and within provided context.
// It will parse returned token into speechly.AccessToken and return it or any error if it happens.
// nolint: interfacer // linter wants to pass a Stringer instead of UUID, which defeats the purpose of type safety.
func (c *Client) Login(ctx context.Context, appID, deviceID uuid.UUID) (t speechly.AccessToken, err error) {
	res := speechly.LoginResponse{}
	req := speechly.LoginRequest{
		AppId:    appID.String(),
		DeviceId: deviceID.String(),
	}

	if err := c.Invoke(ctx, method, &req, &res, grpc.WaitForReady(true)); err != nil {
		return t, err
	}

	if err := t.Parse(res.GetToken()); err != nil {
		return t, err
	}

	return t, nil
}
