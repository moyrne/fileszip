package fileszip

import (
	"net"
	"net/http"
	"strings"
	"time"
)

type Option func(f *FilesZip)

func NewFilesZip(options ...Option) *FilesZip {
	filesZip := &FilesZip{
		debug: true,
		client: &http.Client{
			Transport: &http.Transport{
				DialContext:           (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 10 * time.Second}).DialContext,
				IdleConnTimeout:       10 * time.Second,
				ResponseHeaderTimeout: 5 * time.Second,
				ExpectContinueTimeout: 5 * time.Second,
			},
			Timeout: 10 * time.Second,
		},
		userHook: &DefaultUserHook{
			replacer: strings.NewReplacer("https://", "", "http://", "", "/", "-", ":", "-"),
		},
	}

	for _, option := range options {
		option(filesZip)
	}

	return filesZip
}

func SetDebug() Option {
	return func(f *FilesZip) {
		f.debug = true
	}
}

func SetClient(client Client) Option {
	return func(f *FilesZip) {
		f.client = client
	}
}

func SetUserHook(hook UserHook) Option {
	return func(f *FilesZip) {
		f.userHook = hook
	}
}
