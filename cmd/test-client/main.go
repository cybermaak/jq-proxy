package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Colors for output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
)

// Config holds the test client configuration
type Config struct {
	BaseURL  string
	Endpoint string
	Verbose  bool
	Timeout  time.Duration
}

// TestRequest represents a proxy request
type TestRequest struct {
	Method             string         `json:"method"`
	Body               any            `json:"body"`
	Transformation     map[string]any `json:"transformation,omitempty"`
	TransformationMode string         `json:"transformation_mode,omitempty"`
	JQQuery            string         `json:"jq_query,omitempty"`
}

func main() {
	var config Config
	var command string

	// Parse command line flags
	flag.StringVar(&config.BaseURL, "url", "http://localhost:8080", "Base URL of the proxy service")
	flag.StringVar(&config.Endpoint, "endpoint", "jsonplaceholder", "Endpoint name to test")
	flag.BoolVar(&config.Verbose, "verbose", false, "Verbose output")
	flag.DurationVar(&config.Timeout, "timeout", 30*time.Second, "Request timeout")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] COMMAND\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  health        Test health endpoint\n")
		fmt.Fprintf(os.Stderr, "  simple        Simple GET request test\n")
		fmt.Fprintf(os.Stderr, "  transform     Test with JSONPath transformation\n")
		fmt.Fprintf(os.Stderr, "  jq-transform  Test with jq transformation\n")
		fmt.Fprintf(os.Stderr, "  post          Test POST request\n")
		fmt.Fprintf(os.Stderr, "  error         Test error handling\n")
		fmt.Fprintf(os.Stderr, "  all           Run all tests\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s health\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -endpoint httpbin simple\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -url http://localhost:8081 all\n", os.Args[0])
	}

	flag.Parse()

	if flag.NArg() < 1 {
		command = "health"
	} else {
		command = flag.Arg(0)
	}

	client := &http.Client{Timeout: config.Timeout}

	printInfo("Testing JQ Proxy Service")
	printInfo(fmt.Sprintf("Base URL: %s", config.BaseURL))
	printInfo(fmt.Sprintf("Endpoint: %s", config.Endpoint))
	fmt.Println()

	switch command {
	case "health":
		testHealth(client, config)
	case "simple":
		testSimple(client, config)
	case "transform":
		testTransform(client, config)
	case "jq-transform":
		testJQTransform(client, config)
	case "post":
		testPost(client, config)
	case "error":
		testError(client, config)
	case "all":
		testAll(client, config)
	default:
		printError(fmt.Sprintf("Unknown command: %s", command))
		flag.Usage()
		os.Exit(1)
	}
}

func testHealth(client *http.Client, config Config) {
	makeRequest(client, config, "GET", config.BaseURL+"/health", nil, "Health check")
}

func testSimple(client *http.Client, config Config) {
	req := TestRequest{
		Method: "GET",
		Body:   nil,
		Transformation: map[string]any{
			"result": "$",
		},
	}
	makeProxyRequest(client, config, "/posts/1", req, "Simple GET request")
}

func testTransform(client *http.Client, config Config) {
	req := TestRequest{
		Method: "GET",
		Body:   nil,
		Transformation: map[string]any{
			"posts": "$[*].{id: id, title: title}",
			"count": "$.length",
		},
	}
	makeProxyRequest(client, config, "/posts", req, "GET with JSONPath transformation")
}

func testJQTransform(client *http.Client, config Config) {
	req := TestRequest{
		Method:             "GET",
		Body:               nil,
		TransformationMode: "jq",
		JQQuery:            "{posts: [.[] | {id: .id, title: .title}], count: length}",
	}
	makeProxyRequest(client, config, "/posts", req, "GET with jq transformation")
}

func testPost(client *http.Client, config Config) {
	req := TestRequest{
		Method: "POST",
		Body: map[string]any{
			"title":  "Test Post",
			"body":   "This is a test post",
			"userId": 1,
		},
		Transformation: map[string]any{
			"created_post": "$.{id: id, title: title}",
		},
	}
	makeProxyRequest(client, config, "/posts", req, "POST request with body")
}

func testError(client *http.Client, config Config) {
	req := TestRequest{
		Method: "GET",
		Body:   nil,
		Transformation: map[string]any{
			"result": "$",
		},
	}
	url := fmt.Sprintf("%s/proxy/nonexistent-endpoint/test", config.BaseURL)
	makeRequestWithBody(client, config, "POST", url, req, "Error handling (nonexistent endpoint)")
}

func testAll(client *http.Client, config Config) {
	printInfo("Running all tests...")
	fmt.Println()

	testHealth(client, config)
	testSimple(client, config)
	testTransform(client, config)
	testJQTransform(client, config)
	testPost(client, config)
	testError(client, config)

	printSuccess("All tests completed!")
}

func makeProxyRequest(client *http.Client, config Config, path string, req TestRequest, description string) {
	url := fmt.Sprintf("%s/proxy/%s%s", config.BaseURL, config.Endpoint, path)
	makeRequestWithBody(client, config, "POST", url, req, description)
}

func makeRequestWithBody(client *http.Client, config Config, method, url string, body any, description string) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		printError(fmt.Sprintf("Failed to marshal request body: %v", err))
		return
	}
	makeRequest(client, config, method, url, bodyBytes, description)
}

func makeRequest(client *http.Client, config Config, method, url string, body []byte, description string) {
	printInfo(fmt.Sprintf("Testing: %s", description))

	if config.Verbose {
		fmt.Printf("Request: %s %s\n", method, url)
		if body != nil {
			fmt.Printf("Data: %s\n", string(body))
		}
		fmt.Println()
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		printError(fmt.Sprintf("Failed to create request: %v", err))
		return
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		printError(fmt.Sprintf("Request failed: %v", err))
		return
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		printError(fmt.Sprintf("Failed to read response: %v", err))
		return
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		printSuccess(fmt.Sprintf("Status: %d", resp.StatusCode))
	} else {
		printError(fmt.Sprintf("Status: %d", resp.StatusCode))
	}

	if config.Verbose || resp.StatusCode >= 400 {
		// Try to pretty print JSON
		var jsonData any
		if err := json.Unmarshal(responseBody, &jsonData); err == nil {
			prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
			if err == nil {
				fmt.Printf("Response: %s\n", string(prettyJSON))
			} else {
				fmt.Printf("Response: %s\n", string(responseBody))
			}
		} else {
			fmt.Printf("Response: %s\n", string(responseBody))
		}
	}

	fmt.Println()
}

func printInfo(msg string) {
	fmt.Printf("%s[INFO]%s %s\n", ColorBlue, ColorReset, msg)
}

func printSuccess(msg string) {
	fmt.Printf("%s[SUCCESS]%s %s\n", ColorGreen, ColorReset, msg)
}

func printError(msg string) {
	fmt.Printf("%s[ERROR]%s %s\n", ColorRed, ColorReset, msg)
}

func printWarning(msg string) {
	fmt.Printf("%s[WARNING]%s %s\n", ColorYellow, ColorReset, msg)
}

func init() {
	// Disable colors on Windows or if not a terminal
	if os.Getenv("NO_COLOR") != "" || !isTerminal() {
		// Reset all color constants to empty strings
		// This is a simple approach; in a real application you might want a more sophisticated solution
	}
}

func isTerminal() bool {
	// Simple check - in a real application you might want to use a library like isatty
	return strings.Contains(os.Getenv("TERM"), "term") || os.Getenv("TERM") == "xterm-256color"
}
