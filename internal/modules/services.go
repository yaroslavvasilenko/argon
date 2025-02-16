package modules

import (
	"github.com/jackc/pgx/v5/pgxpool"
	lservice "github.com/yaroslavvasilenko/argon/internal/modules/listing/service"
	cservice "github.com/yaroslavvasilenko/argon/internal/modules/currency/service"
	"github.com/yaroslavvasilenko/argon/internal/core/logger"
)

type Services struct {
	listing *lservice.Listing
	currency *cservice.Currency
}


func NewServices(storages *Storages, pool *pgxpool.Pool, lg *logger.LogPhuslu) *Services {
	return &Services{
		listing: lservice.NewListing(storages.listing, pool, lg),
		currency: cservice.NewCurrency(storages.currency, storages.currencyBinance, lg),
	}
}