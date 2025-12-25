package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/k1m0ch1/bcaqr/internal/client"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to QR Merchant BCA and save session",
	Long:  `Interactive login that prompts for email and password, then saves the session for future use`,
	RunE:  runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Prompt for email
	fmt.Print("Email: ")
	var email string
	_, err := fmt.Scanln(&email)
	if err != nil {
		return fmt.Errorf("failed to read email: %w", err)
	}

	// Prompt for password (hidden input)
	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // New line after password input
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	password := string(passwordBytes)

	if email == "" || password == "" {
		return fmt.Errorf("email and password are required")
	}

	// Create client
	c := client.New("https://qr.klikbca.com")

	// Login
	fmt.Printf("\nLogging in as %s...\n", email)
	if err := c.Login(email, password); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Save session
	sessionFile := getSessionFile()
	if err := c.SaveSession(sessionFile); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	fmt.Println("✓ Login successful")
	fmt.Printf("✓ Session saved to %s\n", sessionFile)
	fmt.Println("\nYou can now use other commands without providing credentials.")

	return nil
}

func getSessionFile() string {
	// Try to get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory
		return ".bcaqr_session.json"
	}
	return homeDir + "/.bcaqr_session.json"
}
