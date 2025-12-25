package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/k1m0ch1/bcaqr/internal/client"
	"github.com/spf13/cobra"
)

var (
	email    string
	password string
	date     string
)

var transactionsCmd = &cobra.Command{
	Use:   "transactions",
	Short: "Fetch transactions from QR Merchant BCA",
	Long:  `Fetch transactions for a specific date. Uses saved session from 'bcaqr login' if available.`,
	RunE:  runTransactions,
}

func init() {
	transactionsCmd.Flags().StringVarP(&email, "email", "e", "", "Email for login (optional if logged in)")
	transactionsCmd.Flags().StringVarP(&password, "password", "p", "", "Password for login (optional if logged in)")
	transactionsCmd.Flags().StringVarP(&date, "date", "d", "", "Date to fetch transactions (YYYY-MM-DD), defaults to today")
}

func runTransactions(cmd *cobra.Command, args []string) error {
	// Default to today if no date specified
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	// Validate date format
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid date format, use YYYY-MM-DD: %w", err)
	}

	var c *client.Client
	sessionFile := getSessionFile()

	// Try to use stored session first
	if email == "" && password == "" {
		// Try to load from session
		if _, err := os.Stat(sessionFile); err == nil {
			fmt.Println("Using saved session...")
			c, err = client.NewFromSession(sessionFile)
			if err != nil {
				return fmt.Errorf("failed to load session: %w\nPlease run 'bcaqr login' or provide --email and --password", err)
			}
			fmt.Println("✓ Session loaded")
		} else {
			return fmt.Errorf("no saved session found\nPlease run 'bcaqr login' or provide --email and --password")
		}
	} else {
		// Use provided credentials
		if email == "" || password == "" {
			return fmt.Errorf("both --email and --password are required when not using saved session")
		}

		c = client.New("https://qr.klikbca.com")
		fmt.Printf("Logging in as %s...\n", email)
		if err := c.Login(email, password); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}
		fmt.Println("✓ Login successful")
	}

	// Fetch transactions
	fmt.Printf("Fetching transactions for %s...\n", date)
	transactions, err := c.GetTransactions(date)
	if err != nil {
		return fmt.Errorf("failed to fetch transactions: %w", err)
	}

	// Display results
	fmt.Printf("\nTransactions for %s:\n", date)
	fmt.Println("============================================")

	if len(transactions) == 0 {
		fmt.Println("No transactions found")
		return nil
	}

	for i, tx := range transactions {
		fmt.Printf("\n#%d\n", i+1)
		fmt.Printf("  From:   %s\n", tx.From)
		fmt.Printf("  Amount: Rp %s\n", tx.Amount)
		fmt.Printf("  Time:   %s\n", tx.Time)
		if tx.RRN != "" {
			fmt.Printf("  RRN:    %s\n", tx.RRN)
		}
	}

	fmt.Printf("\n============================================\n")
	fmt.Printf("Total: %d transactions\n", len(transactions))

	return nil
}
