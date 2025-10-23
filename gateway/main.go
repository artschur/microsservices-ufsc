package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// Service holds the configuration for a backend service
type Service struct {
	URL   *url.URL
	Proxy *httputil.ReverseProxy
}

// NewService creates a new Service instance
func NewService(targetURL string) *Service {
	u, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Failed to parse target URL %s: %v", targetURL, err)
	}
	return &Service{
		URL:   u,
		Proxy: httputil.NewSingleHostReverseProxy(u),
	}
}

func main() {
	// Service registry
	services := map[string]*Service{
		"ingresso": NewService("http://ingressos:8080"),
		"atracoes": NewService("http://atracoes:8081"),
		"usuarios": NewService("http://usuarios:8082"), // Assumed port
		"fila":     NewService("http://fila:8083"),     // Assumed port
		"espera":   NewService("http://espera:8084"),   // Assumed port
	}

	// Main handler to route requests
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var targetService *Service
		path := r.URL.Path

		// Find the service based on the path prefix
		for prefix, service := range services {
			if strings.HasPrefix(path, "/"+prefix) {
				targetService = service
				break
			}
		}

		if targetService != nil {
			log.Printf("Forwarding request for %s to %s", path, targetService.URL)
			// The original request path is preserved by default
			targetService.Proxy.ServeHTTP(w, r)
		} else {
			log.Printf("No service found for path: %s", path)
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	})

	port := "8000"
	fmt.Printf("Gateway is running on port :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
