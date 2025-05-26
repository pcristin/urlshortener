package my_tls

import (
	"golang.org/x/crypto/acme/autocert"
)

// GetTLSManager returns a new TLS manager
func GetTLSManager() *autocert.Manager {
	return &autocert.Manager{
		Cache:      autocert.DirCache("cache-dir"),                       // directory for cache
		Prompt:     autocert.AcceptTOS,                                   // accept terms of service
		HostPolicy: autocert.HostWhitelist("mysite.ru", "www.mysite.ru"), // allow only these hosts
	}
}
