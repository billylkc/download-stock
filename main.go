package downloadstock

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/sirupsen/logrus"
)

func init() {
	functions.HTTP("DownloadStock", DownloadStock)
}

// Function for testing
func TestFunc(w http.ResponseWriter, r *http.Request) {
	codes, err := GetCompanyList()

	if err != nil {
		fmt.Fprint(w, html.EscapeString(err.Error()))
	}
	fmt.Println(len(codes))

	if len(codes) > 10 {
		codes = codes[0:10]
	}
	fmt.Println(PrettyPrint(codes))

}

// DownloadStock
func DownloadStock(w http.ResponseWriter, r *http.Request) {

	// Get date
	loc, _ := time.LoadLocation("Asia/Hong_Kong")
	currentTime := time.Now().In(loc).AddDate(0, 0, -1)
	today := currentTime.Format("2006-01-02")

	// Check record first
	err := CheckRecords(today)
	if err != nil {
		_ = SendSlack(err.Error())
	}

	_ = SendSlack(fmt.Sprintf("Start: %s", currentTime.Format("2006-01-02 15:04:05")))

	// Create Quandl object
	logger := logrus.New()
	logger.Out = io.Writer(os.Stdout)
	q := NewQuandl(logger, today)

	// Get list of companies in code
	codes, err := GetCompanyList()
	if err != nil {
		fmt.Fprint(w, html.EscapeString(err.Error()))
	}
	_ = SendSlack(fmt.Sprintf("List of stocks - %d", len(codes)))

	// Split stocks into 10 different stages
	var records []HistoricalPrice // Final result
	var errCodes []string         // List of codes with no records
	stage, nsplits := 0, 200

	for i, code := range codes {

		// Print stage after every 200 stocks
		if i%nsplits == 0 {

			if len(errCodes) > 0 { // Print error message
				_ = SendSlack(fmt.Sprintf("No records - [%s]", strings.Join(errCodes, ",")))
				errCodes = []string{} // reset
			}
			_ = SendSlack(fmt.Sprintf("Started stage - %d", stage))
			stage += 1
		}

		//
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
	fmt.Fprint(w, html.EscapeString("w/e"))
}
