package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile          string
	keypairPath      string
	rpcURL           string
	rpcAPIKey        string
	luloAPIKey       string
	priorityFee      string
	allowedProtocols []string
)

var rootCmd = &cobra.Command{
	Use:   "golulo",
	Short: "A CLI for interacting with Lulo Protocol",
	Long: `golulo is a command line interface for interacting with the Lulo Protocol
on the Solana blockchain. It provides commands for managing lending positions,
viewing market data, and more.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().StringVar(&keypairPath, "keypair", "", "path to keypair file")
	rootCmd.PersistentFlags().StringVar(&rpcURL, "rpc-url", "", "RPC server URL")
	rootCmd.PersistentFlags().StringVar(&rpcAPIKey, "rpc-api-key", "", "API key for RPC")
	rootCmd.PersistentFlags().StringVar(&luloAPIKey, "lulo-api-key", "", "API key for Lulo")
	rootCmd.PersistentFlags().StringVar(&priorityFee, "priority-fee", "", "Priority fee for transactions")
	rootCmd.PersistentFlags().StringSliceVar(&allowedProtocols, "allowed-protocols", []string{}, "Allowed protocols for transactions")
	// Bind flags to viper
	viper.BindPFlag("keypair", rootCmd.PersistentFlags().Lookup("keypair"))
	viper.BindPFlag("rpc-url", rootCmd.PersistentFlags().Lookup("rpc-url"))
	viper.BindPFlag("rpc-api-key", rootCmd.PersistentFlags().Lookup("rpc-api-key"))
	viper.BindPFlag("lulo-api-key", rootCmd.PersistentFlags().Lookup("lulo-api-key"))
	viper.BindPFlag("priority-fee", rootCmd.PersistentFlags().Lookup("priority-fee"))
	viper.BindPFlag("allowed-protocols", rootCmd.PersistentFlags().Lookup("allowed-protocols"))
}

func initConfig() {
	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Explicitly set the config file path
	configPath := filepath.Join(wd, "config.yaml")
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Read environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("GOLULO")

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file: %s\n", err)
		return
	}
}
