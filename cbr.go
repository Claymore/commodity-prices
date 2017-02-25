package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/net/html/charset"
)

type CBRValCurs struct {
	XMLName xml.Name    `xml:"ValCurs"`
	Records []CBRRecord `xml:"Record"`
}

type CBRRecord struct {
	Nominal int64  `xml:"Nominal"`
	Value   string `xml:"Value"`
	Date    string `xml:"Date,attr"`
	ID      string `xml:"Id,attr"`
}

type CBRValuta struct {
	XMLName xml.Name        `xml:"Valuta"`
	Items   []CBRValutaItem `xml:"Item"`
}

type CBRValutaItem struct {
	ISOCharCode string `xml:"ISO_Char_Code"`
	ID          string `xml:"ID,attr"`
}

func (cbr *CBR) toID(commodity string) (id string, err error) {
	url := "http://www.cbr.ru/scripts/XML_valFull.asp"
	response, err := cbr.client.Get(url)
	if err != nil {
		return id, err
	}
	defer response.Body.Close()
	decoder := xml.NewDecoder(response.Body)
	decoder.CharsetReader = charset.NewReaderLabel
	var valuta CBRValuta
	err = decoder.Decode(&valuta)
	if err != nil {
		return id, err
	}
	for _, i := range valuta.Items {
		if i.ISOCharCode == commodity {
			return i.ID, nil
		}
	}
	return id, fmt.Errorf("unknown commodity: %s", commodity)
}

type CBR struct {
	client http.Client
}

func (cbr *CBR) Prices(commodity, from, till string) (prices []Price, err error) {
	cbrID, err := cbr.toID(commodity)
	if err != nil {
		return prices, err
	}
	fromTime, err := time.Parse("2006-01-02", from)
	if err != nil {
		return prices, err
	}
	tillTime, err := time.Parse("2006-01-02", till)
	if err != nil {
		return prices, err
	}
	cbrFrom := fromTime.Format("02/01/2006")
	cbrTill := tillTime.Format("02/01/2006")
	cbrURL := fmt.Sprintf("http://www.cbr.ru/scripts/XML_dynamic.asp?date_req1=%s&date_req2=%s&VAL_NM_RQ=%s", cbrFrom, cbrTill, cbrID)
	response, err := cbr.client.Get(cbrURL)
	if err != nil {
		return prices, err
	}
	defer response.Body.Close()
	decoder := xml.NewDecoder(response.Body)
	decoder.CharsetReader = charset.NewReaderLabel
	var cbrRecords CBRValCurs
	err = decoder.Decode(&cbrRecords)
	if err != nil {
		return prices, err
	}
	for _, r := range cbrRecords.Records {
		price, err := r.ToPrice(commodity)
		if err != nil {
			return prices, err
		}
		prices = append(prices, price)
	}
	return prices, nil
}

func (r *CBRRecord) ToPrice(commodity string) (price Price, err error) {
	price.Commodity = commodity
	price.Date, err = time.Parse("02.01.2006", r.Date)
	if err != nil {
		return price, err
	}
	price.Price, err = decimal.NewFromString(strings.Replace(r.Value, ",", ".", -1))
	if err != nil {
		return price, err
	}
	price.Price = price.Price.Div(decimal.New(r.Nominal, 0))
	return price, err
}
