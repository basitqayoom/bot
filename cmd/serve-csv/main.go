package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	var (
		port  = flag.Int("port", 8080, "Port to serve the CSV on")
		path  = flag.String("path", "trade_logs/trades_all_symbols.csv", "Path to the CSV file to serve")
		route = flag.String("route", "/csv", "HTTP route to serve the CSV at")
		cors  = flag.Bool("cors", true, "Enable permissive CORS headers (Access-Control-Allow-Origin: *)")
	)
	flag.Parse()

	absPath, err := filepath.Abs(*path)
	if err != nil {
		log.Fatalf("failed to resolve CSV path: %v", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		log.Fatalf("csv not accessible at %s: %v", absPath, err)
	}
	if info.IsDir() {
		log.Fatalf("provided path is a directory, expected file: %s", absPath)
	}

	http.HandleFunc(*route, func(w http.ResponseWriter, r *http.Request) {
		if *cors {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", filepath.Base(absPath)))
		http.ServeFile(w, r, absPath)
	})

	// Simple index to help discover the route.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintf(w, "CSV server is running.\n\nFile: %s\nRoute: %s\n\nDownload: http://%s%s\n", absPath, *route, r.Host, *route)
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Serving %s on http://localhost%s%s (CORS=%v)", absPath, addr, *route, *cors)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
