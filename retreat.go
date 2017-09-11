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

// endpoints holds the URL to retrieve information from.
const endpoint = "https://www.dhamma.org/en-US/courses/do_search"

// Course holds information about a Vipassana course.
type Course struct {
	ID         int    `json:"id"`
	CourseType string `json:"course_type"`
	Location   struct {
		City    string `json:"city"`
		Country string `json:"country"`
		URL     string `json:"website_url"`
	} `json:"location"`
	CourseStart string `json:"course_start_date"`
	Opens       string `json:"enrollment_open_date"`
	Pages       string `json:"pages"`
}

var (
	days     = flag.String("days", "10", "length in days")
	region   = flag.String("region", "Europe", "region")
	from     = flag.String("from", "now", "start date YYYY-MM-DD")
	to       = flag.String("to", "", "end date YYYY-MM-DD")
	noFilter = flag.Bool("all", false, "if set, results are unfiltered")
)

// filterFunc holds the filter function that will be used to filter the results.
var filterFunc = notYetOpen

var lengthMap = map[string]string{
	"1":  "5",
	"2":  "19",
	"3":  "9",
	"10": "3",
	"20": "4",
	"30": "11",
	"45": "12",
	"60": "23",
}

func init() {
	flag.Parse()
	log.SetPrefix("retreat: ")
	log.SetFlags(0)
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
	if *noFilter {
		filterFunc = all
	}
}

// printCourses prints all the courses nicely
func printCourses(courses []Course) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 5, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "Starts\tOpens\tCity\tCountry\tURL")
	for _, c := range courses {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", beautify(c.CourseStart),
			beautify(c.Opens), c.Location.City, c.Location.Country, c.Location.URL)
	}
	w.Flush()
}

// postDataForPage returns post data used to retrieve page number n.
func postDataForPage(n int) url.Values {
	r, ok := regionMap[*region]
	if !ok {
		log.Fatal("region not found")
	}
	l, ok := lengthMap[*days]
	if !ok {
		log.Fatal("can only search for courses of length 1, 2, 3, 10, 20, 30, 45 and 60 days")
	}
	return url.Values{
		"current_state":  []string{"OldStudent"},
		"regions[]":      []string{r},
		"languages[]":    []string{"en"},
		"course_types[]": []string{l},
		"sort_column":    []string{"dates"},
		"sort_direction": []string{"up"},
		"date_format":    []string{"YYYY-MM-DD"},
		"daterange":      []string{fmt.Sprintf("%s - %s", *from, *to)},
		"page":           []string{strconv.Itoa(n)},
	}
}

// getPage creates an HTTP request and returns all courses on page n, as well
// as the total number of pages.
func getPage(n int) ([]Course, int) {
	resp, err := http.PostForm(endpoint, postDataForPage(n))
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

// notYetOpen filters the given list and returns only the courses that have their
// enrollment dates after the start date.
func notYetOpen(list []Course) []Course {
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

// all is a filter that returns the entire list.
func all(list []Course) []Course { return list }

func main() {
	all := make([]Course, 0)
	list, pages := getPage(1)
	all = append(all, list...)
	if pages > 1 {
		for i := 2; i <= pages; i++ {
			list, _ := getPage(i)
			all = append(all, list...)
		}
	}
	printCourses(filterFunc(all))
}
