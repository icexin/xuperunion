package main

import "github.com/spf13/cobra"

// EvmCommand manage evm contract
type EvmCommand struct {
}

// NewEvmCommand new evm cmd
func NewEvmCommand(cli *Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evm",
		Short: "Operate a command with evm, deploy|invoke|query",
	}
	cmd.AddCommand(NewContractDeployCommand(cli, "evm"))
	cmd.AddCommand(NewContractInvokeCommand(cli, "evm"))
	cmd.AddCommand(NewContractQueryCommand(cli, "evm"))
	cmd.AddCommand(NewContractUpgradeCommand(cli))
	return cmd
}

func init() {
	AddCommand(NewEvmCommand)
}
