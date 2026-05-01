package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ConductorOne/baton-bill/pkg/bill"
	"github.com/ConductorOne/baton-bill/pkg/connector"
	cfg "github.com/ConductorOne/baton-bill/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/connectorrunner"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

var version = "dev"

func main() {
	ctx := context.Background()

	_, cmd, err := config.DefineConfiguration(
		ctx,
		"baton-bill",
		getConnector,
		cfg.Config,
		connectorrunner.WithDefaultCapabilitiesConnectorBuilder(&connector.Bill{}),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cmd.Version = version

	err = cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func getConnector(ctx context.Context, c *cfg.Bill) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)

	billConnector, err := connector.New(
		ctx,
		c.Organizationids,
		bill.Credentials{
			Username:     c.Username,
			Password:     c.Password,
			DeveloperKey: c.Developerkey,
		},
		c.BaseUrl,
	)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	conn, err := connectorbuilder.NewConnector(ctx, billConnector)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	return conn, nil
}
