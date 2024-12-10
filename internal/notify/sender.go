// Package notify sends notifications through various services.
package notify

type Sender interface {
	// Send sends a notification.
	Send(message string) error
}
