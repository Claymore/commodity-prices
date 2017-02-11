package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

type Price struct {
	Commodity string
	Date      time.Time
	Price     float64
}

func (p Price) LedgerFormat() string {
	date := p.Date.Format("2006/01/02")
	return fmt.Sprintf("P %s %s %f RUR", date, p.Commodity, p.Price)
}

func (p Price) CSVFormat() string {
	date := p.Date.Format("01.02.2006")
	return fmt.Sprintf("%s,%s,%f", p.Commodity, date, p.Price)
}

func main() {
	var commodity = flag.String("commodity", "USD", "commodity")
	var from = flag.String("from", "", "from")
	var till = flag.String("till", "", "till")
	var format = flag.String("format", "csv", "format output: ledger|csv")
	var source = flag.String("source", "cbr", "source: cbr|moex")
	var market = flag.String("market", "shares", "MOEX market: shares|index")
	flag.Parse()

	var err error
	var prices []Price
	switch *source {
	case "cbr":
		cbr := CBR{}
		prices, err = cbr.Prices(*commodity, *from, *till)
		break
	case "moex":
		moex := MOEX{Market: *market}
		prices, err = moex.Prices(*commodity, *from, *till)
		break
	default:
		err = fmt.Errorf("unknown source: %s", *source)
		break
	}
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(-1)
	}
	switch *format {
	case "ledger":
		for _, p := range prices {
			fmt.Printf("%s\n", p.LedgerFormat())
		}
		break
	case "csv":
		for _, p := range prices {
			fmt.Printf("%s\n", p.CSVFormat())
		}
		break
	default:
		fmt.Printf("unknown format: %s\n", *format)
		os.Exit(-1)
	}
}
