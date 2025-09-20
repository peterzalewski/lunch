package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func printMonth(cmd *cobra.Command, args []string) error {
	config, ok := cmd.Context().Value(lunchConfigKey{}).(*LunchConfig)
	if !ok {
		return fmt.Errorf("could not retrieve config")
	}

	month, err := cmd.Flags().GetString("month")
	if err != nil {
		return err
	} else if month == "" {
		month = strings.ToLower(time.Now().Month().String())
	}

	option := config.Options[0]
	menus, err := config.GetDailyMenu(option, month)
	if err != nil {
		return err
	}

	for _, menu := range menus {
		fmt.Printf("%s: %s\n", menu.Date, strings.Join(menu.Menu, ", "))
	}

	return nil
}

func NewPrintCmd() *cobra.Command {
	var printCmd = &cobra.Command{
		Use:   "print",
		Short: "Print a month of NYS public school meals",
		RunE:  printMonth,
	}

	printCmd.Flags().StringP("month", "m", "", "Which month to print (defaults to current)")

	return printCmd
}
