package pubsub

// PublisherClosedError indicates that the publisher has been closed.
type PublisherClosedError struct{}

func (e PublisherClosedError) Error() string {
	return "publisher is closed"
}

// SubscriberClosedError indicates that the subscriber has been closed.
type SubscriberClosedError struct{}

func (e SubscriberClosedError) Error() string {
	return "subscriber is closed"
}

// Pub-sub errors for in-memory implementation.
var (
	// ErrPublisherClosed indicates that the publisher has been closed.
	ErrPublisherClosed = PublisherClosedError{}

	// ErrSubscriberClosed indicates that the subscriber has been closed.
	ErrSubscriberClosed = SubscriberClosedError{}
)
