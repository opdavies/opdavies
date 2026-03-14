package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mmcdole/gofeed"
)

const (
	readmeFile = "README.md"
)

func main() {
	updateLatestBlogPosts()
	updateLatestTestimonials()
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

	newSection := strings.Join(lines, "\n")

	updateSectionInReadme(startMarker, endMarker, newSection)

	fmt.Println("README.md updated with latest blog posts.")
}

func updateLatestTestimonials() {
	const (
		testimonialsDir = "testimonials"
		startMarker     = "<!-- Start latest testimonials -->"
		endMarker       = "<!-- End latest testimonials -->"
	)

	numToShow := 3

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
		name := ""
		desc := ""
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

		entry := fmt.Sprintf("%s\n\n%s", header, body)

		formatted = append(formatted, entry)
	}

	newSection := strings.Join(formatted, "\n\n")

	updateSectionInReadme(startMarker, endMarker, newSection)

	fmt.Println("README.md updated with latest testimonials.")
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

	err = ioutil.WriteFile(readmeFile, []byte(newReadme), 0644)

	if err != nil {
		fmt.Println("Error writing README.md:", err)

		return
	}
}
