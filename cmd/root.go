package cmd

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "lunch",
	Short: "Display the NYS public school menu",
	RunE:  root,
}

var (
	dateRe             = regexp.MustCompile(`^(?P<date>\S+ \d+; \d+\?\S+)\s*(?P<note>.+)?$`)
	excessWhitespaceRe = regexp.MustCompile(`(?m)\s{2,}`)
)

type LunchConfig struct {
	schoolYear string
	basePath   string
	options    []LunchOption
}

type LunchOption struct {
	name string
	path string
}

type DailyMenu struct {
	date string
	note string
	menu []string
}

func (lc LunchConfig) UrlForOption(lo LunchOption) string {
	now := time.Now()
	return fmt.Sprintf("%s/%s/%s/%s.csv", lc.basePath, lc.schoolYear, strings.ToLower(now.Month().String()), lo.path)
}

func (lc LunchConfig) KeyForOption(lo LunchOption) string {
	now := time.Now()
	return fmt.Sprintf("%s-%s-%s.csv", lc.schoolYear, strings.ToLower(now.Month().String()), lo.path)
}

// TODO: Should return []struct of some sort
func (lc LunchConfig) DataFor(lo LunchOption) (string, error) {
	var builder strings.Builder

	// Get cache key
	// If it exists in the cache, return
	// Else download the file, cache, return
	cacheKey := lc.KeyForOption(lo)
	cacheFile, err := xdg.CacheFile(fmt.Sprintf("lunch/%s", cacheKey))
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(cacheFile); err == nil {
		// TODO: Read and return
		cache, err := os.Open(cacheFile)
		if err != nil {
			return "", err
		}
		reader := bufio.NewScanner(cache)
		for reader.Scan() {
			// TODO: Here (and elsewhere) don't slurp the newline and then add it back in such a heavyweight way
			builder.WriteString(fmt.Sprintf("%s\n", reader.Text()))
		}
	} else {
		// TODO: Should probably be its own function
		lunchUrl := lc.UrlForOption(lo)
		resp, err := http.Get(lunchUrl)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		cache, err := os.Create(cacheFile)
		if err != nil {
			return "", err
		}

		reader := bufio.NewScanner(resp.Body)
		writer := bufio.NewWriter(cache)
		defer writer.Flush()

		for reader.Scan() {
			line := reader.Text()
			builder.WriteString(fmt.Sprintf("%s\n", line))
			writer.WriteString(fmt.Sprintf("%s\n", line))
		}
	}

	return builder.String(), nil
}

func root(cmd *cobra.Command, args []string) error {
	highSchoolCold := LunchOption{
		name: "High School Express Cold Lunch Menu",
		path: "High-School-Express-Cold-Lunch-Menu",
	}
	preK8ExpressCold := LunchOption{
		name: "Pre-K - 8 Express Cold Lunch Menu",
		path: "Pre-K---8-Express-Cold-Lunch-Menu",
	}
	config := &LunchConfig{
		schoolYear: "2025-2026",
		basePath:   "https://www.schools.nyc.gov/docs/default-source/school-menus/",
		options:    []LunchOption{highSchoolCold, preK8ExpressCold},
	}

	data, err := config.DataFor(preK8ExpressCold)
	if err != nil {
		panic(err)
	}

	reader := csv.NewReader(strings.NewReader(data))
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	for _, record := range records[1:] {
		if record[0] == "Daily Offerings" {
			continue
		}

		match := dateRe.FindStringSubmatch(record[0])
		if match == nil {
			panic(fmt.Sprintf("invalid record: %s", record[0]))
		}

		date, err := time.Parse("January 2; 2006?Monday", match[dateRe.SubexpIndex("date")])
		if err != nil {
			panic(err)
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

func Execute() error {
	return rootCmd.Execute()
}
