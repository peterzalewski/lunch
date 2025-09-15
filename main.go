package main

// TODO: Log actions
// TODO: Use cobra for verbose, cache-bust flags
// TODO: Store config in yaml and read with Viper

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/adrg/xdg"
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

func main() {
	highSchoolColdLunch := LunchOption{
		name: "High School Express Cold Lunch Menu",
		path: "High-School-Express-Cold-Lunch-Menu",
	}
	config := &LunchConfig{
		schoolYear: "2025-2026",
		basePath:   "https://www.schools.nyc.gov/docs/default-source/school-menus/",
		options:    []LunchOption{highSchoolColdLunch},
	}

	fmt.Println(config.DataFor(highSchoolColdLunch))
}
