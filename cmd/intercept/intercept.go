package main

import (
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/elazarl/goproxy"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/proxify"
)

// Options to handle intercept
type Options struct {
	ListenAddress string
}

func main() {
	var options Options
	flag.StringVar(&options.ListenAddress, "listen-address", ":8888", "Listen Port")

	gologger.Print().Msgf("Starting Intercepting Proxy")
	proxy, err := proxify.NewProxy(&proxify.Options{
		ListenAddr: options.ListenAddress,
		// Verbose:    true,
		OnRequestCallback: func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			if req.Host == "polling.burpcollaborator.net" && strings.HasSuffix(req.URL.Path, "/burpresults") {
				interceptedBiid := req.URL.Query().Get("biid")
				if interceptedBiid != "" {
					gologger.Print().Msgf("BIID found: %s", interceptedBiid)
					os.Exit(0)
				}
			}
			return req, nil
		},
		OnResponseCallback: func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			return resp
		},
	})
	if err != nil {
		gologger.Fatal().Msgf("%s\n", err)
	}

	gologger.Print().Msgf("%s", proxy.Run())
}
