package modules

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaroslavvasilenko/argon/internal/core/logger"
	bservice "github.com/yaroslavvasilenko/argon/internal/modules/boost/service"
	cservice "github.com/yaroslavvasilenko/argon/internal/modules/currency/service"
	iservice "github.com/yaroslavvasilenko/argon/internal/modules/image/service"
	lservice "github.com/yaroslavvasilenko/argon/internal/modules/listing/service"
	locservice "github.com/yaroslavvasilenko/argon/internal/modules/location/service"
)

type Services struct {
	listing  *lservice.Listing
	currency *cservice.Currency
	location *locservice.Location
	boost    *bservice.Boost
	image    *iservice.Image
}

func NewServices(storages *Storages, pool *pgxpool.Pool, lg *logger.Glog) *Services {
	locationService := locservice.NewLocation(storages.Location, lg)

	return &Services{
		listing:  lservice.NewListing(storages.Listing, pool, lg, locationService),
		currency: cservice.NewCurrency(storages.Currency, storages.CurrencyBinance, lg),
		location: locationService,
		boost:    bservice.NewBoost(storages.Boost, lg),
		image:    iservice.NewImage(storages.image, lg),
	}
}
