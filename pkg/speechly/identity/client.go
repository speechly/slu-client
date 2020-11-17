package identity

import (
	"context"
	"net/url"

	"github.com/google/uuid"
	identityv1 "github.com/speechly/api/go/speechly/identity/v1"
	"google.golang.org/grpc"

	pgrpc "github.com/speechly/slu-client/internal/grpc"
	"github.com/speechly/slu-client/pkg/speechly"
)

// Client is a client for Speechly Identity API.
type Client struct {
	*pgrpc.Client
}

// NewClient returns a new Client configured to use provided URL.
func NewClient(u url.URL) (*Client, error) {
	c, err := pgrpc.NewClient("speechly.identity.v1", u)
	if err != nil {
		return nil, err
	}

	return &Client{c}, nil
}

// Login calls Identity.Login method with provided identifiers and within provided context.
// It will parse returned token into speechly.AccessToken and return it or any error if it happens.
// nolint: interfacer // linter wants to pass a Stringer instead of UUID, which defeats the purpose of type safety.
func (c *Client) Login(ctx context.Context, appID, deviceID uuid.UUID) (t speechly.AccessToken, err error) {
	conn, err := c.Conn()
	if err != nil {
		return t, err
	}

	req := identityv1.LoginRequest{
		AppId:    appID.String(),
		DeviceId: deviceID.String(),
	}

	res, err := identityv1.NewIdentityClient(conn).Login(ctx, &req, grpc.WaitForReady(true))
	if err != nil {
		return t, err
	}

	if err := t.Parse(res.GetToken()); err != nil {
		return t, err
	}

	return t, nil
}
