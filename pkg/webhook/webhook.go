package webhook

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

// Notifier handles webhook notifications
type Notifier struct {
	webhookURL  string
	containerID string
}

// NewNotifier creates a new webhook notifier
func NewNotifier(webhookURL, containerID string) *Notifier {
	return &Notifier{
		webhookURL:  webhookURL,
		containerID: containerID,
	}
}

// NotifyRateLimit sends a notification when rate limit is exceeded
func (n *Notifier) NotifyRateLimit(expectedRPS, receivedRPS int) error {
	if n.webhookURL == "" {
		log.Println("Webhook URL is not set, skipping notification.")
		return nil
	}

	payload := map[string]interface{}{
		"container_id":       n.containerID,
		"limit_expected_rps": expectedRPS,
		"limit_exceeded_rps": receivedRPS,
		"message":            "Rate limit exceeded",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(n.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Printf("Webhook notification sent, response: %d", resp.StatusCode)
	return nil
}
