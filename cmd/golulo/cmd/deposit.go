package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tasiov/golulo/cmd/golulo/internal"
)

var (
	amount      float64
	mintAddress string
)

// DepositRequest represents the request body for the deposit API
type DepositRequest struct {
	Owner         string `json:"owner"`
	MintAddress   string `json:"mintAddress"`
	DepositAmount string `json:"depositAmount"`
}

// TransactionMeta represents a single transaction in the response
type TransactionMeta struct {
	Transaction  string  `json:"transaction"`
	Protocol     string  `json:"protocol"`
	TotalDeposit float64 `json:"totalDeposit"`
}

// DepositResponse represents the response from the deposit API
type DepositResponse struct {
	Data struct {
		TransactionMeta []TransactionMeta `json:"transactionMeta"`
	} `json:"data"`
}

var depositCmd = &cobra.Command{
	Use:   "deposit",
	Short: "Deposit tokens into a Lulo reserve",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create Solana client
		client, err := internal.NewSolanaClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		// Create deposit request
		request := DepositRequest{
			Owner:         client.WalletPubKey().String(),
			MintAddress:   mintAddress,
			DepositAmount: fmt.Sprintf("%.0f", amount),
		}

		logrus.WithFields(logrus.Fields{
			"owner":         request.Owner,
			"mintAddress":   request.MintAddress,
			"depositAmount": request.DepositAmount,
		}).Info("Creating deposit request")

		// Convert request to JSON
		jsonData, err := json.Marshal(request)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}

		// Create HTTP request with priority fee
		url := fmt.Sprintf("https://api.flexlend.fi/generate/account/deposit?priorityFee=%s", viper.GetString("priority-fee"))
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
		var response DepositResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		logrus.WithField("transactionCount", len(response.Data.TransactionMeta)).
			Info("Received transactions from API")

		ctx := context.Background()

		blockhash, err := client.RpcClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
		if err != nil {
			return fmt.Errorf("failed to get latest blockhash: %w", err)
		}

		for _, meta := range response.Data.TransactionMeta {
			logrusger := logrus.WithFields(logrus.Fields{
				"protocol":     meta.Protocol,
				"totalDeposit": meta.TotalDeposit,
			})

			logrusger.Info("Processing transaction")

			// Decode base64 transaction
			txBytes, err := base64.StdEncoding.DecodeString(meta.Transaction)
			if err != nil {
				return fmt.Errorf("failed to decode transaction for %s: %w", meta.Protocol, err)
			}

			// Deserialize the transaction
			tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(txBytes))
			if err != nil {
				return fmt.Errorf("failed to deserialize transaction for %s: %w", meta.Protocol, err)
			}

			tx.Message.RecentBlockhash = blockhash.Value.Blockhash

			logrusger.WithFields(logrus.Fields{
				"requiredSignatures":       tx.Message.Header.NumRequiredSignatures,
				"readonlySigned":           tx.Message.Header.NumReadonlySignedAccounts,
				"readonlyUnsigned":         tx.Message.Header.NumReadonlyUnsignedAccounts,
				"addressTableLookupsCount": len(tx.Message.AddressTableLookups),
			}).Debug("Transaction details")

			// logrus address table lookups if present
			if len(tx.Message.AddressTableLookups) > 0 {
				for i, lookup := range tx.Message.AddressTableLookups {
					logrusger.WithFields(logrus.Fields{
						"index":      i,
						"addressKey": lookup.AccountKey.String(),
					}).Trace("Address table lookup")
				}
			}

			// logrus required signers
			requiredSigners := tx.Message.AccountKeys[:tx.Message.Header.NumRequiredSignatures]
			for i, signer := range requiredSigners {
				logrusger.WithFields(logrus.Fields{
					"index":     i,
					"publicKey": signer.String(),
				}).Debug("Required signer")
			}

			// Create a partially signed transaction
			tx, err = client.SignTransaction(tx)
			if err != nil {
				return fmt.Errorf("failed to sign transaction: %w", err)
			}

			// Verify signatures
			signaturesSet := 0
			for i, sig := range tx.Signatures {
				if !sig.Equals(solana.Signature{}) {
					logrusger.WithFields(logrus.Fields{
						"index":     i,
						"signature": sig.String(),
					}).Trace("Signature present")
					signaturesSet++
				}
			}

			logrusger.WithField("signatureCount", signaturesSet).Debug("Signature verification")

			// Send the partially signed transaction
			sig, err := client.SendTransaction(ctx, tx)
			if err != nil {
				logrusger.WithFields(logrus.Fields{
					"signaturesRequired": tx.Message.Header.NumRequiredSignatures,
					"signaturesPresent":  signaturesSet,
					"error":              err,
				}).Error("Failed to send transaction")
				return fmt.Errorf("failed to send transaction for %s: %w", meta.Protocol, err)
			}

			logrusger.WithField("signature", sig.String()).Info("Transaction sent successfully")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(depositCmd)
	depositCmd.Flags().Float64VarP(&amount, "amount", "a", 0, "Amount to deposit")
	depositCmd.Flags().StringVarP(&mintAddress, "mint", "m", "", "Mint address")
	depositCmd.MarkFlagRequired("amount")
	depositCmd.MarkFlagRequired("mint")
}
