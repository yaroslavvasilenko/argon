package modules

import (
	bcontroller "github.com/yaroslavvasilenko/argon/internal/modules/boost/controller"
	lcontroller "github.com/yaroslavvasilenko/argon/internal/modules/listing/controller"
	ccontroller "github.com/yaroslavvasilenko/argon/internal/modules/currency/controller"
	loccontroller "github.com/yaroslavvasilenko/argon/internal/modules/location/controller"
)

type Controllers struct {
	Listing *lcontroller.Listing
	Currency *ccontroller.Currency
	Location *loccontroller.Location
	Boost *bcontroller.Boost
}


func NewControllers(services *Services) *Controllers {
	return &Controllers{
		Listing: lcontroller.NewListing(services.listing),
		Currency: ccontroller.NewCurrency(services.currency),
		Location: loccontroller.NewLocation(services.location),
		Boost: bcontroller.NewBoost(services.boost),
	}
}