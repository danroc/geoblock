package notify

import "errors"

var ErrNotInitialized = errors.New("service not initialized")

type Service interface {
	Sender

	// Initialize initializes the service from a configuration string.
	Initialize(config string) error
}
