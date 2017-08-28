package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
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
	Pages       string         `json:"pages"`
}

type CourseLocation struct {
	City    string `json:"city"`
	Country string `json:"country"`
	URL     string `json:"website_url"`
}

var (
	studentType = flag.String("student", "old", "'old' or 'new'")
	region      = flag.String("region", "Europe", "region")
	from        = flag.String("from", "now", "from date YYYY-MM-DD")
	to          = flag.String("to", "", "to date")
)

var studentMap = map[string]string{
	"old": "OldStudent",
	"new": "NewStudent",
}

func init() {
	flag.Parse()
	if *from == "now" {
		*from = time.Now().Format("2006-01-02")
	}
	if *to == "" {
		t, err := time.Parse("2006-01-02", *from)
		if err != nil {
			log.Fatal(err)
		}
		*to = t.AddDate(1, 0, 0).Format("2006-01-02")
	}
}

func listCourses(courses []Course) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 5, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "Starts\tOpens\tCity\tCountry\tURL")
	for _, c := range courses {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", beautify(c.CourseStart),
			beautify(c.Opens), c.Location.City, c.Location.Country, c.Location.URL)
	}
	w.Flush()
}

func getPage(n int) ([]Course, int) {
	postData := url.Values{
		"current_state":  []string{studentMap[*studentType]},
		"regions[]":      []string{regionMap[*region]},
		"languages[]":    []string{"en"},
		"course_types[]": []string{"3"},
		"sort_column":    []string{"dates"},
		"sort_direction": []string{"up"},
		"date_format":    []string{"YYYY-MM-DD"},
		"daterange":      []string{fmt.Sprintf("%s - %s", *from, *to)},
		"page":           []string{strconv.Itoa(n)},
	}
	resp, err := http.PostForm(endpoint, postData)
	if err != nil {
		log.Fatal(err)
	}
	var out struct {
		Pages int      `json:"pages"`
		List  []Course `json:"courses"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
	return out.List, out.Pages
}

func beautify(yyyymmdd string) string {
	t, err := time.Parse("2006-01-02", yyyymmdd)
	if err != nil {
		log.Fatal(err)
	}
	return t.Format("02 Jan 2006")
}

// the current time.
func filter(list []Course) []Course {
	out := make([]Course, 0)
	after, err := time.Parse("2006-01-02", *from)
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range list {
		t, err := time.Parse("2006-01-02", c.Opens)
		if err != nil {
			log.Fatal(err)
		}
		if t.After(after) {
			out = append(out, c)
		}
	}
	return out
}

func main() {
	all := make([]Course, 0)
	list, pages := getPage(1)
	all = append(all, filter(list)...)
	if pages > 1 {
		for i := 2; i <= pages; i++ {
			list, _ := getPage(i)
			all = append(all, filter(list)...)
		}
	}
	listCourses(all)
}
