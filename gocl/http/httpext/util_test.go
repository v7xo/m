//go:build test_unit

package httpext_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/http/httpext"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper/stdcase"
)

func Test_RequestMethodURLMetricLabels(t *testing.T) {
	thelper.Pre(t)

	type testCase struct {
		stdcase.CommonProps

		input *http.Request

		expectedMethod    string
		expectedPath      string
		expectedAuthoriry string
	}

	testCases := map[string]testCase{
		"case0": {
			input: &http.Request{
				Method: http.MethodGet,
				URL:    hlp_ParseURL(t, "https://foo.bar/segment1/segment2/segment3/segment4"),
			},

			expectedMethod:    http.MethodGet,
			expectedPath:      "/segment1/segment2" + httpext.PathSegmentPlaceholder + httpext.PathSegmentPlaceholder,
			expectedAuthoriry: "foo.bar",
		},
		"case1": {
			input: &http.Request{
				Method: http.MethodGet,
				URL:    hlp_ParseURL(t, "https://foo.bar/segment1//segment2/segment3/segment4"),
			},

			expectedMethod:    http.MethodGet,
			expectedPath:      "/segment1//segment2" + httpext.PathSegmentPlaceholder + httpext.PathSegmentPlaceholder,
			expectedAuthoriry: "foo.bar",
		},
		"case2": {
			input: &http.Request{
				Method: http.MethodGet,
				URL:    hlp_ParseURL(t, "https://foo.bar/"),
			},

			expectedMethod:    http.MethodGet,
			expectedPath:      "/",
			expectedAuthoriry: "foo.bar",
		},
		"case3": {
			input: &http.Request{
				Method: http.MethodGet,
				URL:    hlp_ParseURL(t, "https://foo.bar/segment1/45896/segment3/segment4"),
			},

			expectedMethod:    http.MethodGet,
			expectedPath:      "/segment1/" + httpext.NumberPlaceholder + httpext.PathSegmentPlaceholder + httpext.PathSegmentPlaceholder,
			expectedAuthoriry: "foo.bar",
		},
		"case4": {
			input: &http.Request{
				Method: http.MethodGet,
				URL:    hlp_ParseURL(t, "https://foo.bar/00000000-0000-0000-0000-000000000000/45896/segment3/985"),
			},

			expectedMethod:    http.MethodGet,
			expectedPath:      "/" + httpext.UUIDPlaceholder + "/" + httpext.NumberPlaceholder + httpext.PathSegmentPlaceholder + httpext.PathSegmentPlaceholder,
			expectedAuthoriry: "foo.bar",
		},
	}

	for tcName := range testCases {
		tc := testCases[tcName]

		t.Run(
			tcName,
			func(t *testing.T) {
				thelper.Pre(t)

				method, urlAuthority, urlPath := httpext.RequestMethodURLMetricLabels(tc.input)
				require.Equal(t, tc.expectedMethod, method)
				require.Equal(t, tc.expectedAuthoriry, urlAuthority)
				require.Equal(t, tc.expectedPath, urlPath)
			},
		)
	}
}

func hlp_ParseURL(t *testing.T, rawURL string) *url.URL {
	r, err := url.Parse(rawURL)
	require.NoError(t, err)

	return r
}
