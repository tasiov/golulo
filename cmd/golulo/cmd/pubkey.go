package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tasiov/golulo/cmd/golulo/internal"
)

var pubkeyCmd = &cobra.Command{
	Use:   "pubkey",
	Short: "Display public key from keypair file",
	RunE: func(cmd *cobra.Command, args []string) error {
		solanaClient, err := internal.NewSolanaClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		fmt.Printf("Public Key: %s\n", solanaClient.PublicKey.String())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pubkeyCmd)
}
