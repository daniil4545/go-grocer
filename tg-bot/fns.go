package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
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
	TicketDate string `json:"ticketDate"`
	Items      []Item `json:"items"`
}

type CheckData struct {
	JSON Receipt `json:"json"`
}

type CheckResponse struct {
	Code int       `json:"code"`
	Data CheckData `json:"data"`
}

func GetReceipt(ctx context.Context, token string, photo []byte) (*Receipt, error) {
	buf := bytes.Buffer{}
	w := multipart.NewWriter(&buf)
	w.WriteField("token", token)
	fw, err := w.CreateFormFile("qrfile", "photo.jpg")
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	fw.Write(photo)
	w.Close()

	const url = "https://proverkacheka.com/api/v1/check/get"
	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	var checkResponse CheckResponse
	if err := json.Unmarshal(data, &checkResponse); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if checkResponse.Code != 1 {
		return nil, fmt.Errorf("api error: code %d", checkResponse.Code)
	}
	return &checkResponse.Data.JSON, nil
}
