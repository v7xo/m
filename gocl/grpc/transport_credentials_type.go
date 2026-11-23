package grpc

import "strings"

type TransportCredentialsPreset string

const (
	TransportCredentialsPresetUnspecified     = TransportCredentialsPreset("")
	TransportCredentialsPresetPlaintext       = TransportCredentialsPreset("plaintext")
	TransportCredentialsPresetTLSInsecure     = TransportCredentialsPreset("tls_insecure")
	TransportCredentialsPresetTLSSecureClient = TransportCredentialsPreset("tls_secure_client")
	TransportCredentialsPresetTLSSecureServer = TransportCredentialsPreset("tls_secure_server")
)

func (v TransportCredentialsPreset) IsUnspecified() bool {
	return v == TransportCredentialsPresetUnspecified
}

func (v TransportCredentialsPreset) IsPlaintext() bool {
	return strings.EqualFold(string(v), string(TransportCredentialsPresetPlaintext))
}

func (v TransportCredentialsPreset) IsTLSInsecure() bool {
	return strings.EqualFold(string(v), string(TransportCredentialsPresetTLSInsecure))
}

func (v TransportCredentialsPreset) IsTLSSecureClient() bool {
	return strings.EqualFold(string(v), string(TransportCredentialsPresetTLSSecureClient))
}

func (v TransportCredentialsPreset) IsTLSSecureServer() bool {
	return strings.EqualFold(string(v), string(TransportCredentialsPresetTLSSecureServer))
}

func (v TransportCredentialsPreset) String() string {
	return string(v)
}
