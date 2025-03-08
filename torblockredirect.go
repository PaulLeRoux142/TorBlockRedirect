// Package TorBlockRedirect implements a Traefik plugin for blocking requests from the Tor network
package TorBlockRedirect

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"text/template"
	"time"
)

var ipRegex = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b|\b[0-9a-fA-F:]{2,39}\b`)

// Config holds the configuration for the plugin.
type Config struct {
	Enabled						bool
	AddressListURL				string
	UpdateIntervalSeconds		int
	RedirectProtocol			string
	RedirectHostname			string
	RedirectSavePath			bool
	ForwardedHeadersCustomName	string
}

// CreateConfig initializes the default configuration for the plugin.
func CreateConfig() *Config {
	return &Config{
		Enabled:					true,
		AddressListURL:				"https://check.torproject.org/exit-addresses",
		UpdateIntervalSeconds:		3600,
		RedirectProtocol:			"http://",
		RedirectHostname:			"",
		RedirectSavePath:			true,
		ForwardedHeadersCustomName:	"X-Forwarded-For",
	}
}

// TorBlock represents the main structure of the plugin.
type TorBlock struct {
	next						http.Handler
	name						string
	template					*template.Template
	enabled						bool
	addressListURL				string
	updateInterval				time.Duration
	blockedIPs					*IPv4Set
	blockedIPv6s				*IPv6Set
	client						*http.Client
	redirectProtocol			string
	redirectHostname			string
	redirectSavePath			bool
	ForwardedHeadersCustomName	string
}

// New creates and initializes a new instance of the TorBlock plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	// Validate the address list URL
	_, err := url.ParseRequestURI(config.AddressListURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse exit-addresses URL")
	}
	// Ensure the update interval is not too short
	if config.UpdateIntervalSeconds < 60 {
		return nil, fmt.Errorf("update interval cannot be less than 60 seconds")
	}

	a := &TorBlock{
		next:						next,
		name:						name,
		template:					template.New("TorBlockRedirect").Delims("[[", "]]"),
		enabled:					config.Enabled,
		addressListURL:				config.AddressListURL,
		updateInterval:				time.Duration(config.UpdateIntervalSeconds) * time.Second,
		blockedIPs:					CreateIPv4Set(),
		blockedIPv6s:				CreateIPv6Set(),
		client: &http.Client{
			Timeout: time.Second * 10,
		},
		redirectProtocol:			config.RedirectProtocol,
		redirectHostname:			config.RedirectHostname,
		redirectSavePath:			config.RedirectSavePath,
		ForwardedHeadersCustomName: config.ForwardedHeadersCustomName,
	}
	a.UpdateBlockedIPs()
	go a.UpdateWorker()

	return a, nil
}

// ServeHTTP processes each incoming request that passes through the plugin.
func (a *TorBlock) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Check if the plugin is enabled
	if !a.enabled {
		a.next.ServeHTTP(rw, req)
		return
	}

	// Extract the client IP from the ForwardedHeadersCustomName header
	remoteHost := req.Header.Get(a.ForwardedHeadersCustomName)
	if remoteHost == "" {
		remoteHost = req.RemoteAddr
	}

	// Try to parse the IP as IPv6, if that fails try IPv4
	remoteIP := net.ParseIP(remoteHost)
	if remoteIP == nil {
		log.Printf("TorBlockRedirect: failed to parse IP from remote address: %v", remoteHost)
		a.next.ServeHTTP(rw, req)
		return
	}

	isTorBlocked := true

	// Check if the IP is in the blocked IPv4 set
	if remoteIP.To4() != nil {
		ipv4 := CreateIPv4(remoteIP[0], remoteIP[1], remoteIP[2], remoteIP[3])
		if a.blockedIPs.ContainsIPv4(ipv4) {
			isTorBlocked = true
		}
	} else { // If it's IPv6, check in the blocked IPv6 set
		var ipv6 [16]uint8
		copy(ipv6[:], remoteIP.To16()) // Convert net.IP to 16-byte array
		if a.blockedIPv6s.ContainsIPv6(CreateIPv6(ipv6)) {
			isTorBlocked = true
		}
	}

	if isTorBlocked {
		// If redirectHostname is set, redirect the user to the .onion address
		if a.redirectHostname != "" {
			redirectURL := a.redirectProtocol + a.redirectHostname
			if a.redirectSavePath {
				redirectURL = redirectURL + req.URL.RequestURI()
			}
			log.Printf("TorBlockRedirect: redirecting to %s", redirectURL)
			http.Redirect(rw, req, redirectURL, http.StatusFound)
			return
		}

		// If no redirect is configured, send a Forbidden response
		log.Printf("TorBlockRedirect: request denied (%s)", remoteHost)
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	// If the IP is not blocked, continue processing the request
	a.next.ServeHTTP(rw, req)
}

// UpdateWorker periodically updates the list of blocked IPs according to the update interval.
func (a *TorBlock) UpdateWorker() {
	// Continuously update blocked IPs at the defined interval
	for range time.Tick(a.updateInterval) {
		a.UpdateBlockedIPs()
	}
}

// UpdateBlockedIPs fetches the list of blocked IPs from the addressListURL and updates the blocked IP sets.
func (a *TorBlock) UpdateBlockedIPs() {
	// Fetch the list of exit node IPs from the Tor Project's address list
	log.Printf("TorBlockRedirect: starting update address list: from %s", a.addressListURL)
	res, err := a.client.Get(a.addressListURL)
	if err != nil {
		log.Printf("TorBlockRedirect: failed to update address list: %s", err)
		return
	}
	if res.StatusCode != 200 {
		log.Printf("TorBlockRedirect: failed to update address list: status code is %d", res.StatusCode)
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("TorBlockRedirect: failed to read address list body: %s", err)
		return
	}
	bodyStr := string(body)

	// Extract all IP addresses from the fetched list
	foundIPStrs := ipRegex.FindAllString(bodyStr, -1)
	newIPv4Set := CreateIPv4Set()
	newIPv6Set := CreateIPv6Set()

	// Parse and add each found IP to the respective set (IPv4 or IPv6)
	for _, ipStr := range foundIPStrs {
		// Try to parse the IP as IPv4
		if ip := net.ParseIP(ipStr); ip != nil && ip.To4() != nil {
			ipv4, err := ParseIPv4(ipStr)
			if err == nil {
				newIPv4Set.AddIPv4(ipv4)
			}
		} else { // Or parse it as IPv6
			ipv6, err := ParseIPv6(ipStr)
			if err == nil {
				newIPv6Set.AddIPv6(ipv6)
			}
		}
	}

	// Update the blocked IP sets with the newly fetched data
	a.blockedIPs.AddIPv4Set(newIPv4Set) // Добавляем все новые IPv4-адреса в текущий набор
	a.blockedIPv6s.AddIPv6Set(newIPv6Set) // Добавляем все новые IPv6-адреса в текущий набор
	//	a.blockedIPs = newIPv4Set
	//	a.blockedIPv6s = newIPv6Set

	log.Printf("TorBlockRedirect: updated blocked IP list (found %d IPv4 addresses, %d IPv6 addresses)", len(newIPv4Set.set), len(newIPv6Set.set))
}