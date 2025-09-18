package cmd

import (
	"bufio"
	"context"
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
	date string
	note string
	menu []string
}

func (lc LunchConfig) UrlForOption(lo LunchOption) string {
	now := time.Now()
	return fmt.Sprintf("%s/%s/%s/%s.csv", lc.BasePath, lc.SchoolYear, strings.ToLower(now.Month().String()), lo.Path)
}

func (lc LunchConfig) KeyForOption(lo LunchOption) string {
	now := time.Now()
	return fmt.Sprintf("%s-%s-%s.csv", lc.SchoolYear, strings.ToLower(now.Month().String()), lo.Path)
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

func init() {
	rootCmd.AddCommand(printCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
