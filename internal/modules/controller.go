package modules

import (

	lcontroller "github.com/yaroslavvasilenko/argon/internal/modules/listing/controller"
	ccontroller "github.com/yaroslavvasilenko/argon/internal/modules/currency/controller"
)

type Controllers struct {
	Listing *lcontroller.Listing
	Currency *ccontroller.Currency
}


func NewControllers(services *Services) *Controllers {
	return &Controllers{
		Listing: lcontroller.NewListing(services.listing),
		Currency: ccontroller.NewCurrency(services.currency),
	}
}