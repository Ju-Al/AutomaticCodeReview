
// userAgentTransport wraps an http.RoundTripper, adding a User-Agent header
// to each request.
type userAgentTransport struct {
	base http.RoundTripper
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid mutating it.
	newReq := *req
	newReq.Header = make(http.Header)
	for k, vv := range req.Header {
		newReq.Header[k] = vv
	}
	// Append, don't overwrite.
	newReq.Header.Set("User-Agent", req.UserAgent()+" "+GoCloudUserAgent)
	return t.base.RoundTrip(&newReq)
}

// HTTPClient wraps client and appends GoCloudUserAgent to the User-Agent header
// for all requests.
func HTTPClient(client *http.Client) *http.Client {
	c := *client
	c.Transport = &userAgentTransport{base: c.Transport}
	return &c
}
