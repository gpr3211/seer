package discovery

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
)

type Registry interface {
	Register(ctx context.Context, instanceID string, serviceName string, hostPort string) error

	Deregister(ctx context.Context, instanceID string, serviceName string, hostPort string) error

	ServiceAddresses(ctx context.Context, serviceID string) ([]string, error)

	// ReportHealthyState push mechanism for reporting healthy state to registry
	ReportHealthyState(instanceID string, serviceName string) error
}

var ErrNotFound = errors.New("No Service addresses are found")

func GenerateInstanceID(serviceName string) string {
	return fmt.Sprintf("%s-%d", serviceName, rand.IntN(1001))

}
