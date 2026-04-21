package resend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/biangacila/email-service/internal/domain"
)

type Client struct {
	apiKey string
	client *http.Client
}

func NewResendClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (r *Client) Send(ctx context.Context, email domain.Email) error {
	url := "https://api.resend.com/emails"

	payload := map[string]interface{}{
		"from":    email.From,
		"to":      email.Receiver,
		"subject": email.Subject,
		"html":    email.HTML,
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("failed: %d", resp.StatusCode)
	}

	return nil
}
