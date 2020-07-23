package application

import (
	"net/url"

	"github.com/google/uuid"
	"golang.org/x/text/language"
)

// Config is the configuration of the CLI app.
type Config struct {
	SluURL       url.URL
	IdentityURL  url.URL
	AppID        uuid.UUID
	DeviceID     uuid.UUID
	LanguageCode language.Tag
	isValid      bool
}

// Parse parses the config from provided string values.
func (c *Config) Parse(sluURL, identityURL, appID, deviceID, languageCode string) error {
	s, err := url.Parse(sluURL)
	if err != nil {
		return err
	}

	i, err := url.Parse(identityURL)
	if err != nil {
		return err
	}

	a, err := uuid.Parse(appID)
	if err != nil {
		return err
	}

	d, err := uuid.Parse(deviceID)
	if err != nil {
		return err
	}

	lang, err := language.Parse(languageCode)
	if err != nil {
		return err
	}

	c.IdentityURL = *i
	c.SluURL = *s
	c.AppID = a
	c.DeviceID = d
	c.LanguageCode = lang
	c.isValid = true

	return nil
}

// IsValid returns true if the config is valid and false otherwise.
func (c *Config) IsValid() bool {
	return c.isValid
}
