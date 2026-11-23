package httpext

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"regexp"
	"strings"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

const (
	NumberPlaceholder      = "<num>"
	UUIDPlaceholder        = "<uuid>"
	PathSegmentPlaceholder = "/<segment>"
)

var (
	uuidRegexpPattern = `([a-fA-F0-9]{8}\-[a-fA-F0-9]{4}\-[a-fA-F0-9]{4}\-[a-fA-F0-9]{4}\-[a-fA-F0-9]{12})`
	uuidRegexp        = regexp.MustCompile(uuidRegexpPattern)

	numberRegexpPattern = `([\p{Nd}]{2,})`
	numberRegexp        = regexp.MustCompile(numberRegexpPattern)

	pathSegmentPattern = `(/[^\/]+)`
	pathSegmentRegexp  = regexp.MustCompile(pathSegmentPattern)
)

// TryUnmarshalJSONContent checks whether specified [http.Header] declares content type as JSON and if it such, then unmarshals JSON to specified value v.
func TryUnmarshalJSONContent(h http.Header, payload []byte, v any) (ok bool, err error) {
	if ok, err = HeaderDeclaresJSONContentType(h); err != nil || !ok {
		return
	}

	if err = json.Unmarshal(payload, v); err != nil {
		err = fmt.Errorf("failed to unmarshal payload as JSON; err: %w", err)
	}

	return
}

// UnmarshalJSONContent checks whether specified [http.Header] declares content type as JSON and if it such, then unmarshals JSON to specified value v.
func UnmarshalJSONContent(h http.Header, payload []byte, v any) error {
	return UnmarshalJSONContentReader(h, bytes.NewBuffer(payload), v)
}

// UnmarshalJSONContentReader checks whether specified [http.Header] declares content type as JSON and if it such, then unmarshals JSON to specified value v.
func UnmarshalJSONContentReader(h http.Header, payload io.Reader, v any) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to unmarshal payload as JSON; err: %w", err)
		}
	}()

	var isJSON bool

	if isJSON, err = HeaderDeclaresJSONContentType(h); err != nil {
		return
	} else if !isJSON {
		err = errors.New("content type header is missing or not declares content as JSON")

		return
	}

	err = json.NewDecoder(payload).Decode(v)

	return
}

func HeaderDeclaresJSONContentType(h http.Header) (bool, error) {
	return headerDeclaresJSONContentType(h, true)
}

func headerDeclaresJSONContentType(h http.Header, recursion bool) (bool, error) {
	vals := h.Values(ContentTypeHeaderName)

	if len(vals) == 0 {
		return false, nil
	} else if len(vals) != 1 {
		switch {
		case recursion:
			for i := range vals {
				res, err := headerDeclaresJSONContentType(http.Header{ContentTypeHeaderName: []string{vals[i]}}, false)
				if err != nil {
					return false, err
				} else if res {
					return true, nil
				}
			}

			return false, nil
		}

		return false, errs.BadArgsWrapWith(errors.New("content type header is ambiguous"))
	}

	name, params, err := mime.ParseMediaType(vals[0])
	if err != nil {
		return false, errs.BadArgsWrapWith(fmt.Errorf("failed to parse content type header as MIME type; err: %w", err))
	}

	if name != JSONContentType {
		return false, nil
	}

	charset, ok := params["charset"]
	if !ok {
		return true, nil
	}

	const utf8CharsetParam = "utf-8"

	if !strings.EqualFold(charset, utf8CharsetParam) {
		return false, errs.BadArgsWrapWith(fmt.Errorf("content type charset parameter mismatches expected; expected: %s; actual: %s", utf8CharsetParam, charset))
	}

	return true, nil
}

func RequestMethodURLMetricLabels(req *http.Request) (method string, urlAuthority string, urlPath string) {
	method, urlAuthority, urlPath = placeholder.Nil, placeholder.Nil, placeholder.Nil

	if req == nil {
		return
	}

	method = req.Method

	if req.URL == nil {
		return
	}

	urlAuthority = req.URL.Host
	urlPath = req.URL.Path

	switch urlPath {
	case "/", "":
		return
	}

	const maxNonSubstitutedPathSegments = 2

	nonSubstitutedPathSegments := 0

	urlPath = pathSegmentRegexp.ReplaceAllStringFunc(
		urlPath,
		func(s string) string {
			if nonSubstitutedPathSegments == maxNonSubstitutedPathSegments {
				return PathSegmentPlaceholder
			}

			nonSubstitutedPathSegments++

			return s
		},
	)

	urlPath = uuidRegexp.ReplaceAllString(urlPath, UUIDPlaceholder)
	urlPath = numberRegexp.ReplaceAllString(urlPath, NumberPlaceholder)

	return
}
