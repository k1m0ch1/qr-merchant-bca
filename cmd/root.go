package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bcaqr",
	Short: "CLI tool to interact with QR Merchant BCA",
	Long:  `A command line tool to login to qr.klikbca.com and fetch transaction data`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(transactionsCmd)
}
