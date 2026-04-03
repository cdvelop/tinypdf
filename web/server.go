//go:build !wasm

package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// Solo un parámetro: nocache
	noCacheFlag := flag.Bool("nocache", false, "Disable caching for static files")
	flag.Parse()

	port := "8080"
	publicDir := "public"

	// Resolver ruta absoluta del directorio public
	absPublicDir, err := filepath.Abs(publicDir)
	if err != nil {
		log.Fatalf("Error resolving public directory path: %v", err)
	}

	// Verificar si el directorio existe
	if _, err := os.Stat(absPublicDir); os.IsNotExist(err) {
		log.Fatalf("Static files directory does not exist: %s", absPublicDir)
	}

	log.Printf("Serving static files from: %s", absPublicDir)
	fs := http.FileServer(http.Dir(absPublicDir))

	mux := http.NewServeMux()

	// Handler con o sin middleware de nocache
	var finalHandler http.Handler = fs
	if *noCacheFlag {
		log.Println("Caching disabled")
		finalHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			fs.ServeHTTP(w, r)
		})
	}

	mux.Handle("/", finalHandler)

	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
