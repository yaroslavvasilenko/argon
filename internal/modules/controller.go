package modules

import (
	bcontroller "github.com/yaroslavvasilenko/argon/internal/modules/boost/controller"
	ccontroller "github.com/yaroslavvasilenko/argon/internal/modules/currency/controller"
	icontroller "github.com/yaroslavvasilenko/argon/internal/modules/image/controller"
	lcontroller "github.com/yaroslavvasilenko/argon/internal/modules/listing/controller"
	loccontroller "github.com/yaroslavvasilenko/argon/internal/modules/location/controller"
)

type Controllers struct {
	Listing  *lcontroller.Listing
	Currency *ccontroller.Currency
	Location *loccontroller.Location
	Boost    *bcontroller.Boost
	Image    *icontroller.Image
}

func NewControllers(services *Services) *Controllers {
	return &Controllers{
		Listing:  lcontroller.NewListing(services.listing),
		Currency: ccontroller.NewCurrency(services.currency),
		Location: loccontroller.NewLocation(services.location),
		Boost:    bcontroller.NewBoost(services.boost),
		Image:    icontroller.NewImage(services.Image),
	}
}
