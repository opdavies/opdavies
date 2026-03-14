package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

const readmeFile = "README.md"

func cleanContributionTitle(title string) string {
	prefix := "opdavies "

	if strings.HasPrefix(strings.ToLower(title), prefix) {
		title = title[len(prefix):]
	}

	if len(title) > 0 {
		title = strings.ToUpper(title[:1]) + title[1:]
	}

	return title
}

func isContribution(title, feedURL string) bool {
	titleLower := strings.ToLower(title)

	keywords := []string{"commit", "pushed", "pull request", "opened issue", "patch", "merge request"}

	for _, k := range keywords {
		if strings.Contains(titleLower, "at opdavies/opdavies") {
			continue
		}

		if strings.Contains(titleLower, "pushed to opdavies") {
			continue
		}

		if strings.Contains(titleLower, k) {
			return true
		}
	}

	return false
}

func main() {
	updateLatestBlogPosts()
	updateLatestTestimonials()
	updateLatestContributions()
}

func normalizeContributionTitle(title string) string {
	lower := strings.ToLower(title)

	// Fix "pushed repo" → "pushed to repo"
	if strings.Contains(lower, " pushed ") && !strings.Contains(lower, " pushed to ") {
		parts := strings.SplitN(title, " pushed ", 2)

		if len(parts) == 2 {
			return parts[0] + " pushed to " + parts[1]
		}
	}

	return title
}

func ordinal(day int) string {
	if day >= 11 && day <= 13 {
		return fmt.Sprintf("%dth", day)
	}

	switch day % 10 {
	case 1:
		return fmt.Sprintf("%dst", day)
	case 2:
		return fmt.Sprintf("%dnd", day)
	case 3:
		return fmt.Sprintf("%drd", day)
	default:
		return fmt.Sprintf("%dth", day)
	}
}

func updateLatestBlogPosts() {
	const (
		blogFeedURL = "https://www.oliverdavies.uk/rss/blog.xml"
		startMarker = "<!-- Start latest blog posts -->"
		endMarker   = "<!-- End latest blog posts -->"
		numToShow   = 5
	)

	resp, err := http.Get(blogFeedURL)

	if err != nil {
		fmt.Println("Error fetching RSS feed:", err)

		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("Error reading RSS feed:", err)

		return
	}

	fp := gofeed.NewParser()
	feed, err := fp.ParseString(string(body))

	if err != nil {
		fmt.Println("Error parsing RSS feed:", err)

		return
	}

	// Sort items by published date descending
	sort.Slice(feed.Items, func(i, j int) bool {
		ti, tj := feed.Items[i].PublishedParsed, feed.Items[j].PublishedParsed

		if ti == nil || tj == nil {
			return false
		}

		return ti.After(*tj)
	})

	limit := numToShow

	if len(feed.Items) < limit {
		limit = len(feed.Items)
	}

	var lines []string

	for _, item := range feed.Items[:limit] {
		t := item.PublishedParsed
		dateStr := fmt.Sprintf("%s %s %d", ordinal(t.Day()), t.Month(), t.Year())
		lines = append(lines, fmt.Sprintf("- [%s](%s) - %s", item.Title, item.Link, dateStr))
	}

	updateSectionInReadme(startMarker, endMarker, strings.Join(lines, "\n"))

	fmt.Println("README.md updated with latest blog posts.")
}

func updateLatestTestimonials() {
	const (
		testimonialsDir = "testimonials"
		startMarker     = "<!-- Start latest testimonials -->"
		endMarker       = "<!-- End latest testimonials -->"
	)

	numToShow := 5

	files, err := ioutil.ReadDir(testimonialsDir)

	if err != nil {
		fmt.Println("Error reading testimonials directory:", err)

		return
	}

	var filenames []string

	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") {
			filenames = append(filenames, f.Name())
		}
	}

	sort.Strings(filenames)

	if len(filenames) < numToShow {
		numToShow = len(filenames)
	}

	latest := filenames[len(filenames)-numToShow:]

	var formatted []string

	// Reverse order so newest appears first
	for i := len(latest) - 1; i >= 0; i-- {
		file := latest[i]

		content, err := ioutil.ReadFile(filepath.Join(testimonialsDir, file))

		if err != nil {
			fmt.Println("Error reading file:", file, err)

			continue
		}

		lines := strings.Split(string(content), "\n")
		name, desc := "", ""
		bodyStart := 0

		// Detect YAML front matter
		if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
			yamlEnd := -1

			for j := 1; j < len(lines); j++ {
				if strings.TrimSpace(lines[j]) == "---" {
					yamlEnd = j
					break
				}

				line := strings.TrimSpace(lines[j])

				if strings.HasPrefix(line, "name:") {
					name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
				}

				if strings.HasPrefix(line, "description:") {
					desc = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
				}
			}

			if yamlEnd != -1 {
				bodyStart = yamlEnd + 1
			}
		}

		body := strings.Join(lines[bodyStart:], "\n")
		body = strings.TrimSpace(body)

		header := "### " + name
		if desc != "" {
			header += " - " + desc
		}

		formatted = append(formatted, fmt.Sprintf("%s\n\n%s", header, body))
	}

	updateSectionInReadme(startMarker, endMarker, strings.Join(formatted, "\n\n"))

	fmt.Println("README.md updated with latest testimonials.")
}

func updateLatestContributions() {
	const (
		startMarker = "<!-- Start latest contributions -->"
		endMarker   = "<!-- End latest contributions -->"
		numToShow   = 20
	)

	feeds := []string{
		"https://git.drupalcode.org/opdavies.atom",
		"https://git.oliverdavies.uk/opdavies.atom",
		"https://github.com/opdavies.atom",
	}

	type Contribution struct {
		Date  time.Time
		Link  string
		Title string
	}

	var all []Contribution

	fp := gofeed.NewParser()

	for _, url := range feeds {
		resp, err := http.Get(url)

		if err != nil {
			fmt.Println("Error fetching feed:", url, err)

			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			fmt.Println("Error reading feed:", url, err)

			continue
		}

		feed, err := fp.ParseString(string(body))

		if err != nil {
			fmt.Println("Error parsing feed:", url, err)

			continue
		}

		for _, item := range feed.Items {
			// Skip if no date
			if item.PublishedParsed == nil {
				continue
			}

			title := normalizeContributionTitle(item.Title)
			title = cleanContributionTitle(title)

			titleLower := strings.ToLower(title)
			date := *item.PublishedParsed

			// Skip issue activity
			if strings.Contains(titleLower, "opened issue") {
				continue
			}

			if isContribution(title, url) {
				all = append(all, Contribution{
					Date:  date,
					Link:  item.Link,
					Title: title,
				})
			}
		}
	}

	// Sort descending by date
	sort.Slice(all, func(i, j int) bool { return all[i].Date.After(all[j].Date) })

	limit := numToShow

	if len(all) < limit {
		limit = len(all)
	}

	var lines []string

	for _, e := range all[:limit] {
		dateStr := fmt.Sprintf("%s %s %d", ordinal(e.Date.Day()), e.Date.Month(), e.Date.Year())
		lines = append(lines, fmt.Sprintf("- [%s](%s) - %s", e.Title, e.Link, dateStr))
	}

	updateSectionInReadme(startMarker, endMarker, strings.Join(lines, "\n"))

	fmt.Println("README.md updated with recent contributions.")
}

func updateSectionInReadme(startMarker, endMarker, newSection string) {
	readmeContent, err := ioutil.ReadFile(readmeFile)

	if err != nil {
		fmt.Println("Error reading README.md:", err)

		return
	}

	contentStr := string(readmeContent)

	startIdx := strings.Index(contentStr, startMarker)
	endIdx := strings.Index(contentStr, endMarker)

	if startIdx == -1 || endIdx == -1 || startIdx > endIdx {
		fmt.Printf("Could not find markers: %s ... %s\n", startMarker, endMarker)

		return
	}

	newReadme := contentStr[:startIdx+len(startMarker)] + "\n\n" + newSection + "\n\n" + contentStr[endIdx:]

	if err := ioutil.WriteFile(readmeFile, []byte(newReadme), 0644); err != nil {
		fmt.Println("Error writing README.md:", err)

		return
	}
}
