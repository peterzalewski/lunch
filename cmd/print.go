package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var SchoolMonths = [...]string{"september", "october", "november", "december", "january", "february", "march", "april", "may", "june"}

func printMonth(cmd *cobra.Command, args []string) error {
	config, ok := cmd.Context().Value(lunchConfigKey{}).(*LunchConfig)
	if !ok {
		return fmt.Errorf("could not retrieve config")
	}

	monthIdx := int(time.Now().Month())
	if monthIdx >= 9 {
		monthIdx -= 9
	} else {
		monthIdx += 3
	}

	next, err := cmd.Flags().GetBool("next")
	if err != nil {
		return err
	}

	if next {
		if monthIdx >= len(SchoolMonths) {
			return fmt.Errorf("no school menu after june")
		}
		monthIdx += 1
	}

	month := SchoolMonths[monthIdx]

	var option LunchOption
	options := make([]huh.Option[LunchOption], 0)
	for _, o := range config.Options {
		options = append(options, huh.NewOption[LunchOption](o.Name, o))
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[LunchOption]().
				Description("Choose a menu").
				Options(options...).
				Height(10).
				Value(&option),
		),
	)
	err = form.Run()
	if err != nil {
		return err
	}

	menus, err := config.GetDailyMenu(option, month)
	if err != nil {
		return err
	}

	fmt.Printf("MENU for %s\n\n", strings.ToTitle(month))
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

	printCmd.Flags().BoolP("next", "n", false, "Print next month's menu")

	return printCmd
}
