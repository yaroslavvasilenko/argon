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
	// Сначала создаем сервис локаций, так как он нужен для сервиса листинга
	locationService := locservice.NewLocation(storages.Location, lg)

	return &Services{
		listing: lservice.NewListing(storages.Listing, pool, lg, locationService),
		currency: cservice.NewCurrency(storages.Currency, storages.CurrencyBinance, lg),
		location: locationService,
		boost: bservice.NewBoost(storages.Boost, lg),
	}
}