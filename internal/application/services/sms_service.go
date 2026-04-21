package services

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/biangacila/email-service/pkg/config"
)

type SmsService interface {
	Send(phone string, message string) error
}

type SmsServiceImpl struct {
	smsGatewayUsername string
	smsGatewayPassword string
	smsGatewayUrl      string
	httpClient         *http.Client
}

func NewSmsService(config *config.Config) *SmsServiceImpl {
	return &SmsServiceImpl{
		smsGatewayUsername: config.SmsGatewayUsername,
		smsGatewayPassword: config.SmsGatewayPassword,
		smsGatewayUrl:      config.SmsGatewayUrl, // e.g. https://smsgw1.gsm.co.za/xml/send
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (s *SmsServiceImpl) Send(phone string, message string) error {
	// Encode query params
	params := url.Values{}
	params.Add("number", phone) // e.g. +27831234567
	params.Add("message", message)

	fullURL := fmt.Sprintf("%s?%s", s.smsGatewayUrl, params.Encode())

	// Create request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Basic Auth
	auth := s.smsGatewayUsername + ":" + s.smsGatewayPassword
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", "Basic "+encoded)

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send sms: %w", err)
	}
	defer resp.Body.Close()

	// Validate response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sms gateway error: %s", resp.Status)
	}

	return nil
}
