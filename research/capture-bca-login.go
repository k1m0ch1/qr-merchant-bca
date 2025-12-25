package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type CapturedRequest struct {
	RequestID string                 `json:"request_id"`
	URL       string                 `json:"url"`
	Method    string                 `json:"method"`
	Headers   map[string]interface{} `json:"headers"`
	PostData  string                 `json:"post_data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

type CapturedResponse struct {
	RequestID string                 `json:"request_id"`
	URL       string                 `json:"url"`
	Status    int64                  `json:"status"`
	Headers   map[string]interface{} `json:"headers"`
	Body      string                 `json:"body,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

func main() {
	// Create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Storage for captured data
	requests := make([]CapturedRequest, 0)
	responses := make([]CapturedResponse, 0)

	// Listen for network events
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			req := CapturedRequest{
				RequestID: string(ev.RequestID),
				URL:       ev.Request.URL,
				Method:    ev.Request.Method,
				Headers:   ev.Request.Headers,
				PostData:  ev.Request.PostData,
				Timestamp: time.Now(),
			}
			requests = append(requests, req)
			fmt.Printf("[REQUEST] %s %s\n", ev.Request.Method, ev.Request.URL)

		case *network.EventResponseReceived:
			resp := CapturedResponse{
				RequestID: string(ev.RequestID),
				URL:       ev.Response.URL,
				Status:    ev.Response.Status,
				Headers:   ev.Response.Headers,
				Timestamp: time.Now(),
			}
			responses = append(responses, resp)
			fmt.Printf("[RESPONSE] %d %s\n", ev.Response.Status, ev.Response.URL)
		}
	})

	// Run browser automation
	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate("https://qr.klikbca.com/login"),
		chromedp.Sleep(3*time.Second),
		chromedp.WaitVisible(`input[type="email"]`, chromedp.ByQuery),
		chromedp.SendKeys(`input[type="email"]`, "someemail@gmail.com", chromedp.ByQuery),
		chromedp.SendKeys(`input[type="password"]`, "somepasword", chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		chromedp.Click(`button:contains("Masuk")`, chromedp.ByQuery),
		chromedp.Sleep(5*time.Second),
	)

	if err != nil {
		log.Fatal(err)
	}

	// Save captured data
	data := map[string]interface{}{
		"requests":  requests,
		"responses": responses,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile("C:\\tmp\\bca-login-capture.json", jsonData, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n✓ Captured %d requests and %d responses\n", len(requests), len(responses))
	fmt.Println("✓ Saved to C:\\tmp\\bca-login-capture.json")
}
