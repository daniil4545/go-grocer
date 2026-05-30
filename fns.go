package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type Item struct {
	Name     string  `json:"name"`
	Price    int64   `json:"price"`
	Quantity float64 `json:"quantity"`
	Sum      int64   `json:"sum"`
}

type Receipt struct {
	TotalSum   int64  `json:"totalSum"`
	User       string `json:"user"`
	UserInn    string `json:"userInn"`
	Address    string `json:"retailPlaceAddress"`
	TicketDate string `json:"ticketDate"`
	Items      []Item `json:"items"`
}

type CheckData struct {
	JSON Receipt `json:"json"`
}

type CheckResponse struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

const fnsURL = "https://proverkacheka.com/api/v1/check/get"

func sendReceiptRequest(ctx context.Context, token string, photo []byte) (*CheckResponse, error) {
	buf := bytes.Buffer{}
	w := multipart.NewWriter(&buf)
	if err := w.WriteField("token", token); err != nil {
		return nil, fmt.Errorf("write token field: %w", err)
	}
	fw, err := w.CreateFormFile("qrfile", "photo.jpg")
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := fw.Write(photo); err != nil {
		return nil, fmt.Errorf("write photo: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fnsURL, &buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var checkResponse CheckResponse
	if err := json.Unmarshal(body, &checkResponse); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &checkResponse, nil
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func GetReceipt(ctx context.Context, token string, photo []byte) (*Receipt, error) {
	const maxAttempts = 5

	for range maxAttempts {
		resp, err := sendReceiptRequest(ctx, token, photo)
		if err != nil {
			return nil, err
		}

		switch resp.Code {
		case 1:
			var checkData CheckData
			if err := json.Unmarshal(resp.Data, &checkData); err != nil {
				return nil, fmt.Errorf("unmarshal receipt data: %w", err)
			}
			if len(checkData.JSON.Items) == 0 {
				return nil, fmt.Errorf("api: receipt has no items")
			}
			return &checkData.JSON, nil
		case 2:
			if err := sleepWithContext(ctx, 2*time.Second); err != nil {
				return nil, err
			}
		case 4:
			if err := sleepWithContext(ctx, 8*time.Second); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("api error: code %d", resp.Code)
		}
	}

	return nil, fmt.Errorf("api: max retry attempts reached")
}
