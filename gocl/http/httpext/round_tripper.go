package httpext

import "net/http"

type RoundTripFn func(req *http.Request, t http.RoundTripper) (*http.Response, error)

type roundTripper struct{ client *http.Client }

func (s *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) { return s.client.Do(req) }

func ClientAsRoundTripper(client *http.Client) http.RoundTripper {
	return &roundTripper{client: client}
}
