package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Current configuration:\n")
		fmt.Printf("RPC URL: %s\n", viper.GetString("rpc-url"))
		if keypair := viper.GetString("keypair"); keypair != "" {
			fmt.Printf("Keypair: %s\n", keypair)
		}
		fmt.Printf("RPC API Key: %s\n", viper.GetString("rpc-api-key"))
		fmt.Printf("Lulo API Key: %s\n", viper.GetString("lulo-api-key"))
		fmt.Printf("Priority Fee: %s\n", viper.GetString("priority-fee"))
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a config value",
	Run: func(cmd *cobra.Command, args []string) {
		viper.Set("keypair", keypairPath)
		viper.Set("rpc-url", rpcURL)

		err := viper.WriteConfig()
		if err != nil {
			if err = viper.SafeWriteConfig(); err != nil {
				fmt.Printf("Error writing config: %s\n", err)
				return
			}
		}
		fmt.Println("Configuration updated successfully")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
}
