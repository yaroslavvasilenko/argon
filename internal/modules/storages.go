package modules

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaroslavvasilenko/argon/config"
	bstorage "github.com/yaroslavvasilenko/argon/internal/modules/boost/storage"
	cstorage "github.com/yaroslavvasilenko/argon/internal/modules/currency/storage"
	istorage "github.com/yaroslavvasilenko/argon/internal/modules/image/storage"
	lstorage "github.com/yaroslavvasilenko/argon/internal/modules/listing/storage"
	locstorage "github.com/yaroslavvasilenko/argon/internal/modules/location/storage"
	"gorm.io/gorm"
)

type Storages struct {
	Listing         *lstorage.Listing
	Currency        *cstorage.Currency
	CurrencyBinance cstorage.IBinance
	Location        *locstorage.Location
	Boost           *bstorage.Boost
	image           *istorage.Image
}

func NewStorages(cfg config.Config, db *gorm.DB, pool *pgxpool.Pool, minio *istorage.Minio) *Storages {
	boost := bstorage.NewBoost(db, pool)

	return &Storages{
		Listing:         lstorage.NewListing(db, pool, boost),
		Currency:        cstorage.NewCurrency(db, pool),
		CurrencyBinance: cstorage.NewBinance(cfg),
		Location:        locstorage.NewLocation(cfg.Nominatim.BaseUrl),
		Boost:           boost,
		image:           istorage.NewImage(db, pool, minio),
	}
}
