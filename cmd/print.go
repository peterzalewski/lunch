package cmd

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/spf13/cobra"
)

var printCmd = &cobra.Command{
	Use:   "print",
	Short: "Print a month of NYS public school meals",
	RunE:  printMonth,
}

func printMonth(cmd *cobra.Command, args []string) error {
	config, ok := cmd.Context().Value(lunchConfigKey{}).(*LunchConfig)
	if !ok {
		return fmt.Errorf("could not retrieve config")
	}

	option := config.Options[rand.IntN(len(config.Options))]
	menus, err := config.GetDailyMenu(option)
	if err != nil {
		return err
	}

	for _, menu := range menus {
		fmt.Printf("%s: %s\n", menu.Date, strings.Join(menu.Menu, ", "))
	}

	return nil
}
