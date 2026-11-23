package openapi

type (
	EchoParams struct {
		Text string `json:"text,omitempty"`
	}

	//nolint:tagliatelle
	EchoReply struct {
		BuildInfo string `json:"buildInfo,omitempty"`
		ClientID  string `json:"clientID,omitempty"`
		Text      string `json:"text,omitempty"`
	}
)
