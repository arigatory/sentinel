package retry

import (
	"errors"
	"net"
	"syscall"
)

type NetworkErrorClassifier struct{}

func NewNetworkErrorClassifier() *NetworkErrorClassifier {
	return &NetworkErrorClassifier{}
}

func (c *NetworkErrorClassifier) Classify(err error) ErrorClassification {
	if err == nil {
		return NonRetriable
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return Retriable
		}
	}

	if errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ETIMEDOUT) ||
		errors.Is(err, syscall.ENETUNREACH) ||
		errors.Is(err, syscall.EHOSTUNREACH) {
		return Retriable
	}

	return NonRetriable
}
