package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexisbouchez/phpgo/interpreter"
	"github.com/alexisbouchez/phpgo/runtime"
)

func main() {
	port := "8080"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is for a PHP file
		if strings.HasSuffix(r.URL.Path, ".php") {
			handlePHPRequest(w, r)
			return
		}

		// Try to serve the file directly
		filePath := "." + r.URL.Path
		// Special case: serve index.html for root path
		if r.URL.Path == "/" {
			filePath = "./index.html"
		}
		if _, err := os.Stat(filePath); err == nil {
			http.ServeFile(w, r, filePath)
			return
		}

		// File not found
		http.NotFound(w, r)
	})

	fmt.Printf("PHPGo Development Server running on http://localhost:%s\n", port)
	fmt.Println("Press Ctrl+C to stop")
	fmt.Printf("Serving PHP files from: %s\n", filepath.Dir("."))

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func handlePHPRequest(w http.ResponseWriter, r *http.Request) {
	// Get the PHP file path
	filePath := "." + r.URL.Path
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// Read the PHP file
	phpCode, err := ioutil.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Error reading PHP file", http.StatusInternalServerError)
		return
	}

	// Create a new PHPGo interpreter
	i := interpreter.New()

	// Set up HTTP context - SetHTTPContext will populate superglobals automatically

	// Convert headers to simple map
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Convert cookies to simple map
	cookies := make(map[string]string)
	for _, cookie := range r.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}

	// Parse POST data
	postData := make(map[string]string)
	if r.Method == "POST" {
		r.ParseForm()
		for key, values := range r.PostForm {
			if len(values) > 0 {
				postData[key] = values[0]
			}
		}
	}

	// Set HTTP context - this automatically populates superglobals
	i.SetHTTPContext(
		r.Method,
		r.RequestURI,
		r.URL.RawQuery,
		headers,
		cookies,
		postData,
		make(map[string][]byte), // No file uploads for now
	)

	// Execute the PHP code
	result := i.Eval(string(phpCode))
	if err, ok := result.(*runtime.Error); ok {
		http.Error(w, fmt.Sprintf("PHP Error: %v", err), http.StatusInternalServerError)
		return
	}

	// Get the output
	output := i.Output()

	// Send the output to the client
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, output)
}

