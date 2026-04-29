package telemetry

import (
	"fmt"
	"os"
	"strings"
)

const (
	BackendCloudWatch = "cloudwatch"

	EnvBackend                 = "TELEMETRY_BACKEND"
	EnvCloudWatchNamespace     = "TELEMETRY_CLOUDWATCH_NAMESPACE"
	EnvCloudWatchRegion        = "TELEMETRY_CLOUDWATCH_REGION"
	DefaultCloudWatchNamespace = "Corecheck"
)

type Tag struct {
	Key   string
	Value string
}

func NewTag(key, value string) Tag {
	return Tag{Key: key, Value: value}
}

type Client interface {
	Metric(name string, value float64, tags ...Tag)
}

type noopClient struct{}

func (noopClient) Metric(name string, value float64, tags ...Tag) {}

var defaultClient Client = noopClient{}

func NewClientFromEnv() (Client, error) {
	backend, err := backendFromEnv()
	if err != nil {
		return nil, err
	}

	switch backend {
	case BackendCloudWatch:
		return NewCloudWatchClientFromEnv()
	default:
		return nil, fmt.Errorf("unsupported telemetry backend %q", backend)
	}
}

func ConfigureDefaultFromEnv() error {
	client, err := NewClientFromEnv()
	if err != nil {
		return err
	}

	SetDefault(client)
	return nil
}

func Default() Client {
	return defaultClient
}

func SetDefault(client Client) {
	if client == nil {
		defaultClient = noopClient{}
		return
	}

	defaultClient = client
}

func Metric(name string, value float64, tags ...Tag) {
	defaultClient.Metric(name, value, tags...)
}

func backendFromEnv() (string, error) {
	backend := strings.ToLower(strings.TrimSpace(os.Getenv(EnvBackend)))
	if backend == "" {
		return BackendCloudWatch, nil
	}

	switch backend {
	case BackendCloudWatch:
		return backend, nil
	default:
		return "", fmt.Errorf("unsupported telemetry backend %q", backend)
	}
}
