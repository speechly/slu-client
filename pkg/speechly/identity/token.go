package identity

import (
	"context"
	"net/url"

	"github.com/google/uuid"

	"speechly/slu-client/pkg/logger"
	"speechly/slu-client/pkg/speechly"
)

// GetAccessToken is a convenience wrapper that instantiates a new identity client and calls Login on it.
func GetAccessToken(
	ctx context.Context, u url.URL, appID, deviceID uuid.UUID, log logger.Logger,
) (t speechly.AccessToken, err error) {
	cli, err := NewClient(u)
	if err != nil {
		return t, err
	}

	if err := cli.Dial(ctx); err != nil {
		return t, err
	}

	defer func() {
		if err := cli.Close(); err != nil {
			log.Warn("Error stopping Speechly Identity API client", err)
		}
	}()

	return cli.Login(ctx, appID, deviceID)
}
