package cbr

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/net/html/charset"

	"github.com/Claymore/commodity-prices/price"
)

type exchangeRate struct {
	XMLName xml.Name `xml:"ValCurs"`
	Records []record `xml:"Record"`
}

type record struct {
	Nominal int64  `xml:"Nominal"`
	Value   string `xml:"Value"`
	Date    string `xml:"Date,attr"`
	ID      string `xml:"Id,attr"`
}

type currency struct {
	XMLName xml.Name `xml:"Valuta"`
	Items   []item   `xml:"Item"`
}

type item struct {
	ISOCharCode string `xml:"ISO_Char_Code"`
	ParentCode  string `xml:"ParentCode"`
}

func (cbr *Client) id(commodity string) (id string, err error) {
	url := "http://www.cbr.ru/scripts/XML_valFull.asp"
	response, err := cbr.client.Get(url)
	if err != nil {
		return id, err
	}
	defer response.Body.Close()
	decoder := xml.NewDecoder(response.Body)
	decoder.CharsetReader = charset.NewReaderLabel
	var currency currency
	err = decoder.Decode(&currency)
	if err != nil {
		return id, err
	}
	for _, i := range currency.Items {
		if i.ISOCharCode == commodity {
			return i.ParentCode, nil
		}
	}
	return id, fmt.Errorf("unknown commodity: %s", commodity)
}

type Client struct {
	client http.Client
}

func NewClient() *Client {
	return &Client{}
}

func (cbr Client) Prices(commodity, from, till string) (prices []price.Price, err error) {
	cbrID, err := cbr.id(commodity)
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
	var rate exchangeRate
	err = decoder.Decode(&rate)
	if err != nil {
		return prices, err
	}
	for _, r := range rate.Records {
		price, err := r.ToPrice(commodity)
		if err != nil {
			return prices, err
		}
		prices = append(prices, price)
	}
	return prices, nil
}

func (r *record) ToPrice(commodity string) (price price.Price, err error) {
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
