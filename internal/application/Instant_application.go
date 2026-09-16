package application

import (
	"core-server/internal/domain"
	"core-server/internal/infras/clog"
)

type InstantService struct {
	repo domain.UserRepoDomain
	log  *clog.Log
}

func NewInstantService(
	repo domain.UserRepoDomain,
	log *clog.Log,
) *InstantService {
	return &InstantService{
		repo: repo,
		log:  log,
	}
}
