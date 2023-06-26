package downloadstock

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
)

// SendSlack sends some messages to the slack notification channel
func SendSlack(msg string) error {
	webhook := os.Getenv("SLACK_WEBHOOK")
	if webhook == "" {
		return errors.New("Empty environment variable slack webhook, please check.")
	}

	// JSON body
	body := []byte(fmt.Sprintf(`{"text": "%s"}`, msg))

	// Create a HTTP post request
	_, err := http.Post(webhook, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	return nil
}
