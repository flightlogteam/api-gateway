package presentation

import (
	"github.com/klyngen/jsend"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type ProxyRoute struct {
	Target string
	DestinationAddress string
}

type proxy struct {
	routes []ProxyRoute
}

func (pr ProxyRoute) isDestination(routeUri string) bool {
	url, err := url.Parse(routeUri)

	if err != nil { return false }

	// Is something like /Users/something/something
	path := url.Path
	var splits []string

	if len(path) == 0 {
		splits = strings.Split(routeUri, "/")
	} else {
		splits = strings.Split(path, "/")
	}

	if len(splits) > 1 && splits[1] == pr.Target {
		return true
	}

	return false
}

func (pr ProxyRoute) serve(requestUrl string, res http.ResponseWriter, req *http.Request) {
	originalUrl, _ := url.Parse(requestUrl)

	url, _ := url.Parse(pr.DestinationAddress)
	url.Path = originalUrl.Path
	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)

	// TODO: add client certificate

	// Update the headers to allow for SSL redirection
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}

func (pr proxy) routeMessage(url string, res http.ResponseWriter, req *http.Request)  {
	// TRY TO FIND THE DESTINATION
	for _, r := range pr.routes {
		if r.isDestination(url) {
			r.serve(url, res, req)
			return
		}
	}

	// IF WE GET TO THIS POINT... IT MEANS THIS IS NOT A VALID ROUTE
	jsend.FormatResponse(res, "No such route", jsend.NotFound)

}
