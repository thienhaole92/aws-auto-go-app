package service

import (
	"github.com/andyle182810/gframework/goredis"
	"github.com/andyle182810/gframework/postgres"
)

type Service struct {
	redisClient    *goredis.Redis
	postgresClient *postgres.Postgres
}

func New(
	redisClient *goredis.Redis,
	postgresClient *postgres.Postgres,
) *Service {
	return &Service{
		redisClient:    redisClient,
		postgresClient: postgresClient,
	}
}
