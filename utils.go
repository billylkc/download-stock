package downloadstock

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
)

// PrettyPrint to print struct in a readable way
func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

// Insert to bq
func InsertStocks(records []HistoricalPrice) error {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "stock-lib")
	defer client.Close()

	if err != nil {
		return err
	}
	ins := client.Dataset("stock").Table("stock").Inserter()

	// Schema is inferred from the score type.
	if err := ins.Put(ctx, records); err != nil {
		return err
	}
	return nil
}

// CheckRecords checks the records in bq, and see if the records already exist
func CheckRecords(date string) error {

	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "stock-lib")
	defer client.Close()
	if err != nil {
		return errors.New("Failed to create bq client")
	}

	q := client.Query(fmt.Sprintf(`SELECT count(1) FROM %s WHERE date = "%s"`, "`"+"stock-lib.stock.stock"+"`", date))
	it, err := q.Read(ctx)
	if err != nil {
		return fmt.Errorf("Something's wrong with the sql statement in check records. %w.", err)
	}

	var values []bigquery.Value
	for {
		err := it.Next(&values)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("Something's wrong with the bq iterator. %w.", err)
		}
	}

	// Check no of records of the date provided
	var nrecords int64
	if len(values) > 0 {
		nrecords = values[0].(int64)
	}
	if nrecords != 0 {
		return fmt.Errorf("Error. Records already exists for date %s.", date)
	}

	return nil
}
