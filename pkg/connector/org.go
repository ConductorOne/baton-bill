package connector

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ConductorOne/baton-bill/pkg/bill"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	grant "github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
)

const orgRoleMember = "member"

type organizationResourceType struct {
	resourceType *v2.ResourceType
	client       *bill.Client
	orgs         map[string]*bill.Organization
}

func (o *organizationResourceType) ResourceType(_ context.Context) *v2.ResourceType {
	return o.resourceType
}

// Create a new connector resource for a Bill Organization.
func organizationResource(ctx context.Context, organization *bill.Organization, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	resource, err := rs.NewResource(
		organization.Name,
		resourceTypeOrganization,
		organization.Id,
		rs.WithParentResourceID(parentResourceID),
		rs.WithAnnotation(
			&v2.ChildResourceType{ResourceTypeId: resourceTypeUser.Id},
		),
	)

	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (o *organizationResourceType) List(ctx context.Context, parentId *v2.ResourceId, _ *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	// Listing organization in Bill does not support pagination
	organizations, err := o.client.GetOrganizations(ctx)
	if err != nil {
		return nil, "", nil, fmt.Errorf("bill-connector: failed to list organizations: %w", err)
	}

	var rv []*v2.Resource
	for _, organization := range organizations {
		if _, ok := o.orgs[organization.Id]; !ok && len(o.orgs) > 0 {
			continue
		}

		// Login to specific organization in Bill
		err = o.client.Login(ctx, organization.Id)
		if err != nil {
			return nil, "", nil, fmt.Errorf("bill-connector: failed to login to organization: %w", err)
		}

		organizationCopy := organization
		or, err := organizationResource(ctx, &organizationCopy, parentId)

		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, or)
	}

	return rv, "", nil, nil
}

func (o *organizationResourceType) Entitlements(ctx context.Context, resource *v2.Resource, token *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	// add membership entitlement just once for the organization (in case of pagination)
	// TODO: Research if in case Bill does not support pagination in organizations listing we can remove handling of pagination in this method
	if token == nil || token.Token == "" {
		assignmentOptions := []ent.EntitlementOption{
			ent.WithGrantableTo(resourceTypeUser),
			ent.WithDisplayName(fmt.Sprintf("%s Org %s", resource.DisplayName, orgRoleMember)),
			ent.WithDescription(fmt.Sprintf("Organization %s membership role in Bill", resource.DisplayName)),
		}

		rv = append(rv, ent.NewAssignmentEntitlement(
			resource,
			orgRoleMember,
			assignmentOptions...,
		))
	}

	page, err := handlePageToken(token)
	if err != nil {
		return nil, "", nil, err
	}

	orgAccessRoles, nextPage, err := o.client.GetUserRoleProfiles(ctx, bill.PaginationParams{Start: page, Max: ResourcesPageSize})
	if err != nil {
		return nil, "", nil, fmt.Errorf("bill-connector: failed to get user roles: %w", err)
	}

	for _, role := range orgAccessRoles {
		permissionOptions := []ent.EntitlementOption{
			ent.WithGrantableTo(resourceTypeUser),
			ent.WithDisplayName(fmt.Sprintf("%s Org %s", resource.DisplayName, titleCaser.String(role.Name))),
			ent.WithDescription(fmt.Sprintf("Organization %s role in Bill: %s", resource.DisplayName, role.Description)),
		}

		// Create a new entitlement for the user role.
		rv = append(rv, ent.NewPermissionEntitlement(
			resource,
			role.Name,
			permissionOptions...,
		))
	}

	return rv, strconv.Itoa(nextPage), nil, nil
}

func (o *organizationResourceType) Grants(ctx context.Context, resource *v2.Resource, token *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	page, err := handlePageToken(token)
	if err != nil {
		return nil, "", nil, err
	}

	users, nextPage, err := o.client.GetUsers(ctx, bill.PaginationParams{
		Start: page,
		Max:   ResourcesPageSize,
	})
	if err != nil {
		return nil, "", nil, fmt.Errorf("bill-connector: failed to get users: %w", err)
	}

	var rv []*v2.Grant
	for _, user := range users {
		role, err := o.client.GetUserRoleProfile(ctx, user.RoleId)
		if err != nil {
			return nil, "", nil, err
		}

		userCopy := user
		ur, err := userResource(ctx, &userCopy, nil)
		if err != nil {
			return nil, "", nil, err
		}

		// Create a new grant for the user role.
		rv = append(rv, grant.NewGrant(
			resource,
			role.Name,
			ur.Id,
		))

		// Create a new grant for the user membership role.
		rv = append(rv, grant.NewGrant(
			resource,
			orgRoleMember,
			ur.Id,
		))
	}

	return rv, strconv.Itoa(nextPage), nil, nil
}

func organizationBuilder(client *bill.Client, organizationIds []string) *organizationResourceType {
	orgsMap := make(map[string]*bill.Organization)

	for _, orgId := range organizationIds {
		orgsMap[orgId] = &bill.Organization{}
	}

	return &organizationResourceType{
		resourceType: resourceTypeOrganization,
		client:       client,
		orgs:         orgsMap,
	}
}
