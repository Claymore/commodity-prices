package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
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
	SecurityID      string   `xml:"SECID,attr"`
	TradeDate       string   `xml:"TRADEDATE,attr"`
	LegalClosePrice *float64 `xml:"LEGALCLOSEPRICE,attr"`
	ClosePrice      *float64 `xml:"CLOSE,attr"`
}

func (moex *MOEX) Prices(commodity, from, till string) (prices []Price, err error) {
	URL := fmt.Sprintf("http://iss.moex.com/iss/history/engines/stock/markets/%s/securities/%s.xml?from=%s&till=%s", moex.Market, commodity, from, till)
	response, err := moex.client.Get(URL)
	if err != nil {
		return prices, err
	}
	defer response.Body.Close()
	decoder := xml.NewDecoder(response.Body)
	var document MOEXDocument
	err = decoder.Decode(&document)
	if err != nil {
		return prices, err
	}
	for _, d := range document.Datum {
		if d.ID != "history" {
			continue
		}
		for _, r := range d.Rows {
			price, err := r.ToPrice()
			if err != nil {
				return prices, err
			}
			prices = append(prices, price)
		}
	}
	return prices, nil
}

func (r *MOEXRow) ToPrice() (price Price, err error) {
	price.Commodity = r.SecurityID
	price.Date, err = time.Parse("2006-01-02", r.TradeDate)
	if err != nil {
		return price, err
	}
	if r.LegalClosePrice != nil {
		price.Price = *r.LegalClosePrice
	}
	if r.ClosePrice != nil {
		price.Price = *r.ClosePrice
	}
	return price, err
}
