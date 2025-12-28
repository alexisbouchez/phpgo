package main

import (
	"fmt"
	"github.com/alexisbouchez/phpgo/interpreter"
)

func main() {
	// Create a new interpreter
	i := interpreter.New()

	// Set HTTP context with test data
	i.SetHTTPContext("GET", "/test.php", "name=John&age=30", map[string]string{
		"User-Agent": "SuperglobalTest/1.0",
		"Accept":     "text/html",
	}, map[string]string{
		"session_id": "abc123",
	}, map[string]string{
		"email": "john@example.com",
	}, map[string][]byte{
		"profile_pic": []byte("fake image data"),
	})

	// Debug: Check if HTTP context was set
	httpCtx := i.GetHTTPContext()
	fmt.Println("HTTP Context URI:", httpCtx.URI)
	fmt.Println("HTTP Context QueryString:", httpCtx.QueryString)
	fmt.Println("HTTP Context ServerVars:", httpCtx.ServerVars)
	fmt.Println("Current Dir:", i.GetCurrentDir())
	
	// Execute a simple test to check superglobals
	result := i.Eval(`<?php
var_dump($_SERVER);
`)

	// Output the result
	fmt.Print(i.Output())
	if result != nil && result.Type() == "error" {
		fmt.Println("Error:", result.Inspect())
	}
}