//go:build wasm
// +build wasm

package env

import (
	"fmt"
	"syscall/js"
)

// SetupDefaultLogger configures the default logger for frontend environments
func SetupDefaultLogger() func(a ...any) {
	return func(a ...any) {
		// Use console.log in browser environment
		args := make([]any, len(a))
		for i, arg := range a {
			args[i] = js.ValueOf(anyToString(arg))
		}
		js.Global().Get("console").Call("log", args...)
	}
}

// SetupDefaultFileWriter configures the default file writer for frontend environments
func SetupDefaultFileWriter() func(filename string, data []byte) error {
	return func(filename string, data []byte) error {
		return fmt.Errorf("file writing not implemented in frontend")
	}
}

// isURL checks if a string is a valid URL
func isURL(str string) bool {
	// Simple check for URL format (starts with http:// or https://)
	return len(str) > 8 && (str[:7] == "http://" || str[:8] == "https://")
}

// FetchURL retrieves content from a URL using JavaScript's fetch API
// Returns the content as []byte and an error if the request fails
func FetchURL(url string) ([]byte, error) {
	// Create a promise to fetch the URL
	fetchPromise := js.Global().Call("fetch", url)

	// Create channels for response and error handling
	respChan := make(chan []byte)
	errChan := make(chan error)

	// Handle the promise resolution
	fetchPromise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Get the response object
		response := args[0]

		// Check if response is ok
		if !response.Get("ok").Bool() {
			statusText := response.Get("statusText").String()
			status := response.Get("status").Int()
			errChan <- fmt.Errorf("HTTP error: %d %s", status, statusText)
			return nil
		}

		// Get the blob from the response
		response.Call("arrayBuffer").Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			// Convert ArrayBuffer to Uint8Array
			arrayBuffer := args[0]
			uint8Array := js.Global().Get("Uint8Array").New(arrayBuffer)

			// Convert to Go []byte
			length := uint8Array.Get("length").Int()
			data := make([]byte, length)
			js.CopyBytesToGo(data, uint8Array)

			respChan <- data
			return nil
		}))
		return nil
	})).Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Handle fetch errors
		err := args[0]
		errChan <- js.Error{Value: err}
		return nil
	}))

	// Wait for either response or error
	select {
	case data := <-respChan:
		return data, nil
	case err := <-errChan:
		return nil, err
	}
}

// FileExists checks if a file exists and returns its contents.
// Accepts string (path/URL) or []byte (content).
// In frontend environment:
// - If given a URL, it will fetch the content using JavaScript fetch API
// - If given local file path, it warns that direct access is not supported
// - If given []byte, it returns the content directly
func FileExists(pathOrContent any) ([]byte, error) {
	console := js.Global().Get("console")

	switch v := pathOrContent.(type) {
	case string:
		// Check if it's a URL
		if isURL(v) {
			// Use the reusable FetchURL function
			return FetchURL(v)
		} else {
			// If it's a local file path, we can't access it directly
			console.Call("log", "FileExists: Local file system access not supported in browser", v)
			return nil, fmt.Errorf("local file system access not supported in browser")
		}

	case []byte:
		// If content is provided directly, return it as is
		return v, nil

	default:
		errMsg := fmt.Sprintf("unsupported type: %T, expected string or []byte", pathOrContent)
		console.Call("log", "FileExists error:", errMsg)
		return nil, fmt.Errorf(errMsg)
	}
}

// GetSize returns the size of content from a URL or a byte slice.
// For frontend: fetches URL content for size, uses len() for byte slices.
// Does not support local file paths.
func GetSize(pathOrContent any) (int64, error) {
	console := js.Global().Get("console")

	switch v := pathOrContent.(type) {
	case string:
		// Check if it's a URL
		if isURL(v) {
			// Fetch the content to get its size
			content, err := FetchURL(v)
			if err != nil {
				return -1, err
			}
			return int64(len(content)), nil
		} else {
			// Local file paths are not supported
			errMsg := "GetSize: Local file system access not supported in browser"
			console.Call("log", errMsg, v)
			return -1, fmt.Errorf(errMsg)
		}
	case []byte:
		// It's already content
		return int64(len(v)), nil
	default:
		errMsg := fmt.Sprintf("unsupported type for GetSize: %T", pathOrContent)
		console.Call("log", "GetSize error:", errMsg)
		return -1, fmt.Errorf(errMsg)
	}
}

// FileOpen opens a file and returns a ReadSeekCloser implementation
// For frontend: creates a ByteReader from fetched URL content or byte array
func FileOpen(pathOrContent any) (ReadSeekCloser, error) {
	switch v := pathOrContent.(type) {
	case string:
		// If it's a URL, fetch the content
		if isURL(v) {
			content, err := FetchURL(v)
			if err != nil {
				return nil, err
			}
			return NewByteReader(content), nil
		} else {
			return nil, fmt.Errorf("local file system access not supported in browser")
		}

	case []byte:
		// Create a ReadSeekCloser from bytes
		return NewByteReader(v), nil

	default:
		return nil, fmt.Errorf("unsupported type for FileOpen: %T", pathOrContent)
	}
}

// anyToString converts any type to string for console logging
func anyToString(v any) string {
	if v == nil {
		return "nil"
	}
	return fmt.Sprint(v)
}
