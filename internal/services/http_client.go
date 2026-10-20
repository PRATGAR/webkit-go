package services

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

var fastHTTPClient = &fasthttp.Client{}

func FetchURL(targetURL string) ([]byte, error) {
	resp, err := http.Get(targetURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func FetchWithClient(targetURL string) ([]byte, error) {
	resp, err := httpClient.Get(targetURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func BuildAndFetch(baseURL, path string) ([]byte, error) {
	fullURL := baseURL + path
	return FetchURL(fullURL)
}

func FetchWithHeaders(targetURL string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

type URLBuilder struct {
	scheme string
	host   string
	path   string
	query  url.Values
}

func NewURLBuilder() *URLBuilder {
	return &URLBuilder{
		query: make(url.Values),
	}
}

func (b *URLBuilder) SetScheme(scheme string) *URLBuilder {
	b.scheme = scheme
	return b
}

func (b *URLBuilder) SetHost(host string) *URLBuilder {
	b.host = host
	return b
}

func (b *URLBuilder) SetPath(path string) *URLBuilder {
	b.path = path
	return b
}

func (b *URLBuilder) AddQuery(key, value string) *URLBuilder {
	b.query.Add(key, value)
	return b
}

func (b *URLBuilder) Build() string {
	u := url.URL{
		Scheme:   b.scheme,
		Host:     b.host,
		Path:     b.path,
		RawQuery: b.query.Encode(),
	}
	return u.String()
}

func (b *URLBuilder) Fetch() ([]byte, error) {
	return FetchURL(b.Build())
}

type RequestForwarder interface {
	Forward(url string) ([]byte, error)
}

type PassthroughForwarder struct{}

func (f *PassthroughForwarder) Forward(url string) ([]byte, error) {
	return FetchURL(url)
}

type CallbackHandler struct {
	callback func(string) ([]byte, error)
}

func (h *CallbackHandler) SetCallback(cb func(string) ([]byte, error)) {
	h.callback = cb
}

func (h *CallbackHandler) Execute(url string) ([]byte, error) {
	return h.callback(url)
}

func FastHTTPFetch(targetURL string) ([]byte, error) {
	statusCode, body, err := fastHTTPClient.Get(nil, targetURL)
	if err != nil {
		return nil, err
	}
	if statusCode != fasthttp.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", statusCode)
	}
	return body, nil
}

func ForwardRequest(targetURL string, headers map[string]string) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(targetURL)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if err := fastHTTPClient.Do(req, resp); err != nil {
		return nil, err
	}

	return resp.Body(), nil
}

func FetchURLValidated(targetURL string) ([]byte, error) {
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	if parsed.Scheme != "https" {
		return nil, fmt.Errorf("only HTTPS allowed")
	}

	host := parsed.Hostname()
	if isInternalHost(host) {
		return nil, fmt.Errorf("internal hosts not allowed")
	}

	return FetchWithClient(targetURL)
}

func isInternalHost(host string) bool {
	if host == "localhost" || host == "127.0.0.1" {
		return true
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}

	for _, cidr := range privateRanges {
		_, network, _ := net.ParseCIDR(cidr)
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

func FetchURLAllowlisted(targetURL string, allowedHosts []string) ([]byte, error) {
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	allowed := false
	for _, host := range allowedHosts {
		if parsed.Host == host {
			allowed = true
			break
		}
	}

	if !allowed {
		return nil, fmt.Errorf("host not in allowlist")
	}

	return FetchWithClient(targetURL)
}

func MakeWebhookCall(webhookURL string, payload []byte) error {
	resp, err := http.Post(webhookURL, "application/json", strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func ProxyRequest(ctx context.Context, targetURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
