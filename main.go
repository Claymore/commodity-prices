package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Claymore/commodity-prices/cbr"
	"github.com/Claymore/commodity-prices/moex"
	"github.com/Claymore/commodity-prices/price"
)

type formatFunc func(price.Price) string

func ledgerFormat(p price.Price) string {
	date := p.Date.Format("2006/01/02")
	price := p.Price.String()
	return fmt.Sprintf("P %s %s %s RUR", date, p.Commodity, price)
}

func csvFormat(p price.Price) string {
	date := p.Date.Format("02.01.2006")
	price := p.Price.String()
	return fmt.Sprintf("%s,%s,%s", p.Commodity, date, price)
}

func monthlyFilter(prices []price.Price) (filtered []price.Price) {
	var lastTradingDay price.Price
	firstRecord := true
	for _, p := range prices {
		if firstRecord {
			lastTradingDay = p
			firstRecord = false
		}
		if p.Date.Month() != lastTradingDay.Date.Month() {
			filtered = append(filtered, lastTradingDay)
		}
		if p.Date.After(lastTradingDay.Date) {
			lastTradingDay = p
		}
	}
	if len(prices) > 0 {
		filtered = append(filtered, lastTradingDay)
	}
	return filtered
}

func main() {
	today := time.Now().Format("2006-01-02")
	var commodity = flag.String("commodity", "USD", "commodity")
	var from = flag.String("from", today, "from")
	var till = flag.String("till", today, "till")
	var format = flag.String("format", "csv", "format output: ledger|csv")
	var source = flag.String("source", "cbr", "source: cbr|moex shares|moex index")
	var board = flag.String("board", "TQTF", "moex board")
	var filter = flag.String("filter", "none", "filter dates: none|monthly")
	flag.Parse()

	var formatFunc formatFunc
	switch *format {
	case "ledger":
		formatFunc = ledgerFormat
	case "csv":
		formatFunc = csvFormat
	default:
		fmt.Printf("error: unknown format: %s\n", *format)
		os.Exit(-1)
	}

	var err error
	var prices []price.Price
	var teller price.Teller
	switch *source {
	case "cbr":
		teller = cbr.NewClient()
		break
	case "moex shares":
		teller = moex.NewClient("shares", *board)
		break
	case "moex index":
		teller = moex.NewClient("index", *board)
		break
	default:
		err = fmt.Errorf("unknown source: %s", *source)
		break
	}

	prices, err = teller.Prices(*commodity, *from, *till)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(-1)
	}

	if *filter == "monthly" {
		prices = monthlyFilter(prices)
	}

	for _, p := range prices {
		output := formatFunc(p)
		fmt.Printf("%s\n", output)
	}
}
