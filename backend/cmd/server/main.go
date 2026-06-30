// Command server is the Cloud Log Access BFF entrypoint.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud-log-access/internal/api"
	"cloud-log-access/internal/api/handler"
	"cloud-log-access/internal/auth"
	"cloud-log-access/internal/config"
	"cloud-log-access/internal/storage"
	awsstore "cloud-log-access/internal/storage/aws"
	azurestore "cloud-log-access/internal/storage/azure"
	gcpstore "cloud-log-access/internal/storage/gcp"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err // logger not configured yet; default slog (stderr) is used
	}

	logger := newLogger(cfg)
	slog.SetDefault(logger)
	logger.Info("starting cloud-log-access bff", "config", cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	users, err := auth.SeedUsers()
	if err != nil {
		return fmt.Errorf("seed users: %w", err)
	}
	store := auth.NewMemoryStore(users)
	issuer := auth.NewIssuer(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience, cfg.JWTExpiry)
	denylist := auth.NewDenylist()
	denylist.StartCleanup(ctx, 5*time.Minute)
	authHandler := handler.NewAuth(store, issuer, denylist, cfg.JWTExpiry, logger)

	storageSvc := storage.NewService(cfg.PresignMinTTL, cfg.PresignMaxTTL)
	s3Adapter := awsstore.New(awsstore.Config{
		Region:           cfg.AWSRegion,
		InternalEndpoint: cfg.S3InternalEndpoint,
		PublicEndpoint:   cfg.S3PublicEndpoint,
		AccessKeyID:      cfg.AWSAccessKeyID,
		SecretAccessKey:  cfg.AWSSecretAccessKey,
		UsePathStyle:     cfg.S3UsePathStyle,
	})
	storageSvc.Register("aws", "AWS S3", cfg.AWSRegion, cfg.S3Buckets, s3Adapter)

	if cfg.GCSInternalEndpoint != "" {
		gcsAdapter, gerr := gcpstore.New(ctx, gcpstore.Config{
			InternalEndpoint: cfg.GCSInternalEndpoint,
			PublicEndpoint:   cfg.GCSPublicEndpoint,
		})
		if gerr != nil {
			return fmt.Errorf("gcs adapter: %w", gerr)
		}
		storageSvc.Register("gcp", "GCP Storage", "", cfg.GCSBuckets, gcsAdapter)
	}
	if cfg.AzureInternalEndpoint != "" {
		azAdapter, aerr := azurestore.New(azurestore.Config{
			AccountName:      cfg.AzureAccount,
			AccountKey:       cfg.AzureKey,
			InternalEndpoint: cfg.AzureInternalEndpoint,
			PublicEndpoint:   cfg.AzurePublicEndpoint,
		})
		if aerr != nil {
			return fmt.Errorf("azure adapter: %w", aerr)
		}
		storageSvc.Register("azure", "Azure Blob", "", cfg.AzureContainers, azAdapter)
	}
	storageHandler := handler.NewStorage(storageSvc, logger)

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.Port),
		Handler: api.NewRouter(api.RouterDeps{
			Logger:    logger,
			Config:    cfg,
			Auth:      authHandler,
			Storage:   storageHandler,
			Readiness: storageSvc,
			Validator: issuer,
			Denylist:  denylist,
		}),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		// WriteTimeout is intentionally 0: downloads stream, so handlers are
		// bounded by per-route context deadlines instead.
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		logger.Info("shutdown signal received, draining connections")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}
		logger.Info("shutdown complete")
	}
	return nil
}

func newLogger(cfg *config.Config) *slog.Logger {
	level := slog.LevelDebug
	if cfg.IsProd() {
		level = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: level}

	var h slog.Handler
	if cfg.IsProd() {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(h)
}
