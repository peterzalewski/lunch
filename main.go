package main

// TODO: Log actions
// TODO: Use cobra for verbose, cache-bust flags
// TODO: Store config in yaml and read with Viper

import (
	"lunch/cmd"
)

func main() {
	cmd.Execute()
}
