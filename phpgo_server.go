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

		// Serve static files (HTML, CSS, JS, images, etc.)
		if r.URL.Path == "/" {
			r.URL.Path = "/index.html"
		}

		// Try to serve the file directly
		filePath := "." + r.URL.Path
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

	// Set up superglobals
	setupSuperglobals(i, r)

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

func setupSuperglobals(i *interpreter.Interpreter, r *http.Request) {
	// Set up $_SERVER
	serverArray := runtime.NewArray()
	serverVars := map[string]string{
		"REQUEST_METHOD":      r.Method,
		"REQUEST_URI":         r.RequestURI,
		"SCRIPT_NAME":         r.URL.Path,
		"DOCUMENT_ROOT":       ".",
		"REMOTE_ADDR":         r.RemoteAddr,
		"HTTP_HOST":           r.Host,
		"HTTP_USER_AGENT":     r.UserAgent(),
		"HTTP_ACCEPT":         r.Header.Get("Accept"),
		"HTTP_ACCEPT_LANGUAGE": r.Header.Get("Accept-Language"),
		"HTTP_ACCEPT_ENCODING": r.Header.Get("Accept-Encoding"),
		"HTTPS":               "off",
	}

	for key, value := range serverVars {
		serverArray.Set(runtime.NewString(key), runtime.NewString(value))
	}
	i.env.SetGlobal("_SERVER", serverArray)

	// Set up $_GET
	if len(r.URL.Query()) > 0 {
		getArray := runtime.NewArray()
		for key, values := range r.URL.Query() {
			if len(values) > 0 {
				getArray.Set(runtime.NewString(key), runtime.NewString(values[0]))
			}
		}
		i.env.SetGlobal("_GET", getArray)
	}

	// Set up $_POST
	if r.Method == "POST" {
		r.ParseForm()
		if len(r.PostForm) > 0 {
			postArray := runtime.NewArray()
			for key, values := range r.PostForm {
				if len(values) > 0 {
					postArray.Set(runtime.NewString(key), runtime.NewString(values[0]))
				}
			}
			i.env.SetGlobal("_POST", postArray)
		}
	}

	// Set up $_COOKIE
	if len(r.Cookies()) > 0 {
		cookieArray := runtime.NewArray()
		for _, cookie := range r.Cookies() {
			cookieArray.Set(runtime.NewString(cookie.Name), runtime.NewString(cookie.Value))
		}
		i.env.SetGlobal("_COOKIE", cookieArray)
	}

	// Set up $_REQUEST (combined GET, POST, COOKIE)
	requestArray := runtime.NewArray()
	if getArray, ok := i.env.GetGlobal("_GET").(*runtime.Array); ok && getArray != nil {
		for _, key := range getArray.Keys() {
			if keyVal, ok := key.(runtime.String); ok {
				requestArray.Set(key, getArray.Get(key))
			}
		}
	}
	if postArray, ok := i.env.GetGlobal("_POST").(*runtime.Array); ok && postArray != nil {
		for _, key := range postArray.Keys() {
			if keyVal, ok := key.(runtime.String); ok {
				requestArray.Set(key, postArray.Get(key))
			}
		}
	}
	if cookieArray, ok := i.env.GetGlobal("_COOKIE").(*runtime.Array); ok && cookieArray != nil {
		for _, key := range cookieArray.Keys() {
			if keyVal, ok := key.(runtime.String); ok {
				requestArray.Set(key, cookieArray.Get(key))
			}
		}
	}
	i.env.SetGlobal("_REQUEST", requestArray)
}