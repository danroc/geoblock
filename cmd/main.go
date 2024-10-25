package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/danroc/geoblock/pkg/rules"
	"github.com/danroc/geoblock/pkg/schema"
)

const (
	HeaderXForwardedMethod = "X-Forwarded-Method"
	HeaderXForwardedProto  = "X-Forwarded-Proto"
	HeaderXForwardedHost   = "X-Forwarded-Host"
	HeaderXForwardedURI    = "X-Forwarded-Uri"
	HeaderXForwardedFor    = "X-Forwarded-For"
)

type Rule struct {
	Policy    string
	Networks  []net.IPNet
	Domains   []string
	Countries []string
}

type Service struct {
	DefaultPolicy string
	Rules         []Rule
}

func getAuthorize(
	w http.ResponseWriter,
	r *http.Request,
) {
	origins := r.Header[HeaderXForwardedFor]

	// Block request: missing header
	if origins == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Find the country code for the client IP (first IP in the list)

	// Block request: IP not found in the database

	// Allow request: country code is in the allowed set
	// w.WriteHeader(http.StatusOK)
	// return

	// Block request: default case
	w.WriteHeader(http.StatusForbidden)
}

func main() {
	cfg, err := schema.ReadFile("examples/configuration.yaml")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%+v\n", cfg)

	engine := rules.NewEngine(cfg.AccessControl)
	query := rules.Query{
		RequestedDomain: "bc.gas.ovh",
		SourceIP:        net.ParseIP("62.35.255.120"),
		SourceCountry:   "US",
		SourceASN:       1235,
	}

	if engine.Authorize(query) {
		fmt.Println("Request authorized")
	} else {
		fmt.Println("Request denied")
	}

	// db, err := database.NewDatabase(countryIPv4URL)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// resolver, err := database.NewResolver()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// match := resolver.Resolve(net.ParseIP("62.35.255.250"))
	// fmt.Println(match)

	// allowedCountryCodes := set.NewSet[string]()
	// allowedCountryCodes.Add("FR")

	// mux := http.NewServeMux()
	// mux.HandleFunc("/v1/authorize", func(w http.ResponseWriter, r *http.Request) {
	// 	getAuthorize(entries, allowedCountryCodes, w, r)
	// })

	// server := http.Server{
	// 	Addr:    ":8080",
	// 	Handler: mux,
	// }

	// log.Printf("Starting server at %s", server.Addr)
	// log.Fatal(server.ListenAndServe())
}
