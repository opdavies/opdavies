package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
	"golang.org/x/text/unicode/norm"
)

type Testimonial struct {
	Entity string `yaml:"Entity"`
	Name   string `yaml:"Name"`
	Role   string `yaml:"Role"`
	Text   string `yaml:"Text"`
	URL    string `yaml:"URL"`
}

func main() {
	data, err := ioutil.ReadFile("bin/testimonials.yaml")
	if err != nil {
		panic(err)
	}

	var testimonials []Testimonial
	err = yaml.Unmarshal(data, &testimonials)
	if err != nil {
		panic(err)
	}

	outputDir := "testimonials"
	os.MkdirAll(outputDir, os.ModePerm)

	for i, t := range testimonials {
		slug := slugify(t.Name)
		filename := fmt.Sprintf("%03d-%s.md", i+1, slug)
		path := filepath.Join(outputDir, filename)

		var builder strings.Builder
		builder.WriteString("---\n")
		builder.WriteString(fmt.Sprintf("name: %s\n", t.Name))
		if t.Role != "" {
			builder.WriteString(fmt.Sprintf("description: %s", t.Role))
			if t.Entity != "" {
				builder.WriteString(fmt.Sprintf(", %s", t.Entity))
			}
			builder.WriteString("\n")
		} else if t.Entity != "" {
			builder.WriteString(fmt.Sprintf("description: %s\n", t.Entity))
		}
		if t.URL != "" {
			builder.WriteString(fmt.Sprintf("url: %s\n", t.URL))
		}
		builder.WriteString("---\n\n")
		builder.WriteString(t.Text)
		builder.WriteString("\n")

		err = ioutil.WriteFile(path, []byte(builder.String()), 0644)
		if err != nil {
			panic(err)
		}

		fmt.Println("Created:", filename)
	}
}

func slugify(s string) string {
	t := norm.NFD.String(s)
	var b strings.Builder
	for _, r := range t {
		if unicode.Is(unicode.Mn, r) { // remove accents
			continue
		}
		b.WriteRune(r)
	}
	str := strings.ToLower(b.String())
	reg := regexp.MustCompile(`[^\w\s-]`)
	str = reg.ReplaceAllString(str, "")
	str = strings.ReplaceAll(str, " ", "-")
	return str
}
