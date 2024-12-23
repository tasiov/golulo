package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tasiov/golulo/cmd/golulo/internal"
)

var (
	withdrawAll bool
)

// WithdrawRequest represents the request body for the withdraw API
type WithdrawRequest struct {
	Owner          string `json:"owner"`
	MintAddress    string `json:"mintAddress"`
	WithdrawAmount string `json:"withdrawAmount"`
	WithdrawAll    bool   `json:"withdrawAll"`
}

// TransactionMeta represents a single transaction in the response
type WithdrawTransactionMeta struct {
	Transaction   string `json:"transaction"`
	Protocol      string `json:"protocol"`
	TotalWithdraw string `json:"totalWithdraw"`
}

// WithdrawResponse represents the response from the withdraw API
type WithdrawResponse struct {
	Data struct {
		TransactionMeta []WithdrawTransactionMeta `json:"transactionMeta"`
	} `json:"data"`
}

var withdrawCmd = &cobra.Command{
	Use:   "withdraw",
	Short: "Withdraw tokens from a Lulo reserve",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create Solana client
		client, err := internal.NewSolanaClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		// Create withdraw request
		request := WithdrawRequest{
			Owner:          client.WalletPubKey().String(),
			MintAddress:    mintAddress,
			WithdrawAmount: fmt.Sprintf("%.0f", amount),
			WithdrawAll:    withdrawAll,
		}

		logrus.WithFields(logrus.Fields{
			"owner":          request.Owner,
			"mintAddress":    request.MintAddress,
			"withdrawAmount": request.WithdrawAmount,
			"withdrawAll":    request.WithdrawAll,
		}).Info("Creating withdraw request")

		// Convert request to JSON
		jsonData, err := json.Marshal(request)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}

		// Create HTTP request with priority fee
		url := fmt.Sprintf("https://api.flexlend.fi/generate/account/withdraw?priorityFee=%s", viper.GetString("priority-fee"))
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		logrus.WithField("url", url).Debug("Making API request")

		// Set headers
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-wallet-pubkey", client.WalletPubKey().String())
		req.Header.Set("x-api-key", viper.GetString("lulo-api-key"))

		// Make the request
		httpClient := &http.Client{}
		resp, err := httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to make request: %w", err)
		}
		defer resp.Body.Close()

		// Check status code
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		// Parse response
		var response WithdrawResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		logrus.WithField("transactionCount", len(response.Data.TransactionMeta)).
			Info("Received transactions from API")

		b64_txs := make([]string, len(response.Data.TransactionMeta))
		for i, meta := range response.Data.TransactionMeta {
			b64_txs[i] = meta.Transaction
		}

		err = client.HandleB64Transactions(b64_txs)
		if err != nil {
			return fmt.Errorf("failed to handle transactions: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(withdrawCmd)
	withdrawCmd.Flags().Float64VarP(&amount, "amount", "a", 0, "Amount to withdraw")
	withdrawCmd.Flags().StringVarP(&mintAddress, "mint", "m", "", "Mint address")
	withdrawCmd.Flags().BoolVar(&withdrawAll, "all", false, "Withdraw all tokens")

	// Only require amount and mint if not withdrawing all
	withdrawCmd.MarkFlagRequired("mint")

	// Custom validation
	withdrawCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if !withdrawAll && amount == 0 {
			return fmt.Errorf("either --amount or --all flag must be specified")
		}
		return nil
	}
}
