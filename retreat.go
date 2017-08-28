package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"text/tabwriter"
	"time"
)

const endpoint = "https://www.dhamma.org/en-US/courses/do_search"

type Course struct {
	ID          int            `json:"id"`
	CourseType  string         `json:"course_type"`
	Location    CourseLocation `json:"location"`
	CourseStart string         `json:"course_start_date"`
	Opens       string         `json:"enrollment_open_date"`
}

type CourseLocation struct {
	City    string `json:"city"`
	Country string `json:"country"`
	URL     string `json:"website_url"`
}

var (
	studentType = flag.String("student", "old", "'old' or 'new'")
	region      = flag.String("region", "eu", "region")
	from        = flag.String("from", "now", "from date DD-MM-YYYY")
	to          = flag.String("to", "", "to date")
)

var (
	studentMap = map[string]string{
		"old": "OldStudent",
		"new": "NewStudent",
	}
	regionMap = map[string]string{
		"eu": "region_117",
	}
)

func listCourses(courses []Course) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 5, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "Starts\tOpens\tCity\tCountry\tURL")
	for _, c := range courses {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", c.CourseStart, c.Opens, c.Location.City, c.Location.Country, c.Location.URL)
	}
	w.Flush()
}

func main() {
	flag.Parse()

	if *from == "now" {
		*from = time.Now().Format("02-01-2006")
	}
	if *to == "" {
		t, err := time.Parse("02-01-2006", *from)
		if err != nil {
			log.Fatal(err)
		}
		*to = t.AddDate(1, 0, 0).Format("02-01-2006")
	}
	postData := url.Values{
		"current_state":  []string{studentMap[*studentType]},
		"regions[]":      []string{regionMap[*region]},
		"languages[]":    []string{"en"},
		"sort_column":    []string{"dates"},
		"sort_direction": []string{"up"},
		"date_format":    []string{"DD-MM-YYYY"},
		"date_range":     []string{fmt.Sprintf("%s+-+%s", *from, *to)},
	}
	resp, err := http.PostForm(endpoint, postData)
	if err != nil {
		log.Fatal(err)
	}
	var out struct {
		List []Course `json:"courses"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	listCourses(out.List)
}
