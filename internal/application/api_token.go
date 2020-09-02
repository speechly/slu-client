package application

import (
	"bufio"
	"context"
	"io"
	"net/url"
	"os"

	"github.com/google/uuid"

	"github.com/speechly/slu-client/pkg/logger"
	"github.com/speechly/slu-client/pkg/speechly"
	"github.com/speechly/slu-client/pkg/speechly/identity"
)

const (
	filePerms = 0600
	fileStart = 0
)

// GetAPIToken fetches Speechly API token from a cache file or refreshes it by calling Speechly Identity API.
func GetAPIToken(
	ctx context.Context, path string, identityURL url.URL, appID, deviceID uuid.UUID, log logger.Logger,
) (speechly.AccessToken, error) {
	var (
		tokenStr string
		token    speechly.AccessToken
	)

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, filePerms) // nolint: gosec
	if err != nil {
		return token, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Warn("Error closing token file", err)
		}
	}()

	s, err := f.Stat()
	if err != nil {
		return token, err
	}

	// If file exists, then it should have a cached token, fetch it.
	if s.Size() != 0 {
		r := bufio.NewReader(f)
		tokenStr, err = r.ReadString('\n')
		if err != nil && err != io.EOF {
			return token, err
		}

		// Check if cached token is still valid.
		if err := token.Parse(tokenStr); err == nil {
			return token, nil
		}
	}

	// Fetch new API token from Identity.
	token, err = identity.GetAccessToken(ctx, identityURL, appID, deviceID, log)
	if err != nil {
		return token, err
	}

	// Save new token to file.
	if err := f.Truncate(fileStart); err != nil {
		return token, err
	}

	if _, err = f.Seek(fileStart, io.SeekStart); err != nil {
		return token, err
	}

	if _, err = f.WriteString(token.String()); err != nil {
		return token, err
	}

	return token, nil
}
