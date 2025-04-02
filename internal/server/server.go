package server

import (
	"context"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
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
	client  *resty.Client
	cfg     *config.Config
	storage storage.DataKeeper
	logger  *logrus.Entry

	pauseMu    sync.Mutex
	pauseUntil time.Time

	rootCtx   context.Context
	cancelCtx func()
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
	//ctx
	s.rootCtx, s.cancelCtx = context.WithCancel(context.Background())

	s.storage = do.MustInvoke[*storage.PostgresStorage](i)

	//middleware
	s.echo.Use(middleware.Recover(), middleware.Gzip(), s.logHandler, s.CORSMiddleware)

	//free routes
	s.echo.POST(`/api/user/register`, s.onRegister)
	s.echo.POST(`/api/user/login`, s.onLogin)

	//authorized users
	user := s.echo.Group(``, s.authMiddleware)
	user.POST(`/api/user/orders`, s.onPostOrders)
	user.GET(`/api/user/orders`, s.onGetOrders)
	user.GET(`/api/user/balance`, s.onGetUserBalance)
	user.POST(`/api/user/balance/withdraw`, s.onWithDraw)
	user.GET(`/api/user/withdrawals`, s.GetUserBills)

	//accrual client
	s.client = resty.New()

	return s, nil
}

func (s *Server) Start() {
	s.logger.Info("server started...")
	go s.echo.Start(s.cfg.Server.RunAddress)

	//watch orders
	s.watch(s.rootCtx)
}
