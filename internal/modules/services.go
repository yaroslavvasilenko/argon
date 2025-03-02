package modules

import (
	"github.com/jackc/pgx/v5/pgxpool"
	lservice "github.com/yaroslavvasilenko/argon/internal/modules/listing/service"
	cservice "github.com/yaroslavvasilenko/argon/internal/modules/currency/service"
	locservice "github.com/yaroslavvasilenko/argon/internal/modules/location/service"
	bservice "github.com/yaroslavvasilenko/argon/internal/modules/boost/service"
	"github.com/yaroslavvasilenko/argon/internal/core/logger"
)

type Services struct {
	listing *lservice.Listing
	currency *cservice.Currency
	location *locservice.Location
	boost *bservice.Boost
}


func NewServices(storages *Storages, pool *pgxpool.Pool, lg *logger.Glog) *Services {
	return &Services{
		listing: lservice.NewListing(storages.listing, pool, lg),
		currency: cservice.NewCurrency(storages.currency, storages.currencyBinance, lg),
		location: locservice.NewLocation(storages.location, lg),
		boost: bservice.NewBoost(storages.boost, lg),
	}
}