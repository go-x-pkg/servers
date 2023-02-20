package servers

import "errors"

var (
	ErrUnmarshalUnknownKind = errors.New("unknown server-kind, no server associated with kind")

	ErrTLSCertFileNotExists       = errors.New("tls cert-file doesn't exists")
	ErrTLSCertFilePathNotProvided = errors.New("tls cert-file path is not provided")

	ErrTLSKeyFileNotExists       = errors.New("tls key-file doesn't exists")
	ErrTLSKeyFilePathNotProvided = errors.New("tls key-file path is not provided")

	ErrUnixSocketParentDirNotExists = errors.New("unix socket parent dir doesn't exists")
	ErrUnixSocketPathNotProvided    = errors.New("tls key-file path is not provided")

	ErrGotBothInetAndUnix = errors.New("provided server is both unix and inet")

	ErrLoadTrustedCA = errors.New("error load trusted CA")

	ErrUnknownVersionTLS = errors.New("unknown version TLS")

	ErrUnknownClientAuthTypeTLS = errors.New("unknown client auth type TLS")
)
