package modules

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/gorm"
	"github.com/yaroslavvasilenko/argon/config"
	cstorage "github.com/yaroslavvasilenko/argon/internal/modules/currency/storage"
	lstorage "github.com/yaroslavvasilenko/argon/internal/modules/listing/storage"
	locstorage "github.com/yaroslavvasilenko/argon/internal/modules/location/storage"
	bstorage "github.com/yaroslavvasilenko/argon/internal/modules/boost/storage"
)

type Storages struct {
	listing *lstorage.Listing
	currency *cstorage.Currency
	currencyBinance cstorage.IBinance
	location *locstorage.Location
	boost *bstorage.Boost
}

func NewStorages(cfg config.Config, db *gorm.DB, pool *pgxpool.Pool) *Storages {
	boost := bstorage.NewBoost(db, pool)

	return &Storages{
		listing: lstorage.NewListing(db, pool, boost),
		currency: cstorage.NewCurrency(db, pool),
		currencyBinance: cstorage.NewBinance(cfg),
		location: locstorage.NewLocation(cfg.Nominatim.BaseUrl),
		boost: boost,
	}
}