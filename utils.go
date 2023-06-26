package downloadstock

import (
	"context"
	"encoding/json"

	"cloud.google.com/go/bigquery"
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
