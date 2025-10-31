package main

import (
	"fmt"
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	minitiaapp "github.com/initia-labs/miniwasm/app"
)

func main() {
	rootCmd, _ := NewRootCmd()

	if err := svrcmd.Execute(rootCmd, minitiaapp.EnvPrefix, minitiaapp.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err) //nolint:errcheck
		os.Exit(1)
	}
}
