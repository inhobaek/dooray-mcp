package socketmode

import "errors"

var (
	ErrNoChannel   = errors.New("socketmode: no channel ID available for reply")
	ErrNoWebClient = errors.New("socketmode: web client not initialized")
	ErrNoToken     = errors.New("socketmode: agent token is required")
	ErrNoDomain    = errors.New("socketmode: domain is required (e.g. WithDomain(\"company.dooray.com\"))")
)
