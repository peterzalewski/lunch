package main

import (
	"bufio"
	"fmt"
	"net/http"
)

const (
	LunchURL = "https://www.schools.nyc.gov/school-life/food/menus/school-lunch-meals"
)

func main()  {
	resp, err := http.Get(LunchURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}
