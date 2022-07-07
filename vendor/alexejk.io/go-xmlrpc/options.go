package xmlrpc

import "net/http"

// Option is a function that configures a Client by mutating it
type Option func(client *Client)

// Headers option allows setting custom headers that will be passed with every request
func Headers(headers map[string]string) Option {
	return func(client *Client) {
		client.codec.customHeaders = headers
	}
}

// HttpClient option allows setting custom HTTP Client to be used for every request
func HttpClient(httpClient *http.Client) Option {
	return func(client *Client) {
		client.codec.httpClient = httpClient
	}
}

// UserAgent option allows setting custom User-Agent header.
// This is a convenience method when only UA needs to be modified. For other cases use Headers option.
func UserAgent(userAgent string) Option {
	return func(client *Client) {
		client.codec.userAgent = userAgent
	}
}
