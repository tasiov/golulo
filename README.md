# Golulo CLI

A command line interface for interacting with the Lulo Protocol on the Solana blockchain. This tool enables users to manage lending positions, view market data, and perform other protocol-related operations directly from the terminal.

## Features

- Manage lending positions
- View market data and account information
- Deposit and withdraw tokens from Lulo reserves
- Configure priority fees and RPC settings
- Support for multiple protocols

## Installation

[Installation instructions to be added]

## Usage

```bash
golulo [command]
```

### Available Commands

- `account` - Get account information
- `completion` - Generate the autocompletion script for the specified shell
- `config` - Manage CLI configuration
- `deposit` - Deposit tokens into a Lulo reserve
- `help` - Help about any command
- `pubkey` - Display public key from keypair file
- `version` - Print the version number
- `withdraw` - Withdraw tokens from a Lulo reserve

### Global Flags

```
--allowed-protocols strings   Allowed protocols for transactions
--config string              Config file (default is ./config.yaml)
--keypair string             Path to keypair file
--lulo-api-key string        API key for Lulo
--priority-fee string        Priority fee for transactions
--rpc-api-key string         API key for RPC
--rpc-url string             RPC server URL
-h, --help                   Help for golulo
```

## Configuration

The CLI can be configured using a YAML file. By default, it looks for `config.yaml` in the current directory. You can specify a different configuration file using the `--config` flag.

### Example Configuration

```yaml
keypair: /path/to/keypair.json
rpc-url: https://your-rpc-endpoint
rpc-api-key: your-rpc-api-key
lulo-api-key: your-lulo-api-key
allowed-protocols:
  - protocol1
  - protocol2
priority-fee: 5000
```

## Getting Help

To get more information about any command, use:

```bash
golulo [command] --help
```

## License

[License information to be added]

## Contributing

[Contributing guidelines to be added]
