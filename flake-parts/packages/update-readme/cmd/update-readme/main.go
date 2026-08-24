// Command update-readme refreshes the generated sections of README.md from
// oliverdavies.uk.
//
// That site is the canonical source. This repository holds no copy of the
// content it displays: it asks for /blog.json and /testimonials.json and
// rewrites the marked sections with what comes back.
package main

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	readmeFile     = "README.md"
	defaultBaseURL = "https://www.oliverdavies.uk"
	userAgent      = "opdavies-update-readme (+https://github.com/opdavies/opdavies)"

	numBlogPosts    = 10
	numTestimonials = 5
)

// baseURL allows the source site to be pointed elsewhere, so the program can
// be run against a local build before a change to the site is deployed.
func baseURL() string {
	if url := os.Getenv("SITE_URL"); url != "" {
		return strings.TrimSuffix(url, "/")
	}

	return defaultBaseURL
}

type blogFeed struct {
	Items []struct {
		Title         string `json:"title"`
		URL           string `json:"url"`
		DatePublished string `json:"date_published"`
	} `json:"items"`
}

// The testimonials endpoint returns a bare array: it is a list, not a feed.
type testimonial struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
	IsFeatured  bool   `json:"is_featured"`
}

func main() {
	if err := run(); err != nil {
		// Fail loudly. Quietly carrying on is how this stopped updating
		// without anyone noticing.
		fmt.Fprintln(os.Stderr, "update-readme:", err)
		os.Exit(1)
	}
}

func run() error {
	readme, err := os.ReadFile(readmeFile)
	if err != nil {
		return fmt.Errorf("reading %s: %w", readmeFile, err)
	}

	posts, err := latestBlogPosts()
	if err != nil {
		return err
	}

	testimonials, err := latestTestimonials()
	if err != nil {
		return err
	}

	content := string(readme)

	for _, section := range []struct{ name, body string }{
		{"latest blog posts", posts},
		{"latest testimonials", testimonials},
	} {
		content, err = replaceSection(content, section.name, section.body)
		if err != nil {
			return err
		}
	}

	if content == string(readme) {
		fmt.Println("README.md is already up to date.")

		return nil
	}

	if err := os.WriteFile(readmeFile, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", readmeFile, err)
	}

	fmt.Println("README.md updated.")

	return nil
}

func fetchJSON(url string, into any) error {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("building request for %s: %w", url, err)
	}

	// Say who this is. Go's default agent is indistinguishable from any
	// other script, which is no help to whoever is reading the logs or
	// deciding whether to let the request through.
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fetching %s: %w", url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetching %s: %s", url, resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(into); err != nil {
		return fmt.Errorf("decoding %s: %w", url, err)
	}

	return nil
}

func latestBlogPosts() (string, error) {
	url := baseURL() + "/blog.json"

	var feed blogFeed

	if err := fetchJSON(url, &feed); err != nil {
		return "", err
	}

	if len(feed.Items) == 0 {
		return "", fmt.Errorf("%s returned no posts", url)
	}

	items := feed.Items
	if len(items) > numBlogPosts {
		items = items[:numBlogPosts]
	}

	var lines []string

	for _, item := range items {
		date, err := time.Parse(time.RFC3339, item.DatePublished)
		if err != nil {
			return "", fmt.Errorf("parsing date for %q: %w", item.Title, err)
		}

		lines = append(lines, fmt.Sprintf("- [%s](%s) - %s", item.Title, item.URL, formatDate(date)))
	}

	return strings.Join(lines, "\n"), nil
}

func latestTestimonials() (string, error) {
	url := baseURL() + "/testimonials.json"

	var items []testimonial

	if err := fetchJSON(url, &items); err != nil {
		return "", err
	}

	if len(items) == 0 {
		return "", fmt.Errorf("%s returned no testimonials", url)
	}

	// The endpoint serves every testimonial and marks the featured ones.
	// Which to show is this program's decision, not the site's.
	featured := make([]testimonial, 0, len(items))

	for _, item := range items {
		if item.IsFeatured {
			featured = append(featured, item)
		}
	}

	if len(featured) == 0 {
		return "", fmt.Errorf("%s returned no featured testimonials", url)
	}

	items = featured

	if len(items) > numTestimonials {
		items = items[:numTestimonials]
	}

	var sections []string

	for _, item := range items {
		heading := "### " + item.Name

		if item.Description != "" {
			heading += " - " + item.Description
		}

		sections = append(sections, heading+"\n\n"+markdownFromHTML(item.Content))
	}

	return strings.Join(sections, "\n\n"), nil
}

var (
	linkPattern      = regexp.MustCompile(`(?is)<a\b[^>]*href="([^"]*)"[^>]*>(.*?)</a>`)
	paragraphPattern = regexp.MustCompile(`(?i)</p>`)
	tagPattern       = regexp.MustCompile(`(?s)<[^>]+>`)
	blankLinePattern = regexp.MustCompile(`\n{3,}`)
)

// markdownFromHTML turns the rendered HTML back into the Markdown a README
// wants. The input is Sculpin's output from Markdown, so it is paragraphs and
// the occasional link rather than arbitrary HTML.
func markdownFromHTML(in string) string {
	out := linkPattern.ReplaceAllString(in, "[$2]($1)")
	out = paragraphPattern.ReplaceAllString(out, "\n\n")
	out = tagPattern.ReplaceAllString(out, "")
	out = html.UnescapeString(out)
	out = blankLinePattern.ReplaceAllString(out, "\n\n")

	return strings.TrimSpace(out)
}

func replaceSection(content, name, body string) (string, error) {
	start := fmt.Sprintf("<!-- Start %s -->", name)
	end := fmt.Sprintf("<!-- End %s -->", name)

	startIdx := strings.Index(content, start)
	endIdx := strings.Index(content, end)

	if startIdx == -1 || endIdx == -1 || startIdx > endIdx {
		return "", fmt.Errorf("could not find markers %q and %q in %s", start, end, readmeFile)
	}

	return content[:startIdx+len(start)] + "\n\n" + body + "\n\n" + content[endIdx:], nil
}

func formatDate(t time.Time) string {
	return fmt.Sprintf("%s %s %d", ordinal(t.Day()), t.Month(), t.Year())
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
