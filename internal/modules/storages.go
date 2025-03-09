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
	Listing *lstorage.Listing
	Currency *cstorage.Currency
	CurrencyBinance cstorage.IBinance
	Location *locstorage.Location
	Boost *bstorage.Boost
}

func NewStorages(cfg config.Config, db *gorm.DB, pool *pgxpool.Pool) *Storages {
	boost := bstorage.NewBoost(db, pool)

	return &Storages{
		Listing: lstorage.NewListing(db, pool, boost),
		Currency: cstorage.NewCurrency(db, pool),
		CurrencyBinance: cstorage.NewBinance(cfg),
		Location: locstorage.NewLocation(cfg.Nominatim.BaseUrl),
		Boost: boost,
	}
}