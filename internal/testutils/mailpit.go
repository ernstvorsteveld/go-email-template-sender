package testutils

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type MailpitContainer struct {
	SMTPHost string
	SMTPPort int
	APIURL   string
	Cleanup  func()
}

func SetupMailpit(ctx context.Context) (*MailpitContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "axllent/mailpit:latest",
		ExposedPorts: []string{"1025/tcp", "8025/tcp"},
		WaitingFor:   wait.ForHTTP("/api/v1/info").WithPort("8025/tcp").WithStartupTimeout(15 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	smtpPort, err := container.MappedPort(ctx, "1025/tcp")
	if err != nil {
		return nil, err
	}

	apiPort, err := container.MappedPort(ctx, "8025/tcp")
	if err != nil {
		return nil, err
	}

	cleanup := func() {
		container.Terminate(context.Background())
	}

	portInt, _ := strconv.Atoi(smtpPort.Port())

	return &MailpitContainer{
		SMTPHost: host,
		SMTPPort: portInt,
		APIURL:   fmt.Sprintf("http://%s:%s/api/v1", host, apiPort.Port()),
		Cleanup:  cleanup,
	}, nil
}
