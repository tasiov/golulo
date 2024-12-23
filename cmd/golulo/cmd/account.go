package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tasiov/golulo/cmd/golulo/internal"
)

// AccountSettings represents user account settings
type AccountSettings struct {
	Owner            string  `json:"owner"`
	AllowedProtocols string  `json:"allowedProtocols"`
	Homebase         *string `json:"homebase"`
	MinimumRate      float64 `json:"minimumRate"` // number
}

// AccountResponse represents the response from the account API
type AccountResponse struct {
	Data struct {
		TotalValue     float64         `json:"totalValue"`
		InterestEarned float64         `json:"interestEarned"`
		RealtimeAPY    float64         `json:"realtimeAPY"`
		Settings       AccountSettings `json:"settings"`
	} `json:"data"`
}

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Get account information",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create Solana client to get wallet pubkey
		client, err := internal.NewSolanaClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		log.WithField("wallet", client.WalletPubKey().String()).
			Info("Fetching account information")

		// Create HTTP request
		url := "https://api.flexlend.fi/account"
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		log.WithField("url", url).Debug("Making API request")

		// Set headers
		req.Header.Set("Accept", "application/json")
		req.Header.Set("x-wallet-pubkey", client.WalletPubKey().String())

		apiKey := viper.GetString("lulo-api-key")
		if apiKey == "" {
			return fmt.Errorf("FLEXLEND_API_KEY environment variable not set")
		}
		req.Header.Set("x-api-key", apiKey)

		// Make the request
		httpClient := &http.Client{}
		resp, err := httpClient.Do(req)
		if err != nil {
			log.WithError(err).Error("Failed to make request")
			return fmt.Errorf("failed to make request: %w", err)
		}
		defer resp.Body.Close()

		// Check status code
		if resp.StatusCode != http.StatusOK {
			log.WithField("statusCode", resp.StatusCode).Error("Unexpected status code")
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		// Parse response
		var response AccountResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			log.WithError(err).Error("Failed to decode response")
			return fmt.Errorf("failed to decode response: %w", err)
		}

		// Log account information
		log.WithFields(log.Fields{
			"totalValue":     response.Data.TotalValue,
			"interestEarned": response.Data.InterestEarned,
			"realtimeAPY":    response.Data.RealtimeAPY,
		}).Info("Account overview")

		log.WithFields(log.Fields{
			"owner":            response.Data.Settings.Owner,
			"allowedProtocols": response.Data.Settings.AllowedProtocols,
			"homebase":         response.Data.Settings.Homebase,
			"minimumRate":      response.Data.Settings.MinimumRate,
		}).Debug("Account settings")

		// Pretty print the response
		prettyJSON, err := json.MarshalIndent(response.Data, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format response: %w", err)
		}
		fmt.Println(string(prettyJSON))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(accountCmd)
}
