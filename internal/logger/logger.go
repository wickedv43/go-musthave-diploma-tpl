package logger

import (
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

func NewLogger(_ do.Injector) (*Logger, error) {

	log := logrus.New()
	log.SetLevel(logrus.InfoLevel)

	return &Logger{
		log,
	}, nil
}
