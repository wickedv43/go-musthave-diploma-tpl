package config

import (
	"flag"
	"os"

	"github.com/joho/godotenv"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/logger"
)

type Config struct {
	Server        Server
	AccrualSystem AccrualSystem
	Database      Database

	log *logrus.Entry
}

type Server struct {
	RunAddress string
}

type Database struct {
	DSN string
}

type AccrualSystem struct {
	URL string
}

func NewConfig(i do.Injector) (*Config, error) {
	var cfg Config

	cfg.log = do.MustInvoke[*logger.Logger](i).WithField("component", "config")

	//flags
	flag.StringVar(&cfg.Server.RunAddress, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.Database.DSN, "d", "", "DSN")
	flag.StringVar(&cfg.AccrualSystem.URL, "r", "", "accrual system url")
	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		cfg.log.Warn(err, "loading .env file")
	}

	//env
	runAdress := os.Getenv("RUN_ADDRESS")
	if runAdress != "" {
		cfg.Server.RunAddress = runAdress
	}

	DatabaseDSN := os.Getenv("DATABASE_DSN")
	if DatabaseDSN != "" {
		cfg.Database.DSN = DatabaseDSN
	}

	AccrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS")
	if AccrualSystemAddress != "" {
		cfg.Database.DSN = AccrualSystemAddress
	}

	return &cfg, nil
}
