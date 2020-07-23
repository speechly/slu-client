package slu

import (
	"context"
	"net/url"

	"golang.org/x/text/language"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pgrpc "github.com/speechly/slu-client/internal/grpc"
	"github.com/speechly/slu-client/pkg/logger"
	"github.com/speechly/slu-client/pkg/speechly"
)

// Config is the configuration of an SLU recognition stream.
// It is used by the client to send Config requests when starting new streams.
// For more information, check the Speechly SLU API documentation.
type Config struct {
	NumChannels     int32
	SampleRateHertz int32
	LanguageCode    language.Tag
}

// Client is a client for Speechly SLU API.
type Client struct {
	*pgrpc.Client
	token speechly.AccessToken
	log   logger.Logger
}

// NewClient returns a new Client that will access provided URL with provided access token.
func NewClient(u url.URL, t speechly.AccessToken, log logger.Logger) (*Client, error) {
	c, err := pgrpc.NewClient("speechly.slu.v1", u)
	if err != nil {
		return nil, err
	}

	return &Client{c, t, log}, nil
}

// StreamingRecognise starts a new SLU recognition stream with specified Config.
func (c *Client) StreamingRecognise(ctx context.Context, fmt Config) (RecogniseStream, error) {
	conn, err := c.Conn()
	if err != nil {
		return nil, err
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", "Bearer "+c.token.String())
	str, err := speechly.NewSLUClient(conn).Stream(ctx, grpc.WaitForReady(true))
	if err != nil {
		return nil, err
	}

	return newStream(str, fmt, c.log)
}
