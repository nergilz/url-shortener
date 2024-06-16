package save

import (
	"log/slog"
	"net/http"
	"url-shortener/internal/common"
	resp "url-shortener/internal/lid/api/response"
	"url-shortener/internal/lid/logger/sl"
	"url-shortener/internal/lid/random"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Url   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

type URLServer interface {
	SaveURL(urlToSave, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			// возвращаем json с ошибкой клиенту
			render.JSON(w, r, resp.RespError("failed to decode request body"))
			// render не прекращает работу функции
			// останавливаем т.к. обработка запроса продолжается
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validErr := err.(validator.ValidationErrors)
			// логируем всю ошибку
			log.Error("invalid request", sl.Err(err))
			render.JSON(w, r, resp.RespError("invalid request"))
			// возвращаем читаемую ошибку
			render.JSON(w, r, resp.ValidationError(validErr))
			return
		}

		alias := req.Alias

		if alias == "" {
			alias = random.NewRandomString(common.AliasLength)
		}

		id, err := urlSaver.SaveURL(req.Url, req.Alias)
		if err != nil {
			log.Error("save error", sl.Err(err))
			render.JSON(w, r, resp.RespError("can not saved url"))
			return
		}

		log.Info("url added", slog.Int64("id", id))

		render.JSON(w, r, Response{
			Response: resp.RespOK(),
			Alias:    alias,
		})
	}
}
