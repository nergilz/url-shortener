package redirect

import (
	"errors"
	"log/slog"
	"net/http"
	"url-shortener/internal/lid/api/response"
	"url-shortener/internal/lid/logger/sl"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type UrlGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter UrlGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")
			render.JSON(w, r, response.RespError("invalid request, alias is empty"))
			return
		}

		redirUrl, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrUrlNotFound) {
			log.Info("url not found", sl.Err(err), "alias:", alias)
			render.JSON(w, r, response.RespError("url not found"))
			return
		}
		if err != nil {
			log.Info("failed to get url", sl.Err(err))
			render.JSON(w, r, response.RespError("failed to get url"))
			return
		}

		log.Info("got url", slog.String("url", redirUrl))

		http.Redirect(w, r, redirUrl, http.StatusFound) // http.StatusFound не кешируется браузером
	}

}
