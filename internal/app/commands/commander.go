package commands

import (
	"database/sql"
	"git.codenrock.com/avito-testirovanie-na-backend-1270/cnrprod1725726738-team-78269/zadanie-6105/internal/service/bid"
	"git.codenrock.com/avito-testirovanie-na-backend-1270/cnrprod1725726738-team-78269/zadanie-6105/internal/service/tender"
)

type Commander struct {
	db            *sql.DB
	tenderService *tender.Service
	bidService    *bid.Service
}

func NewCommander(db *sql.DB, tenderService *tender.Service, bidService *bid.Service) *Commander {
	return &Commander{
		db:            db,
		tenderService: tenderService,
		bidService:    bidService,
	}
}
