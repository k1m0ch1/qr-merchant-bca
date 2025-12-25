package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

type SessionData struct {
	Token   string              `json:"token"`
	Cookies []*http.Cookie      `json:"cookies"`
	BaseURL string              `json:"base_url"`
	SavedAt time.Time           `json:"saved_at"`
}

type Transaction struct {
	From   string `json:"from"`
	Amount string `json:"amount"`
	Time   string `json:"time"`
	RRN    string `json:"rrn"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Status  string `json:"status"`
	Token   string `json:"token"`
	Message string `json:"message"`
}

type transactionResponse struct {
	Status       string        `json:"status"`
	Transactions []Transaction `json:"transactions"`
	Message      string        `json:"message"`
}

func New(baseURL string) *Client {
	// Create cookie jar to maintain session
	jar, _ := cookiejar.New(nil)

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		},
	}
}

func (c *Client) Login(email, password string) error {
	// First, visit the login page to get session cookies
	loginPageResp, err := c.httpClient.Get(c.baseURL + "/login")
	if err != nil {
		return fmt.Errorf("failed to visit login page: %w", err)
	}
	loginPageResp.Body.Close()

	// Encrypt password using AES-256-CBC (matches the website's "encryptMessi" function)
	encryptedPassword, err := encryptPassword(password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	loginReq := loginRequest{
		Email:    email,
		Password: encryptedPassword,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %w", err)
	}

	// Note: API uses nginx reverse proxy at qr.klikbca.com/api/*
	apiURL := c.baseURL + "/api/session/v1.0.0/add"

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Origin", "https://qr.klikbca.com")
	req.Header.Set("Referer", "https://qr.klikbca.com/login")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("sec-ch-ua", `"Not_A Brand";v="8", "Chromium";v="120"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp loginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return fmt.Errorf("failed to parse login response: %w", err)
	}

	if loginResp.Status != "success" {
		return fmt.Errorf("login failed: %s", loginResp.Message)
	}

	c.token = loginResp.Token
	return nil
}

func (c *Client) GetTransactions(date string) ([]Transaction, error) {
	if c.token == "" {
		return nil, fmt.Errorf("not logged in, call Login() first")
	}

	// Parse date to get start and end timestamps
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	startTime := t.Format("2006-01-02T00:00:00Z")
	endTime := t.Add(24 * time.Hour).Format("2006-01-02T00:00:00Z")

	url := fmt.Sprintf("%s/api/transaction-v2/v2.0.0/list?start_date=%s&end_date=%s",
		c.baseURL, startTime, endTime)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch transactions with status %d: %s", resp.StatusCode, string(body))
	}

	var txResp transactionResponse
	if err := json.Unmarshal(body, &txResp); err != nil {
		return nil, fmt.Errorf("failed to parse transaction response: %w", err)
	}

	return txResp.Transactions, nil
}

// SaveSession saves the current session (token and cookies) to a file
func (c *Client) SaveSession(filename string) error {
	if c.token == "" {
		return fmt.Errorf("no active session to save")
	}

	// Get cookies from the jar
	baseURL, _ := url.Parse(c.baseURL)
	cookies := c.httpClient.Jar.Cookies(baseURL)

	session := SessionData{
		Token:   c.token,
		Cookies: cookies,
		BaseURL: c.baseURL,
		SavedAt: time.Now(),
	}

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// LoadSession loads a session from a file
func (c *Client) LoadSession(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read session file: %w", err)
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return fmt.Errorf("failed to parse session file: %w", err)
	}

	// Check if session is too old (e.g., more than 24 hours)
	if time.Since(session.SavedAt) > 24*time.Hour {
		return fmt.Errorf("session expired, please login again")
	}

	// Set the token
	c.token = session.Token
	c.baseURL = session.BaseURL

	// Restore cookies to the jar
	baseURL, _ := url.Parse(c.baseURL)
	c.httpClient.Jar.SetCookies(baseURL, session.Cookies)

	return nil
}

// NewFromSession creates a client and loads an existing session
func NewFromSession(filename string) (*Client, error) {
	c := New("https://qr.klikbca.com")
	if err := c.LoadSession(filename); err != nil {
		return nil, err
	}
	return c, nil
}
