package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/cod3rboy/prevue-rproxy/env"
)

// Supports <word>-<port> pairs in subdomain part. Delimiter - is used only once in subdomain string
// to separate container name and application port number
var subdomainRegex = regexp.MustCompile(`^([^-]+)(-\d{4})?$`)

func ListenAndReverse() {
	http.ListenAndServe(fmt.Sprintf(":%s", env.Port), http.HandlerFunc(proxyRequestHandler))
}

func proxyRequestHandler(rw http.ResponseWriter, req *http.Request) {
	containerName, port, err := splitContainerNameAndPort(req.Host)
	if err != nil {
		SendErrorResponse(rw, err)
		return
	}
	containerURL, err := makeContainerAccessURL(containerName, port)
	if err != nil {
		SendErrorResponse(rw, err)
		return
	}

	// Modifying request to forward it to container application
	req.Host = containerURL.Host
	req.URL.Host = containerURL.Host
	req.URL.Scheme = containerURL.Scheme
	req.RequestURI = ""

	// Send request and get response from container application
	containerResponse, err := http.DefaultClient.Do(req)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(rw, err)
		return
	}
	// Return response to the client
	rw.WriteHeader(http.StatusOK)
	io.Copy(rw, containerResponse.Body)
}

func splitContainerNameAndPort(host string) (string, string, error) {
	subdomain := ""
	if hostParts := strings.Split(host, "."); len(hostParts) > 1 {
		subdomain = hostParts[0]
	} else {
		return "", "", ErrorNotFound("proxy: url without subdomain is not allowed")
	}
	matches := subdomainRegex.FindStringSubmatch(subdomain)
	if matches == nil {
		return "", "", ErrorNotFound("proxy: invalid subdomain format")
	}
	matches = matches[1:] // Discard whole match value

	return matches[0], strings.TrimLeft(matches[1], "-"), nil
}

func makeContainerAccessURL(containerName, applicationPort string) (*url.URL, error) {
	if applicationPort != "" {
		applicationPort = fmt.Sprintf(":%s", applicationPort)
	}
	rawContainerURL := fmt.Sprintf("http://%s%s", containerName, applicationPort)
	containerUrl, err := url.Parse(rawContainerURL)
	if err != nil {
		return nil, ErrorContainerUrlMalformed(rawContainerURL)
	}
	return containerUrl, nil
}
