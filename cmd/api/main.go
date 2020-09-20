package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	_ "github.com/lib/pq"
	"github.com/oklog/run"

	"github.com/risentveber/wallet-api/services/transfers"
)

func retry(pause time.Duration, maxCount uint, call func() error, logger log.Logger) error {
	var err error
	for i := uint(0); i < maxCount; i++ {
		err = call()
		if err == nil {
			return nil
		}
		_ = logger.Log("msg", "retry after: "+pause.String()+", retries left: "+strconv.Itoa(int(maxCount-i)-1)+", error: "+err.Error())
		time.Sleep(pause)
	}

	return err
}

func main() { // nolint funlen
	logger := log.With(log.NewJSONLogger(os.Stdout), "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				_ = logger.Log("fatal", err.Error())
			} else {
				_ = logger.Log("fatal", r)
			}
			os.Exit(1)
		}
	}()
	c := NewConfig()
	_ = logger.Log("config", c.port)

	db, err := sql.Open("postgres", c.dbConnectionURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = retry(c.dbConnectRetryTimout, c.dbConnectRetryCount, db.Ping, logger)
	if err != nil {
		panic(err)
	}

	repo := transfers.NewRepository(db)
	service := transfers.NewService(repo)
	endpoints := transfers.NewEndpoints(service)
	httpHandler := transfers.NewHTTPHandler(endpoints)
	httpServer := &http.Server{Handler: httpHandler}

	_ = logger.Log("msg", "started on port "+c.port)
	var g run.Group
	{
		httpListener, err := net.Listen("tcp", ":"+c.port)
		if err != nil {
			panic(err)
		}

		g.Add(func() error {
			return httpServer.Serve(httpListener)
		}, func(error) {
			_ = httpListener.Close()
		})
	}
	{
		execute, interrupt := run.SignalHandler(context.Background(), syscall.SIGHUP)
		g.Add(execute, interrupt)
	}
	err = g.Run()

	if sig, ok := err.(run.SignalError); ok {
		_ = logger.Log("msg", "exiting by signal "+sig.Signal.String())
		err = nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.shutdownTimeout)
	if err := httpServer.Shutdown(ctx); err != nil {
		cancel()
		_ = logger.Log("msg", "graceful shutdown "+err.Error())
	}
	if err != nil {
		panic(err)
	}
	_ = logger.Log("msg", "server exits normally")
}
