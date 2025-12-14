/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/frostyeti/kpv/cmd"

	_ "github.com/frostyeti/kpv/cmd/config"
	_ "github.com/frostyeti/kpv/cmd/config/aliases"
	_ "github.com/frostyeti/kpv/cmd/config/secret"
	_ "github.com/frostyeti/kpv/cmd/secrets"
)

func main() {
	cmd.Execute()
}
