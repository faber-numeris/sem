package domain

import "errors"

var (
	ErrNotFound             = errors.New("requested resource does not exist")
	ErrInvalidVersion       = errors.New("invalid API version")
	ErrInvalidGrant         = errors.New("invalid authentication grant")
	ErrInvalidClient        = errors.New("invalid client credentials")
	ErrInvalidCredentials   = errors.New("invalid user credentials")
	ErrInvalidToken         = errors.New("invalid or expired token")
	ErrInvalidRequest       = errors.New("invalid request parameters")
	ErrRequiredFieldMissing = errors.New("required field missing")
	ErrUnsupportedSObject   = errors.New("unsupported sObject type")
	ErrInvalidSObject       = ErrUnsupportedSObject
	ErrMalformedQuery       = errors.New("malformed SOQL or SOSL query")
	ErrMalformedSearch      = errors.New("malformed SOSL search query")
	ErrInvalidJobState      = errors.New("invalid bulk job state transition")
)
