package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/spf13/viper"
)

// SolanaClient wraps RPC client and keypair info
type SolanaClient struct {
	RpcClient  *rpc.Client
	PublicKey  solana.PublicKey
	PrivateKey solana.PrivateKey
}

// NewSolanaClient creates a new client from config values
func NewSolanaClient() (*SolanaClient, error) {
	// Get keypair path from config
	keypairPath := viper.GetString("keypair")
	if keypairPath == "" {
		return nil, fmt.Errorf("keypair path not set in config")
	}

	// Read keypair file
	keypairBytes, err := os.ReadFile(keypairPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read keypair file: %w", err)
	}

	// Parse JSON array
	var secretKey []uint8
	if err := json.Unmarshal(keypairBytes, &secretKey); err != nil {
		return nil, fmt.Errorf("failed to parse keypair file: %w", err)
	}

	// Convert to Solana private key
	privateKey := solana.PrivateKey(secretKey)
	publicKey := privateKey.PublicKey()

	// Create RPC client
	rpcURL := viper.GetString("rpc-url")
	if rpcURL == "" {
		return nil, fmt.Errorf("RPC URL not set in config")
	}
	rpcURL += "?api-key=" + viper.GetString("rpc-api-key")

	return &SolanaClient{
		RpcClient:  rpc.New(rpcURL),
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}

// WalletPubKey returns the client's public key
func (c *SolanaClient) WalletPubKey() solana.PublicKey {
	return c.PublicKey
}

// CreateTransaction creates a new transaction with the provided instructions
func (c *SolanaClient) CreateTransaction(ctx context.Context, instructions []solana.Instruction) (*solana.Transaction, error) {
	// Get recent blockhash
	recent, err := c.RpcClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent blockhash: %w", err)
	}

	tx, err := solana.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(c.PublicKey),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return tx, nil
}

// SignTransaction signs a transaction with the client's private key
func (c *SolanaClient) SignTransaction(tx *solana.Transaction) (*solana.Transaction, error) {
	_, err := tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(c.PublicKey) {
			return &c.PrivateKey
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return tx, nil
}

// SendTransaction sends a signed transaction
func (c *SolanaClient) SendTransaction(ctx context.Context, tx *solana.Transaction) (solana.Signature, error) {
	sig, err := c.RpcClient.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: rpc.CommitmentFinalized,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return sig, nil
}

// CreateSignAndSendTransaction combines transaction creation, signing, and sending into one method
func (c *SolanaClient) CreateSignAndSendTransaction(ctx context.Context, instructions []solana.Instruction) (solana.Signature, error) {
	// Create transaction
	tx, err := c.CreateTransaction(ctx, instructions)
	if err != nil {
		return solana.Signature{}, err
	}

	// Sign transaction
	signedTx, err := c.SignTransaction(tx)
	if err != nil {
		return solana.Signature{}, err
	}

	// Send transaction
	return c.SendTransaction(ctx, signedTx)
}
