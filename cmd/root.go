package cmd

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "lunch",
	Short: "Interact with the NYS public school menu",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Open("config.yaml")
		if err != nil {
			return err
		}

		data, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		var config LunchConfig
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return err
		}

		cmd.SetContext(context.WithValue(cmd.Context(), lunchConfigKey{}, &config))

		return nil
	},
}

var (
	dateRe             = regexp.MustCompile(`^(?P<date>\S+ \d+; \d+\?\S+)\s*(?P<note>.+)?$`)
	excessWhitespaceRe = regexp.MustCompile(`(?m)\s{2,}`)
	ErrInvalidRecord   = "invalid lunch record"
)

type lunchConfigKey struct{}

type LunchConfig struct {
	SchoolYear string        `yaml:"schoolYear"`
	BasePath   string        `yaml:"basePath"`
	Options    []LunchOption `yaml:"options"`
}

type LunchOption struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

type DailyMenu struct {
	Date time.Time
	Note string
	Menu []string
}

func (lc LunchConfig) UrlForOption(lo LunchOption) string {
	now := time.Now()
	return fmt.Sprintf("%s/%s/%s/%s.csv", lc.BasePath, lc.SchoolYear, strings.ToLower(now.Month().String()), lo.Path)
}

func (lc LunchConfig) KeyForOption(lo LunchOption) string {
	now := time.Now()
	return fmt.Sprintf("%s-%s-%s.csv", lc.SchoolYear, strings.ToLower(now.Month().String()), lo.Path)
}

func pullCacheableUrl(url string, cacheKey string) (string, error) {
	var builder strings.Builder

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
		resp, err := http.Get(url)
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

// TODO: Should return []struct of some sort
func (lc LunchConfig) GetDailyMenu(lo LunchOption) ([]DailyMenu, error) {
	lunchUrl := lc.UrlForOption(lo)
	cacheKey := lc.KeyForOption(lo)
	data, err := pullCacheableUrl(lunchUrl, cacheKey)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(strings.NewReader(data))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	menus := make([]DailyMenu, 0)
	for _, record := range records[1:] {
		if record[0] == "Daily Offerings" {
			continue
		}

		menu, err := NewDailyMenu(record)
		if err != nil {
			return nil, err
		}

		menus = append(menus, *menu)
	}

	return menus, nil
}

func NewDailyMenu(record []string) (*DailyMenu, error) {
	daily := &DailyMenu{}
	match := dateRe.FindStringSubmatch(record[0])
	if match == nil {
		return nil, fmt.Errorf(`record "%s": %w`, record[0], ErrInvalidRecord)
	}

	date, err := time.Parse("January 2; 2006?Monday", match[dateRe.SubexpIndex("date")])
	if err != nil {
		return nil, err
	}

	menu := make([]string, 0)
	for _, entry := range strings.Split(record[1], "|") {
		entry = strings.TrimSpace(excessWhitespaceRe.ReplaceAllString(entry, " "))
		menu = append(menu, entry)
	}

	daily.Date = date
	daily.Menu = menu

	return daily, nil
}

func init() {
	rootCmd.AddCommand(printCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
