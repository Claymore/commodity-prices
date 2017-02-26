package price

import (
	"time"

	"github.com/shopspring/decimal"
)

type Price struct {
	Commodity string
	Date      time.Time
	Price     decimal.Decimal
}

type Teller interface {
	Prices(commodity, from, till string) (prices []Price, err error)
}
