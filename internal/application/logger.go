package application

import (
	"io"

	"github.com/sirupsen/logrus"

	"speechly/slu-client/pkg/logger"
)

// NewLogger returns a new logger that writes to out with specified level.
func NewLogger(out io.Writer, level logrus.Level) logger.Logger {
	l := logrus.New()

	l.SetFormatter(&logrus.TextFormatter{})
	l.SetOutput(out)
	l.SetLevel(level)

	return l
}
