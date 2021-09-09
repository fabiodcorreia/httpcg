package httpcg

import (
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"golang.org/x/net/http2"
)

type HTTPClientBuilder struct {
	connectionTimeout      time.Duration
	keepAlive              time.Duration
	expectContinueTimeout  time.Duration
	idleConnTimeout        time.Duration
	maxIdleConnections     int
	maxHostIdleConnections int
	responseHeaderTimeout  time.Duration
	tlsHandshake           time.Duration
	proxy                  func(*http.Request) (*url.URL, error)
	http2                  bool
	storeCookies           bool
}

// NewBuilder will return a new builder for the http.Client.
func NewBuilder() HTTPClientBuilder {
	return HTTPClientBuilder{
		connectionTimeout:      5 * time.Second,
		expectContinueTimeout:  1 * time.Second,
		idleConnTimeout:        90 * time.Second,
		keepAlive:              30 * time.Second,
		maxIdleConnections:     100,
		maxHostIdleConnections: 10,
		responseHeaderTimeout:  5 * time.Second,
		tlsHandshake:           5 * time.Second,
		proxy:                  http.ProxyFromEnvironment,
		http2:                  false,
		storeCookies:           false,
	}
}

// MaxIdleConn will set how many connections are allowed
// to be on idle in total and per host.
//
// The value of all should be always bigger than host.
func (b HTTPClientBuilder) MaxIdleConn(all, host int) HTTPClientBuilder {
	b.maxHostIdleConnections = host
	b.maxIdleConnections = all
	return b
}

// ConnectionTimeout will set the max amount of time to wait
// until the TCP connections gets established.
func (b HTTPClientBuilder) ConnectionTimeout(t time.Duration) HTTPClientBuilder {
	b.connectionTimeout = t
	return b
}

// TLSHandshakeTimeout will set the max amount of time to wait
// until the TSL Handshake get completed.
func (b HTTPClientBuilder) TLSHandshakeTimeout(t time.Duration) HTTPClientBuilder {
	b.tlsHandshake = t
	return b
}

func (b HTTPClientBuilder) ExpectContinueTimeout(t time.Duration) HTTPClientBuilder {
	b.expectContinueTimeout = t
	return b
}

func (b HTTPClientBuilder) WithKeepAlive(t time.Duration) HTTPClientBuilder {
	b.expectContinueTimeout = t
	return b
}

func (b HTTPClientBuilder) IdleConnTimeout(t time.Duration) HTTPClientBuilder {
	b.idleConnTimeout = t
	return b
}

func (b HTTPClientBuilder) ResponseHeaderTimeout(t time.Duration) HTTPClientBuilder {
	b.idleConnTimeout = t
	return b
}

func (b HTTPClientBuilder) WithHTTP2() HTTPClientBuilder {
	b.http2 = true
	return b
}

func (b HTTPClientBuilder) WithCookies() HTTPClientBuilder {
	b.storeCookies = true
	return b
}

// Build will grab all the builder settings and generate an http.client
func (b HTTPClientBuilder) Build() (*http.Client, error) {
	tr := &http.Transport{
		ResponseHeaderTimeout: b.responseHeaderTimeout,
		Proxy:                 b.proxy,
		DialContext: (&net.Dialer{
			KeepAlive: b.keepAlive,
			Timeout:   b.connectionTimeout,
		}).DialContext,
		MaxIdleConns:          b.maxIdleConnections,
		IdleConnTimeout:       b.idleConnTimeout,
		TLSHandshakeTimeout:   b.tlsHandshake,
		MaxIdleConnsPerHost:   b.maxHostIdleConnections,
		ExpectContinueTimeout: b.expectContinueTimeout,
		ForceAttemptHTTP2:     b.http2,
	}

	if b.http2 {
		err := addHTTP2(tr)
		if err != nil {
			return nil, err
		}
	}

	if b.storeCookies {
		return addCookies(tr)
	}

	return &http.Client{Transport: tr}, nil
}

func addHTTP2(tr *http.Transport) error {
	return http2.ConfigureTransport(tr)
}

func addCookies(tr *http.Transport) (*http.Client, error) {
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: tr,
		Jar:       jar,
	}, nil
}
