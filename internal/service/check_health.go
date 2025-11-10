package service

import (
	"net/http"
	"time"

	"github.com/andyle182810/gframework/goredis"
	"github.com/andyle182810/gframework/httpserver"
	"github.com/andyle182810/gframework/notifylog"
	"github.com/andyle182810/gframework/postgres"
	"github.com/labstack/echo/v4"
)

type CheckHealthRequest struct{}

type CheckHealthResponse struct {
	Status   string `example:"healthy" json:"status"`
	Postgres string `example:"ok"      json:"postgres"`
	Redis    string `example:"ok"      json:"redis"`
}

func (s *Service) CheckHealth(ctx echo.Context, req *CheckHealthRequest) (any, *echo.HTTPError) {
	delegator := func(
		log notifylog.NotifyLog,
		ctx echo.Context,
		req *CheckHealthRequest,
	) (*httpserver.HandlerResponse[CheckHealthResponse], *echo.HTTPError) {
		handler := NewHealthHandler(
			log,
			s.postgresClient,
			s.redisClient,
		)

		return handler.Handle(ctx, req)
	}

	return httpserver.ExecuteStandardized(ctx, req, "CheckHealth", delegator)
}

type HealthHandler struct {
	log            notifylog.NotifyLog
	postgresClient *postgres.Postgres
	redisClient    *goredis.Redis
}

func NewHealthHandler(
	log notifylog.NotifyLog,
	postgresClient *postgres.Postgres,
	redisClient *goredis.Redis,
) *HealthHandler {
	return &HealthHandler{
		log:            log,
		postgresClient: postgresClient,
		redisClient:    redisClient,
	}
}

func (h *HealthHandler) Handle(
	ctx echo.Context,
	_ *CheckHealthRequest,
) (*httpserver.HandlerResponse[CheckHealthResponse], *echo.HTTPError) {
	start := time.Now()
	reqCtx := ctx.Request().Context()

	h.log.Info().Msg("Starting health check")

	if _, err := h.redisClient.Ping(reqCtx).Result(); err != nil {
		h.log.Error().
			Err(err).
			Dur("elapsed", time.Since(start)).
			Msg("Redis health check failed")

		return nil, echo.NewHTTPError(http.StatusServiceUnavailable, "redis connection failed")
	}

	h.log.Debug().
		Dur("elapsed", time.Since(start)).
		Msg("Redis connection healthy")

	if err := h.postgresClient.Ping(reqCtx); err != nil {
		h.log.Error().
			Err(err).
			Dur("elapsed", time.Since(start)).
			Msg("Postgres health check failed")

		return nil, echo.NewHTTPError(http.StatusServiceUnavailable, "postgres connection failed")
	}

	h.log.Debug().
		Dur("elapsed", time.Since(start)).
		Msg("Postgres connection healthy")

	duration := time.Since(start)

	h.log.Info().
		Str("status", "healthy").
		Dur("duration", duration).
		Msg("Health check completed successfully")

	return &httpserver.HandlerResponse[CheckHealthResponse]{
		Data: CheckHealthResponse{
			Status:   "healthy",
			Postgres: "ok",
			Redis:    "ok",
		},
		Pagination: nil,
	}, nil
}
