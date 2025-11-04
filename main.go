package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func replaceHref(v any, pattern, replacement, suffix string) any {
	re := regexp.MustCompile(regexp.QuoteMeta(pattern) + `[^/]+/`)

	switch val := v.(type) {
	case map[string]any:
		for k, v2 := range val {
			if k == "href" {
				if s, ok := v2.(string); ok && strings.Contains(s, pattern) {
					newHref := re.ReplaceAllString(s, replacement+suffix+"/")
					val[k] = newHref
				}
			} else {
				val[k] = replaceHref(v2, pattern, replacement, suffix)
			}
		}
	case []any:
		for i, v2 := range val {
			val[i] = replaceHref(v2, pattern, replacement, suffix)
		}
	}
	return v
}

func handler(w http.ResponseWriter, r *http.Request) {
	upstream := os.Getenv("UPSTREAM_URL")
	pattern := os.Getenv("HREF_PATTERN")
	replacement := os.Getenv("HREF_REPLACEMENT")
	defaultSuffix := os.Getenv("DEFAULT_SUFFIX")

	for k, v := range map[string]string{
		"UPSTREAM_URL":     upstream,
		"HREF_PATTERN":     pattern,
		"HREF_REPLACEMENT": replacement,
		"DEFAULT_SUFFIX":   defaultSuffix,
	} {
		if v == "" {
			errMsg := fmt.Sprintf("Missing required environment variable: %s", k)
			log.Fatal("❌ " + errMsg)
		}
	}

	suffix := r.Header.Get("X-Issuer-Suffix")
	if suffix == "" {
		suffix = defaultSuffix
	}

	target := upstream + r.URL.RequestURI()
	log.Printf("→ Fetching %s (suffix=%s)", target, suffix)

	resp, err := http.Get(target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		for k, v := range resp.Header {
			if len(v) > 0 {
				w.Header().Set(k, v[0])
			}
		}
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
		return
	}

	data = replaceHref(data, pattern, replacement, suffix)
	modified, _ := json.MarshalIndent(data, "", "  ")

	for k, v := range resp.Header {
		if strings.ToLower(k) == "content-length" {
			continue
		}
		if len(v) > 0 {
			w.Header().Set(k, v[0])
		}
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	w.WriteHeader(resp.StatusCode)
	w.Write(modified)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/.well-known/webfinger", handler)
	log.Printf("✅ Listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
