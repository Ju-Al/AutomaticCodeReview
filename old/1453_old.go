package push

import (
	"net/http"
	"strings"

	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func (h Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {

	pusher, hasPusher := w.(http.Pusher)

	// No Pusher, no cry
	if !hasPusher {
		return h.Next.ServeHTTP(w, r)
	}

	// This is request for the pushed resource - it should not be recursive
	if _, exists := r.Header[pushHeader]; exists {
		return h.Next.ServeHTTP(w, r)
	}

	headers := h.filterProxiedHeaders(r.Header)

	// Push first
outer:
	for _, rule := range h.Rules {
		if httpserver.Path(r.URL.Path).Matches(rule.Path) {
			for _, resource := range rule.Resources {
				pushErr := pusher.Push(resource.Path, &http.PushOptions{
					Method: resource.Method,
					Header: h.mergeHeaders(headers, resource.Header),
				})
				if pushErr != nil {
					// If we cannot push (either not supported or concurrent streams are full - break)
					break outer
				}
			}
		}
	}

	// Serve later
	code, err := h.Next.ServeHTTP(w, r)

	if links, exists := w.Header()["Link"]; exists {
		h.servePreloadLinks(pusher, headers, links)
	}

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	return code, err
}

func (h Middleware) servePreloadLinks(pusher http.Pusher, headers http.Header, links []string) {
outer:
	for _, link := range links {
		parts := strings.Split(link, ";")

		if link == "" || strings.HasSuffix(link, "nopush") {
			continue
		}

		target := strings.TrimSuffix(strings.TrimPrefix(parts[0], "<"), ">")

		err := pusher.Push(target, &http.PushOptions{Method: http.MethodGet})
		if err != nil {
			break outer
		}
	}
}
			filter[header] = val
		}
	}

	return filter
}

var proxiedHeaders = []string{
	"Accept-Encoding",
	"Accept-Language",
	"Cache-Control",
	"Host",
	"User-Agent",
}
