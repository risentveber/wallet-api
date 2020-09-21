package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	_ "github.com/lib/pq"
	"github.com/oklog/run"

	"github.com/risentveber/wallet-api/integration"
	"github.com/risentveber/wallet-api/services/transfers"
)

// give stack when panic is recovered.
func trimPanicStack() string {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return string(buf[:n])
		}
		buf = make([]byte, 2*len(buf)) // nolint gomnd
	}
}

func RecoverWrap(h http.Handler, logger log.Logger) http.Handler {
	debugLogger := level.Debug(logger)
	errorLogger := level.Error(logger)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(begin time.Time) {
			_ = debugLogger.Log("path", r.URL.String(), "latency", time.Since(begin))
		}(time.Now())
		defer func() {
			r := recover()
			if r != nil {
				msg := ""
				switch t := r.(type) {
				case string:
					msg = t
				case error:
					msg = t.Error()
				default:
					msg = "unknown error"
				}
				_ = errorLogger.Log("panic", msg, "stack", trimPanicStack())
				http.Error(w, msg, http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}

func main() { // nolint funlen
	c := NewConfig()
	logger := level.NewFilter(log.NewJSONLogger(os.Stdout), c.logLevel)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				_ = level.Error(logger).Log("fatal", err.Error())
			} else {
				_ = level.Error(logger).Log("fatal", r)
			}
			os.Exit(1)
		}
	}()

	db, err := sql.Open("postgres", c.dbConnectionURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = integration.Retry(c.dbConnectRetryTimout, c.dbConnectRetryCount, db.Ping, logger)
	if err != nil {
		panic(err)
	}

	repo := transfers.NewRepository(db)
	service := transfers.NewService(repo)
	endpoints := transfers.NewEndpoints(service)
	httpHandler := transfers.NewHTTPHandler(endpoints, logger)
	httpServer := &http.Server{Handler: RecoverWrap(httpHandler, logger)}

	_ = level.Info(logger).Log("msg", "started on port "+c.port)
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
		_ = level.Info(logger).Log("msg", "exiting by signal "+sig.Signal.String())
		err = nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.shutdownTimeout)
	if err := httpServer.Shutdown(ctx); err != nil {
		cancel()
		_ = level.Info(logger).Log("msg", "graceful shutdown "+err.Error())
	}
	if err != nil {
		panic(err)
	}
	_ = level.Info(logger).Log("msg", "server exits normally")
}
