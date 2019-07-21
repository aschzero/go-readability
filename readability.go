// Package readability is a Go package that find the main readable
// content from a HTML page. It works by removing clutter like buttons,
// ads, background images, script, etc.
//
// This package is based from Readability.js by Mozilla, and written line
// by line to make sure it looks and works as similar as possible. This
// way, hopefully all web page that can be parsed by Readability.js
// are parse-able by go-readability as well.
package readability

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	nurl "net/url"
	"strings"
	"time"
)

// FromReader parses input from an `io.Reader` and returns the
// readable content. It's the wrapper for `Parser.Parse()` and useful
// if you only want to use the default parser.
func FromReader(input io.Reader, pageURL string) (Article, error) {
	parser := NewParser()
	return parser.Parse(input, pageURL)
}

// IsReadable decides whether or not the document is reader-able
// without parsing the whole thing. It's the wrapper for
// `Parser.IsReadable()` and useful if you only use the default parser.
func IsReadable(input io.Reader) bool {
	parser := NewParser()
	return parser.IsReadable(input)
}

// FromURL fetch the web page from specified url, check if it's
// readable, then parses the response to find the readable content.
func FromURL(pageURL string, timeout time.Duration) (Article, error) {
	// Make sure URL is valid
	_, err := nurl.ParseRequestURI(pageURL)
	if err != nil {
		return Article{}, fmt.Errorf("failed to parse URL: %v", err)
	}

	// Fetch page from URL
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return Article{}, fmt.Errorf("failed to request page: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0 Safari/605.1.15")

	resp, err := client.Do(req)
	if err != nil {
		return Article{}, fmt.Errorf("failed to fetch the page: %v", err)
	}
	defer resp.Body.Close()

	// Make sure content type is HTML
	cp := resp.Header.Get("Content-Type")
	if !strings.Contains(cp, "text/html") {
		return Article{}, fmt.Errorf("URL is not a HTML document")
	}

	// Check if the page is readable
	var buffer bytes.Buffer
	tee := io.TeeReader(resp.Body, &buffer)

	parser := NewParser()
	if !parser.IsReadable(tee) {
		return Article{}, fmt.Errorf("the page is not readable")
	}

	// Parse content
	return parser.Parse(&buffer, pageURL)
}
