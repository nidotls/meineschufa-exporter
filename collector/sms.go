package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func NewSmsApp() (SmsApp, error) {
	provider := os.Getenv("SMS_PROVIDER")

	if provider == "httpsms" {
		return &httpSmsApp{
			number: os.Getenv("HTTPSMS_NUMBER"),
			apiKey: os.Getenv("HTTPSMS_API_KEY"),
		}, nil
	} else if provider == "webhook" {
		return &webhookSmsApp{}, nil
	}

	return nil, fmt.Errorf("unknown sms provider: %s", provider)
}

type SmsApp interface {
	GetCode() (string, error)
}

type httpSmsApp struct {
	number string
	apiKey string
}

type httpSmsResult struct {
	Data []struct {
		Id                string      `json:"id"`
		RequestId         interface{} `json:"request_id"`
		Owner             string      `json:"owner"`
		UserId            string      `json:"user_id"`
		Contact           string      `json:"contact"`
		Content           string      `json:"content"`
		Encrypted         bool        `json:"encrypted"`
		Type              string      `json:"type"`
		Status            string      `json:"status"`
		Sim               string      `json:"sim"`
		SendTime          interface{} `json:"send_time"`
		RequestReceivedAt time.Time   `json:"request_received_at"`
		CreatedAt         time.Time   `json:"created_at"`
		UpdatedAt         time.Time   `json:"updated_at"`
		OrderTimestamp    time.Time   `json:"order_timestamp"`
		LastAttemptedAt   interface{} `json:"last_attempted_at"`
		ScheduledAt       interface{} `json:"scheduled_at"`
		SentAt            interface{} `json:"sent_at"`
		ScheduledSendTime interface{} `json:"scheduled_send_time"`
		DeliveredAt       interface{} `json:"delivered_at"`
		ExpiredAt         interface{} `json:"expired_at"`
		FailedAt          interface{} `json:"failed_at"`
		CanBePolled       bool        `json:"can_be_polled"`
		SendAttemptCount  int         `json:"send_attempt_count"`
		MaxSendAttempts   int         `json:"max_send_attempts"`
		ReceivedAt        time.Time   `json:"received_at"`
		FailureReason     interface{} `json:"failure_reason"`
	} `json:"data"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func (a *httpSmsApp) GetCode() (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://api.httpsms.com/v1/messages?"+url.Values{
		"owner":   {a.number},
		"contact": {"Schufa"},
		"limit":   {"1"},
	}.Encode(), nil)

	if err != nil {
		return "", err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-API-KEY", a.apiKey)

	res, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	var result httpSmsResult

	// parse json
	err = json.Unmarshal(body, &result)

	if err != nil {
		return "", err
	}

	code := strings.Replace(result.Data[0].Content, " ist Ihr Sicherheitscode für den Login auf www.meineSCHUFA.de", "", 1)

	return code, nil
}

type webhookSmsApp struct{}

func (a *webhookSmsApp) GetCode() (string, error) {
	return "", nil
}
