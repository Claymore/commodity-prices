package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

type MOEX struct {
	Market string
	client http.Client
}

type MOEXDocument struct {
	XMLName xml.Name   `xml:"document"`
	Datum   []MOEXData `xml:"data"`
}

type MOEXData struct {
	ID   string    `xml:"id,attr"`
	Rows []MOEXRow `xml:"rows>row"`
}

type MOEXRow struct {
	SecurityID      string `xml:"SECID,attr"`
	TradeDate       string `xml:"TRADEDATE,attr"`
	LegalClosePrice string `xml:"LEGALCLOSEPRICE,attr"`
	ClosePrice      string `xml:"CLOSE,attr"`
	Index           int    `xml:"INDEX,attr"`
	PageSize        int    `xml:"PAGESIZE,attr"`
	Total           int    `xml:"TOTAL,attr"`
}

func (moex *MOEX) page(commodity, from, till string, index int) (prices []Price, eof bool, err error) {
	URL := fmt.Sprintf("http://iss.moex.com/iss/history/engines/stock/markets/%s/securities/%s.xml?from=%s&till=%s&start=%d", moex.Market, commodity, from, till, index)
	response, err := moex.client.Get(URL)
	if err != nil {
		return prices, eof, err
	}
	defer response.Body.Close()
	decoder := xml.NewDecoder(response.Body)
	var document MOEXDocument
	err = decoder.Decode(&document)
	if err != nil {
		return prices, eof, err
	}
	for _, d := range document.Datum {
		switch d.ID {
		case "history":
			for _, r := range d.Rows {
				price, err := r.ToPrice()
				if err != nil {
					return prices, eof, err
				}
				prices = append(prices, price)
			}
			break
		case "history.cursor":
			if len(d.Rows) == 0 {
				continue
			}
			r := d.Rows[0]
			eof = (r.Index + r.PageSize) >= r.Total
			break
		}
	}
	return prices, eof, nil
}

func (moex *MOEX) Prices(commodity, from, till string) (prices []Price, err error) {
	eof := false
	for index := 0; !eof; index += 100 {
		var ps []Price
		ps, eof, err = moex.page(commodity, from, till, index)
		if err != nil {
			return prices, err
		}
		prices = append(prices, ps...)
	}
	return prices, nil
}

func (r *MOEXRow) ToPrice() (price Price, err error) {
	price.Commodity = r.SecurityID
	price.Date, err = time.Parse("2006-01-02", r.TradeDate)
	if err != nil {
		return price, err
	}
	if len(r.LegalClosePrice) > 0 {
		price.Price, err = decimal.NewFromString(r.LegalClosePrice)
		if err != nil {
			return price, err
		}
	}
	if len(r.ClosePrice) > 0 {
		price.Price, err = decimal.NewFromString(r.ClosePrice)
		if err != nil {
			return price, err
		}
	}
	return price, err
}
