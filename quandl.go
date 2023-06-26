package downloadstock

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Quandl
type Quandl struct {
	logger *logrus.Logger
	limit  int
	start  string // not using start date right now
	end    string
	order  string
}

// HistoricalPrice as the struct for the API result
type HistoricalPrice struct {
	Code     int     `csv:"-"`
	CodeF    string  `csv:"-"` // code in string format
	Date     string  `csv:"Date"`
	Ask      float64 `csv:"Ask"`
	Bid      float64 `csv:"Bid"`
	Open     float64 `csv:"Previous Close"` // open is missing in quandl, using prev close
	High     float64 `csv:"High"`
	Low      float64 `csv:"Low"`
	Close    float64 `csv:"Nominal Price"`
	Volume   int     `csv:"Share Volume (000)"`
	Turnover int     `csv:"Turnover (000)"`
}

// GetStock is the underlying function to get the stock by different code and date settings
func (q *Quandl) GetStock(code int, date string) ([]HistoricalPrice, error) {
	var data []HistoricalPrice

	// Derive input
	if date == "" {
		today := time.Now().Format("2006-01-02")
		q.option(setEndDate(today))
		q.option(setLimit(10000))
	} else {
		q.option(setEndDate(date))
		q.option(setLimit(10))
	}

	codeF := fmt.Sprintf("%05d", code)
	endpoint, _ := q.getEndpoint(code)

	response, err := http.Get(endpoint)
	if err != nil {
		return data, errors.Wrap(err, "something is wrong with the request")
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err := gocsv.UnmarshalBytes(body, &data); err != nil {
		q.logger.Error("unable to unmarshal the response")
		return data, errors.New("unable to unmarshal the response")
	}

	for i, _ := range data {
		data[i].Code = code
		data[i].CodeF = codeF
		data[i].Volume = data[i].Volume * 1000
		data[i].Turnover = data[i].Turnover * 1000
	}

	// Handle date logic
	var matched bool
	var result []HistoricalPrice
	if date == "" {
		matched = true
		result = data
	} else {
		for _, d := range data {
			if d.Date == date {
				matched = true
				result = []HistoricalPrice{d}
			}
		}
	}
	if !matched {
		return []HistoricalPrice{}, errors.New(fmt.Sprintf("Records not found - %d", code))
	}
	return result, nil
}

type option func(*Quandl)

// NewQuandl as Quandl constructor
func NewQuandl(logger *logrus.Logger, date string) Quandl {

	return Quandl{
		logger: logger,
		limit:  10,
		end:    date,
		order:  "desc",
	}
}

// Option sets the options specified.
func (q *Quandl) option(opts ...option) {
	for _, opt := range opts {
		opt(q)
	}
}

// getEndpoint gets the endpoint for the quandl api
func (q *Quandl) getEndpoint(code int) (string, error) {
	token, err := getToken()
	if err != nil {
		return "", err
	}
	codeF := fmt.Sprintf("%05d", code)
	endpoint := fmt.Sprintf("https://www.quandl.com/api/v3/datasets/HKEX/%s/data.csv?limit=%d&end_date=%s&order=%s&api_key=%s", codeF, q.limit, q.end, q.order, token)
	return endpoint, nil
}

// getToken returns the quandl api token
func getToken() (string, error) {
	token := os.Getenv("QUANDL_TOKEN")
	fmt.Println(token)

	if token == "" {
		return "", errors.New("please check you env variable QUANDL_TOKEN")
	}
	return token, nil
}

func setLimit(n int) option {
	return func(q *Quandl) {
		q.limit = n
	}
}
func setOrder(settings string) option {
	return func(q *Quandl) {
		q.order = settings
	}
}
func setStartDate(settings string) option {
	return func(q *Quandl) {
		q.start = settings
	}
}
func setEndDate(settings string) option {
	return func(q *Quandl) {
		q.end = settings
	}
}
