// nolint gomnd
package main

import (
	"flag"
	"time"

	"github.com/go-kit/kit/log/level"
)

type Config struct {
	port                 string
	logLevel             level.Option
	dbConnectionURL      string
	shutdownTimeout      time.Duration
	dbConnectRetryCount  uint
	dbConnectRetryTimout time.Duration
}

func NewConfig() Config {
	var c Config
	flag.StringVar(&c.port, "port", "8080", "port")
	flag.StringVar(&c.dbConnectionURL, "db", "", "db connections credentials")
	flag.DurationVar(&c.shutdownTimeout, "shutdownTimeout", 10*time.Second, "graceful shutdown timeout")
	flag.UintVar(&c.dbConnectRetryCount, "dbRetryCount", 10, "retry count for connecting to db")
	flag.DurationVar(&c.dbConnectRetryTimout, "dbRetryTimeout", 2*time.Second, "retry timeout for connecting to db")
	logLevel := flag.String("logLevel", "info", "debug|info|warn|error")
	flag.Parse()
	switch *logLevel {
	case "error":
		c.logLevel = level.AllowError()
	case "debug":
		c.logLevel = level.AllowDebug()
	case "warn":
		c.logLevel = level.AllowWarn()
	default:
		c.logLevel = level.AllowInfo()
	}
	return c
}
