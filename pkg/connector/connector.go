package connector

import (
	"context"

	"github.com/ConductorOne/baton-bill/pkg/bill"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	resourceTypeOrganization = &v2.ResourceType{
		Id:          "organization",
		DisplayName: "Organization",
	}
	resourceTypeUser = &v2.ResourceType{
		Id:          "user",
		DisplayName: "User",
		Traits: []v2.ResourceType_Trait{
			v2.ResourceType_TRAIT_USER,
		},
	}
	resourceTypeRole = &v2.ResourceType{
		Id:          "role",
		DisplayName: "Role",
		Traits: []v2.ResourceType_Trait{
			v2.ResourceType_TRAIT_ROLE,
		},
	}
)

type Bill struct {
	client *bill.Client
	orgs   []string
}

func (b *Bill) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		organizationBuilder(b.client, b.orgs),
		userBuilder(b.client),
		roleBuilder(b.client),
	}
}

func (b *Bill) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "Bill",
	}, nil
}

func (b *Bill) Validate(ctx context.Context) (annotations.Annotations, error) {
	// TODO: research if this is run after `New` since we would have to login here.
	_, err := b.client.GetSessionDetails(ctx)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Provided Access Token is invalid")
	}

	return nil, nil
}

// New returns the Bill connector.
func New(ctx context.Context, organizationIds []string, credentials bill.Credentials) (*Bill, error) {
	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, ctxzap.Extract(ctx)))

	if err != nil {
		return nil, err
	}

	billClient := bill.NewClient(httpClient, credentials)

	return &Bill{
		client: billClient,
		orgs:   organizationIds,
	}, nil
}
