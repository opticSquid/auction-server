package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"sync/atomic"

	"github.com/go-chi/chi/v5"
	"github.com/knadh/koanf/v2"
	"github.com/opticSquid/auction-server/configuration"
	"go.uber.org/zap"
)

var k = koanf.New(".")
var isReady atomic.Bool

func main() {
	log.Println("starting auction server app configuration")
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("error in starting zap logger. exiting application")
	}
	defer logger.Sync()

	logger.Info("zap logger started")
	configuration.LoadConfig(k, logger)
	logger = logger.With(zap.String("application-name", k.String("application.name")))

	logger.Info("application configuration complete")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	basePath := k.String("server.basepath")
	r := chi.NewRouter()
	r.Use(panicRecovery)
	r.Route("/"+basePath, func(r chi.Router) {
		// health check
		r.Group(func(r chi.Router) {
			r.Get("/live", livenessHandler)
			r.Get("/ready", readinessHandler)
		})

		r.Group(func(r chi.Router) {
			r.Use(readinessMiddleware)
			// future: /bid
		})
	})

	srv := &http.Server{
		Addr:              ":" + k.String("server.port"),
		Handler:           r,
		ReadTimeout:       2 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		WriteTimeout:      2 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    8 * 1024,
	}

	go func() {
		<-ctx.Done()
		logger.Info("starting graceful shutdown")
		isReady.Store(false)
		logger.Debug("starting shutdown sequence for http server")
		// giving LB time to drain connections
		time.Sleep(2 * time.Second)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("failed to shutdown http server", zap.Error(err))
		}
		logger.Debug("http server shutdown sequence complete")
		logger.Info("application stopped")
	}()

	isReady.Store(true)
	logger.Info("application started, http server live", zap.Int("port", k.Int("server.port")))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("failed to start http server", zap.Error(err))
	}
}

var ok = []byte("ok")

func livenessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write(ok)
}

func handleReady(w http.ResponseWriter) bool {
	if !isReady.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		return false
	}
	return true
}
func readinessHandler(w http.ResponseWriter, r *http.Request) {
	if !handleReady(w) {
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(ok)
}

func readinessMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !handleReady(w) {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func panicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered:  %v", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
