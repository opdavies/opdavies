package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
)

const (
	readmeFile = "README.md"
)

func main() {
	updateLatestTestimonials()
}

func updateLatestTestimonials() {
	const (
		testimonialsDir = "testimonials"
		startMarker     = "<!-- Start latest testimonials -->"
		endMarker       = "<!-- End latest testimonials -->"
	)

	x := 3 // Number of latest testimonials to include

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

	if len(filenames) < x {
		x = len(filenames)
	}

	latest := filenames[len(filenames)-x:]

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

		// Detect YAML front matter correctly
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

	// Read README.md
	readmeContent, err := ioutil.ReadFile(readmeFile)

	if err != nil {
		fmt.Println("Error reading README.md:", err)
		return
	}

	contentStr := string(readmeContent)

	startIdx := strings.Index(contentStr, startMarker)
	endIdx := strings.Index(contentStr, endMarker)

	if startIdx == -1 || endIdx == -1 || startIdx > endIdx {
		fmt.Println("Could not find markers in README.md")

		return
	}

	newReadme := contentStr[:startIdx+len(startMarker)] + "\n\n" + newSection + "\n\n" + contentStr[endIdx:]

	// Write back to README.md
	err = ioutil.WriteFile(readmeFile, []byte(newReadme), 0644)

	if err != nil {
		fmt.Println("Error writing README.md:", err)

		return
	}

	fmt.Println("README.md updated with latest testimonials.")
}
