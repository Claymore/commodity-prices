package moex

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"

	"github.com/Claymore/commodity-prices/price"
)

type Client struct {
	market string
	client http.Client
}

func NewClient(market string) *Client {
	return &Client{market: market}
}

type document struct {
	XMLName xml.Name `xml:"document"`
	Datum   []data   `xml:"data"`
}

type data struct {
	ID   string `xml:"id,attr"`
	Rows []row  `xml:"rows>row"`
}

type row struct {
	SecurityID      string `xml:"SECID,attr"`
	TradeDate       string `xml:"TRADEDATE,attr"`
	LegalClosePrice string `xml:"LEGALCLOSEPRICE,attr"`
	ClosePrice      string `xml:"CLOSE,attr"`
	Index           int    `xml:"INDEX,attr"`
	PageSize        int    `xml:"PAGESIZE,attr"`
	Total           int    `xml:"TOTAL,attr"`
}

func (moex *Client) page(commodity, from, till string, index int) (prices []price.Price, eof bool, err error) {
	URL := fmt.Sprintf("http://iss.moex.com/iss/history/engines/stock/markets/%s/securities/%s.xml?from=%s&till=%s&start=%d", moex.market, commodity, from, till, index)
	response, err := moex.client.Get(URL)
	if err != nil {
		return prices, eof, err
	}
	defer response.Body.Close()
	decoder := xml.NewDecoder(response.Body)
	var document document
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

func (moex Client) Prices(commodity, from, till string) (prices []price.Price, err error) {
	eof := false
	for index := 0; !eof; index += 100 {
		var ps []price.Price
		ps, eof, err = moex.page(commodity, from, till, index)
		if err != nil {
			return prices, err
		}
		prices = append(prices, ps...)
	}
	return prices, nil
}

func (r *row) ToPrice() (price price.Price, err error) {
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
