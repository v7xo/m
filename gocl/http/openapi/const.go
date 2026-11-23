package openapi

import "rn.m.vo7.io/vo7io/m/gocl/http/httpext"

const (
	jsonContentType            = httpext.JSONContentType
	jsonContentTypeWithCharset = httpext.JSONContentTypeWithCharset

	contentTypeHeaderName   = httpext.ContentTypeHeaderName
	contentLengthHeaderName = "Content-Length"

	failedToWriteResponseMsgTemplate = "failed to write response; request URI: %s"

	DefaultContentLengthLimit = 4194304 // 4 MB
)
