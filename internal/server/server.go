package server

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/config"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/logger"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/storage"
)

type Server struct {
	echo    *echo.Echo
	cfg     *config.Config
	storage storage.DataKeeper
	logger  *logrus.Entry
}

func NewServer(i do.Injector) (*Server, error) {
	s, err := do.InvokeStruct[Server](i)
	if err != nil {
		return nil, errors.Wrap(err, "invoke struct error")
	}

	//init
	s.echo = echo.New()
	s.cfg = do.MustInvoke[*config.Config](i)
	s.logger = do.MustInvoke[*logger.Logger](i).WithField("component", "server")

	//middleware
	s.echo.Use(middleware.Recover())

	//routes

	return s, nil
}

func (s *Server) Start() {
	s.logger.Info("server started...")
	err := s.echo.Start(s.cfg.Server.RunAddress)
	if err != nil {
		s.logger.Fatal(errors.Wrap(err, "start server"))
	}
}
