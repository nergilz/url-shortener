package urlshortener

import (
	"log/slog"
	"net/http"
	"os"
	"url-shortener/internal/common"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/lid/logger/sl"
	"url-shortener/internal/storage/sqlite"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func AppRun() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log.Info("starting url-shortiner", slog.String("env", cfg.Env))
	log.Debug("debug are enabled")

	appStorage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	_ = appStorage

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Post("/url", save.New(log, appStorage))

	srv := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      router,
		ReadTimeout:  cfg.Server.TimeOut,
		WriteTimeout: cfg.Server.TimeOut,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	log.Info("start server", slog.String("address", cfg.Server.Address))

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed run server", sl.Err(err))
	}

	log.Info("stop server")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case common.EnvLocal:
		// slog level debug for local
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case common.EnvDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case common.EnvProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log

}
