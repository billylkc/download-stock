package downloadstock

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/sirupsen/logrus"
)

type PubSubMsg struct {
	Message struct {
		Attributes struct {
			Date  string `json:"date"`
			Stage string `json:"stage"`
		} `json:"attributes"`
		Data         string    `json:"data"`
		MessageID    string    `json:"messageId"`
		MessageID0   string    `json:"message_id"`
		PublishTime  time.Time `json:"publishTime"`
		PublishTime0 time.Time `json:"publish_time"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

func init() {
	functions.CloudEvent("DownloadStockEvent", DownloadStockEvent)
}

// DownloadStock
func DownloadStockEvent(ctx context.Context, e event.Event) error {

	var err error

	// Get date
	loc, _ := time.LoadLocation("Asia/Hong_Kong")
	currentTime := time.Now().In(loc).AddDate(0, 0, -1)
	today := currentTime.Format("2006-01-02")

	_ = SendSlack(fmt.Sprintf("Start: %s", currentTime.Format("2006-01-02 15:04:05")))

	// Check Sat and Sun
	weekday := currentTime.Weekday()
	if (weekday == 0) || (weekday == 6) {
		err = fmt.Errorf("Error. Weekend.")
		_ = SendSlack(err.Error())
		return err
	}

	// Create Quandl object
	logger := logrus.New()
	logger.Out = io.Writer(os.Stdout)
	q := NewQuandl(logger, today)

	// Get list of companies in code
	codes, err := GetCompanyList()
	if err != nil {
		return fmt.Errorf("Can not get company list. %v.\n", err)
	}

	// TODO: Add checking in bq records, maybe print min max stock code as well

	// Handle stages, split the codes in two parts. one or two.
	var halfCodes []Company
	var obj PubSubMsg
	err = json.Unmarshal([]byte(e.Data()), &obj)
	if err != nil {
		return err
	}
	inputStage := obj.Message.Attributes.Stage
	midIdx := len(codes) / 2 // find mid element
	if inputStage == "one" {
		halfCodes = codes[:midIdx]
	} else if inputStage == "two" {
		halfCodes = codes[midIdx:]
	} else {
		return fmt.Errorf("Invalid input from pubsub. Input stage - %s.", inputStage)
	}

	_ = SendSlack(fmt.Sprintf("Stage - %s. List of stocks - %d", inputStage, len(halfCodes)))

	// Split stocks into 10 different stages
	var records []HistoricalPrice // Final result
	var errCodes []string         // List of codes with no records
	stage, nsplits := 0, 200

	for i, code := range halfCodes {

		// Print stage after every 200 stocks
		if i%nsplits == 0 {

			if len(errCodes) > 0 { // Print error message
				_ = SendSlack(fmt.Sprintf("No records - [%s]", strings.Join(errCodes, ",")))
				errCodes = []string{} // reset
			}
			_ = SendSlack(fmt.Sprintf("Started stage - %d", stage))
			stage += 1
		}

		r, err := q.GetStock(code.StockCode, today)
		if err != nil {
			errMsg := err.Error()
			errCode := strings.ReplaceAll(errMsg, "Records not found - ", "")
			errCodes = append(errCodes, errCode)
		}
		records = append(records, r...)
	}

	// Send last error codes
	if len(errCodes) > 0 {
		_ = SendSlack(fmt.Sprintf("No records - [%s]", strings.Join(errCodes, ",")))
	}

	// Insert to bigquery
	err = InsertStocks(records)
	if err != nil {
		errMsg := fmt.Sprintf("Something wrong during bq indert. %s\n", err.Error())
		_ = SendSlack(errMsg)
	}

	_ = SendSlack(fmt.Sprintf("Successfully Insert - %d records", len(records)))
	_ = SendSlack(fmt.Sprintf("Finished - %d", len(records)))
	_ = SendSlack(fmt.Sprintf("Finished: %s", time.Now().In(loc).Format("2006-01-02 15:04:05")))
	return nil
}
