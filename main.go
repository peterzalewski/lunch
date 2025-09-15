package main

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
	"time"
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

	lunchUrl := config.UrlForOption(config.options[0])
	resp, err := http.Get(lunchUrl)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}
