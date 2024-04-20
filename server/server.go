package server

import (
	"context"
	"errors"
	"fmt"
	l "log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

// WebLogger - logs all incoming requests
func WebLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		l.Info("request",
			l.String("method", r.Method),
			l.String("requestURI", r.RequestURI),
			l.String("requesterIP", r.RemoteAddr),
			l.Duration("requestProcessingTime", time.Since(start)))
	})
}

// Start - starts http-server and starts processing incoming requests
//
// Requires `TACS_PORT` in environment variables.
// `TACS_ADDR`, `TACS_CERT`, `TACS_CERT_KEY` can also be specified.
func Start() {
	addr := os.Getenv("TACS_ADDR")
	port := os.Getenv("TACS_PORT")
	cert := os.Getenv("TACS_CERT")
	key := os.Getenv("TACS_CERT_KEY")

	l.Debug("server params",
		l.String("TACS_ADDR", addr),
		l.String("TACS_PORT", port))

	// Create router and declare endpoints
	r := http.NewServeMux()
	r.HandleFunc("/favicon.ico", http.NotFound)
	r.HandleFunc("/{username}", Analysis)

	// Setting server parameters
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%s", addr, port),
		Handler:           WebLogger(r),
		ReadTimeout:       time.Second * 5,
		ReadHeaderTimeout: time.Second * 5,
		WriteTimeout:      time.Second * 10,
		IdleTimeout:       time.Second * 5,
	}

	// Selecting a server startup mode
	if key != "" && cert != "" {
		// Startup using SSL
		go func() {
			if err := srv.ListenAndServeTLS(
				filepath.Clean(cert), filepath.Clean(key)); !errors.
				Is(err, http.ErrServerClosed) {
				l.Error("server error",
					l.String("TACS_ADDR", addr),
					l.String("TACS_PORT", port),
					l.String("TACS_CERT", cert),
					l.String("TACS_CERT_KEY", key),
					l.Any("err", err))
			}
		}()
	} else {
		// Startup without SSL
		go func() {
			if err := srv.ListenAndServe(); !errors.
				Is(err, http.ErrServerClosed) {
				l.Error("server error",
					l.String("TACS_ADDR", addr),
					l.String("TACS_PORT", port),
					l.String("TACS_CERT", cert),
					l.String("TACS_CERT_KEY", key),
					l.Any("err", err))
			}
		}()
	}

	// Creating context for smooth server shutdown
	ctx, ctxCancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Locking the main goroutine until signaled to terminate (CTRL+Z)
	<-c
	l.Info("Background processes are terminated. Please wait")

	// Sends a signal to the server to shutdown.
	// It closes registration of new connections,
	// finishes processing old ones and stops
	if err := srv.Shutdown(ctx); err != nil {
		l.Error("server shutdown error", l.Any("err", err))
	}
	ctxCancel()

	l.Info("the server shutdown")
}
