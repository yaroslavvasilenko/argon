package modules

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/gorm"
	"github.com/yaroslavvasilenko/argon/config"
	cstorage "github.com/yaroslavvasilenko/argon/internal/modules/currency/storage"
	lstorage "github.com/yaroslavvasilenko/argon/internal/modules/listing/storage"
	locstorage "github.com/yaroslavvasilenko/argon/internal/modules/location/storage"
)

type Storages struct {
	listing *lstorage.Listing
	currency *cstorage.Currency
	currencyBinance cstorage.IBinance
	location *locstorage.Location
}

func NewStorages(cfg config.Config, db *gorm.DB, pool *pgxpool.Pool) *Storages {
	return &Storages{
		listing: lstorage.NewListing(db, pool),
		currency: cstorage.NewCurrency(db, pool),
		currencyBinance: cstorage.NewBinance(cfg),
		location: locstorage.NewLocation(cfg.Nominatim.BaseUrl),
	}
}