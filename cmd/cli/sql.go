/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xuperchain/xuperunion/pb"
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
	fmt.Printf("XuperChain SQL client version %s\n", xchainVersion())
	fmt.Printf("Connected to xchain node, contract:%s database:%s\n", c.contract, c.dbname)
	scanner := bufio.NewScanner(r)
	prompt := "xsql> "
	fmt.Print(prompt)
	var sql string
	for scanner.Scan() {
		txt := strings.TrimSpace(scanner.Text())
		if sql == "" {
			sql = txt
		} else {
			sql = sql + "\n" + txt
		}
		if !strings.HasSuffix(txt, ";") {
			fmt.Print("   . ")
			continue
		}
		err := c.run(sql)
		if err != nil {
			fmt.Println(err)
		}
		sql = ""
		fmt.Print(prompt)
	}
	return nil
}

func (c *SQLCommand) run(sql string) error {
	if sql == "" {
		return nil
	}
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
	resps, reqs, err := ct.GenPreExeRes(context.TODO())
	if err != nil {
		return err
	}
	var resp *pb.ContractResponse
	for i := range reqs {
		if reqs[i].GetContractName() == c.contract {
			resp = resps.GetResponse().GetResponses()[i]
		}
	}
	if resp == nil {
		return nil
	}
	var result [][]interface{}
	err = json.Unmarshal(resp.GetBody(), &result)
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, row := range result {
		var rowslice []string
		for _, elem := range row {
			rowslice = append(rowslice, fmt.Sprintf("%v", elem))
		}
		fmt.Fprintf(w, "%s\n", strings.Join(rowslice, "\t"))
	}
	w.Flush()
	return nil
}

func (c *SQLCommand) invoke(sql string) error {
	ct := &CommTrans{
		AutoFee:      true,
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
