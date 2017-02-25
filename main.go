package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/shopspring/decimal"
)

type Price struct {
	Commodity string
	Date      time.Time
	Price     decimal.Decimal
}

func (p Price) Format(format string) (str string, err error) {
	switch format {
	case "ledger":
		return p.LedgerFormat(), nil
	case "csv":
		return p.CSVFormat(), nil
	default:
		return str, fmt.Errorf("unknown format: %s", format)
	}
}

func (p Price) LedgerFormat() string {
	date := p.Date.Format("2006/01/02")
	price := p.Price.String()
	return fmt.Sprintf("P %s %s %s RUR", date, p.Commodity, price)
}

func (p Price) CSVFormat() string {
	date := p.Date.Format("02.01.2006")
	price := p.Price.String()
	return fmt.Sprintf("%s,%s,%s", p.Commodity, date, price)
}

func monthlyFilter(prices []Price) (filtered []Price) {
	var lastTradingDay Price
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
	var filter = flag.String("filter", "none", "filter dates: none|monthly")
	flag.Parse()

	var err error
	var prices []Price
	switch *source {
	case "cbr":
		cbr := CBR{}
		prices, err = cbr.Prices(*commodity, *from, *till)
		break
	case "moex shares":
		moex := MOEX{Market: "shares"}
		prices, err = moex.Prices(*commodity, *from, *till)
		break
	case "moex index":
		moex := MOEX{Market: "index"}
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
	if *filter == "monthly" {
		prices = monthlyFilter(prices)
	}
	for _, p := range prices {
		output, err := p.Format(*format)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			os.Exit(-1)
		}
		fmt.Printf("%s\n", output)
	}
}
