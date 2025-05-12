package tls

import (
	"golang.org/x/crypto/acme/autocert"
)

func GetTLSManager() *autocert.Manager {
	return &autocert.Manager{
		Cache:      autocert.DirCache("cache-dir"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("mysite.ru", "www.mysite.ru"),
	}
}
