package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
)

type CBRValCurs struct {
	XMLName xml.Name    `xml:"ValCurs"`
	Records []CBRRecord `xml:"Record"`
}

type CBRRecord struct {
	Nominal int    `xml:"Nominal"`
	Value   string `xml:"Value"`
	Date    string `xml:"Date,attr"`
	ID      string `xml:"Id,attr"`
}

func toCommodity(id string) string {
	// http://www.cbr.ru/scripts/XML_valFull.asp
	switch id {
	case "R01235":
		return "USD"
	case "R01090":
		return "BYN"
	case "R01135":
		return "HUF"
	default:
		return "Unknown"
	}
}

func toID(commodity string) string {
	switch commodity {
	case "USD":
		return "R01235"
	case "BYN":
		return "R01090"
	case "HUF":
		return "R01135"
	default:
		return "Unknown"
	}
}

type CBR struct {
	client http.Client
}

func (cbr *CBR) Prices(commodity, from, till string) (prices []Price, err error) {
	cbrID := toID(commodity)
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
		price, err := r.ToPrice()
		if err != nil {
			return prices, err
		}
		prices = append(prices, price)
	}
	return prices, nil
}

func (r *CBRRecord) ToPrice() (price Price, err error) {
	price.Commodity = toCommodity(r.ID)
	price.Date, err = time.Parse("02.01.2006", r.Date)
	if err != nil {
		return price, err
	}
	price.Price, err = strconv.ParseFloat(strings.Replace(r.Value, ",", ".", -1), 64)
	if err != nil {
		return price, err
	}
	price.Price = price.Price / float64(r.Nominal)
	return price, err
}
