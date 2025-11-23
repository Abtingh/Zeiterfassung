package middleware

import (
	"net/http"
	"time"
)

// NoCache is a middleware function that wraps an http.HandlerFunc.
// It takes the 'next' handler (the actual logic you want to run) as input
// and returns a new handler that adds headers before running 'next'.
func NoCache(next http.HandlerFunc) http.HandlerFunc {
    // We return an anonymous function that matches the http.HandlerFunc signature.
    // This function acts as the new request handler.
    return func(w http.ResponseWriter, r *http.Request) {
        // Sets the standard HTTP 1.1 header to disable caching.
        // 'no-store': Don't save the response to disk/memory.
        // 'no-cache': Must re-validate with the server before using a cached copy.
        // 'must-revalidate': Once expired, must verify with server.
        // 'proxy-revalidate': Same as above, but specifically for shared caches (proxies).
        w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")

        // Sets the HTTP 1.0 header for backward compatibility with older clients/proxies
        // that might not understand Cache-Control.
        w.Header().Set("Pragma", "no-cache")

		// Sets the expiration date to a time in the past.
		// This forces the browser to consider the content immediately stale.
		w.Header().Set("Expires", time.Unix(0, 0).Format(http.TimeFormat))        // Sets a header specifically for CDNs (Content Delivery Networks) and reverse proxies
        // (like Fastly or Varnish) to tell them not to store this response.
        w.Header().Set("Surrogate-Control", "no-store")

        // Finally, execute the original handler logic (the 'next' function)
        // passing the ResponseWriter and Request to it.
        next.ServeHTTP(w, r)
    }
}