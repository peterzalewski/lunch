package cmd

import (
	"encoding/csv"
	"fmt"
	"strings"
	"time"

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

	data, err := config.DataFor(config.Options[0])
	if err != nil {
		return err
	}

	reader := csv.NewReader(strings.NewReader(data))
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	for _, record := range records[1:] {
		if record[0] == "Daily Offerings" {
			continue
		}

		match := dateRe.FindStringSubmatch(record[0])
		if match == nil {
			return fmt.Errorf(`record "%s": %w`, record[0], ErrInvalidRecord)
		}

		date, err := time.Parse("January 2; 2006?Monday", match[dateRe.SubexpIndex("date")])
		if err != nil {
			return err
		}

		menu := make([]string, 0)
		for _, entry := range strings.Split(record[1], "|") {
			entry = strings.TrimSpace(excessWhitespaceRe.ReplaceAllString(entry, " "))
			menu = append(menu, entry)
		}

		fmt.Printf("%s: %s\n", date, strings.Join(menu, ", "))
	}

	return nil
}
