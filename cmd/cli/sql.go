/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xuperchain/xuperunion/utxo"
)

// SQLCommand run SQL on xchain
type SQLCommand struct {
	cli *Cli
	cmd *cobra.Command

	dbname   string
	contract string
}

// NewSQLCommand new SQLCommand
func NewSQLCommand(cli *Cli) *cobra.Command {
	c := new(SQLCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:     "sql [sql]",
		Short:   "run sql on xchain",
		Example: c.example(),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return c.run(args[0])
			}
			return c.runloop(os.Stdin)
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *SQLCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.dbname, "db", "d", "test", "db name")
	c.cmd.Flags().StringVarP(&c.contract, "cname", "c", "sql", "contract name")

}

func (c *SQLCommand) example() string {
	return `
xchain wasm sql $sql'
`
}

func (c *SQLCommand) runloop(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	fmt.Print("> ")
	var sql string
	for scanner.Scan() {
		txt := strings.TrimSpace(scanner.Text())
		if sql == "" {
			sql = txt
		} else {
			sql = sql + "\n" + txt
		}
		if !strings.HasSuffix(txt, ";") {
			fmt.Print(". ")
			continue
		}
		err := c.run(sql)
		if err != nil {
			fmt.Println(err)
		}
		sql = ""
		fmt.Print("> ")
	}
	return nil
}

func (c *SQLCommand) run(sql string) error {
	if strings.HasPrefix(sql, "select") || strings.HasPrefix(sql, "SELECT") {
		return c.query(sql)
	}
	return c.invoke(sql)
}

func (c *SQLCommand) query(sql string) error {
	ct := &CommTrans{
		Version:      utxo.TxVersion,
		ModuleName:   "sql",
		ContractName: c.contract,
		MethodName:   "dummy",
		ChainName:    c.cli.RootOptions.Name,
		Keys:         c.cli.RootOptions.Keys,
		XchainClient: c.cli.XchainClient(),
		CryptoType:   c.cli.RootOptions.CryptoType,
	}
	ct.Args = map[string][]byte{
		"db":  []byte(c.dbname),
		"sql": []byte(sql),
	}
	_, _, err := ct.GenPreExeRes(context.TODO())
	return err
}

func (c *SQLCommand) invoke(sql string) error {
	ct := &CommTrans{
		Fee:          "15000",
		FrozenHeight: 0,
		Version:      utxo.TxVersion,
		ModuleName:   "sql",
		ContractName: c.contract,
		MethodName:   "dummy",
		ChainName:    c.cli.RootOptions.Name,
		Keys:         c.cli.RootOptions.Keys,
		XchainClient: c.cli.XchainClient(),
		CryptoType:   c.cli.RootOptions.CryptoType,
	}

	ct.Args = map[string][]byte{
		"db":  []byte(c.dbname),
		"sql": []byte(sql),
	}

	return ct.Transfer(context.TODO())
}

func init() {
	AddCommand(NewSQLCommand)
}
