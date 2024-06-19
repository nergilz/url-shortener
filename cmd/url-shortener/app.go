package urlshortener

import (
	"log/slog"
	"net/http"
	"os"
	"url-shortener/internal/common"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/lid/logger/sl"
	"url-shortener/internal/storage/sqlite"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func AppRun() {
	appCfg := config.MustLoad()

	log := setupLogger(appCfg.Env)
	log.Info("starting url-shortiner", slog.String("env", appCfg.Env))
	log.Debug("debug are enabled")

	appStorage, err := sqlite.New(appCfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	_ = appStorage

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortiner", map[string]string{
			appCfg.Server.User: appCfg.Server.Password,
		}))

		r.Post("/add", save.New(log, appStorage))
	})

	router.Get("/redirect/{alias}", redirect.New(log, appStorage))

	srv := &http.Server{
		Addr:         appCfg.Server.Address,
		Handler:      router,
		ReadTimeout:  appCfg.Server.TimeOut,
		WriteTimeout: appCfg.Server.TimeOut,
		IdleTimeout:  appCfg.Server.IdleTimeout,
	}

	log.Info("start server", slog.String("address", appCfg.Server.Address))

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
