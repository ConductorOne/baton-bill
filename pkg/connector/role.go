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

const roleMember = "member"

type roleResourceType struct {
	resourceType *v2.ResourceType
	client       *bill.Client
}

func (o *roleResourceType) ResourceType(_ context.Context) *v2.ResourceType {
	return o.resourceType
}

// Create a new connector resource for an Bill User Profile Role.
func roleResource(ctx context.Context, role *bill.UserRoleProfile, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"role_id":   role.Id,
		"role_name": role.Name,
	}

	roleTraitOptions := []rs.RoleTraitOption{
		rs.WithRoleProfile(profile),
	}

	resource, err := rs.NewRoleResource(
		role.Name,
		resourceTypeRole,
		role.Id,
		roleTraitOptions,
		rs.WithParentResourceID(parentResourceID),
	)

	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (o *roleResourceType) List(ctx context.Context, parentId *v2.ResourceId, token *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	if parentId == nil {
		return nil, "", nil, nil
	}

	page, err := handlePageToken(token)
	if err != nil {
		return nil, "", nil, err
	}

	orgAccessRoles, nextPage, err := o.client.GetUserRoleProfiles(
		ctx,
		bill.PaginationParams{Max: ResourcesPageSize, Start: page},
	)
	if err != nil {
		return nil, "", nil, fmt.Errorf("bill-connector: failed to list user roles: %w", err)
	}

	var rv []*v2.Resource
	for _, role := range orgAccessRoles {
		roleCopy := role

		rr, err := roleResource(ctx, &roleCopy, parentId)
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, rr)
	}

	return rv, strconv.Itoa(nextPage), nil, nil
}

func (o *roleResourceType) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	// add membership entitlement (later use for user granting roles)
	assignmentOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(resourceTypeUser),
		ent.WithDisplayName(fmt.Sprintf("%s Role %s", resource.DisplayName, titleCaser.String(orgRoleMember))),
		ent.WithDescription(fmt.Sprintf("%s Bill.com Role", resource.DisplayName)),
	}

	rv = append(rv, ent.NewAssignmentEntitlement(
		resource,
		orgRoleMember,
		assignmentOptions...,
	))

	// parse the role id from profile
	roleTrait, err := rs.GetRoleTrait(resource)
	if err != nil {
		return nil, "", nil, err
	}

	roleId, ok := rs.GetProfileStringValue(roleTrait.Profile, "role_id")
	if !ok {
		return nil, "", nil, fmt.Errorf("bill-connector: failed to get role id from profile")
	}

	// add permissions entitlements
	userRolePermissions, err := o.client.GetUserRolePermissions(ctx, roleId)
	if err != nil {
		return nil, "", nil, fmt.Errorf("bill-connector: failed to get user role permissions: %w", err)
	}

	for pName, pValue := range userRolePermissions {
		// skip if permission is not enabled
		if !pValue {
			continue
		}

		// add entitlement for each permission
		rv = append(rv, ent.NewPermissionEntitlement(
			resource,
			pName,
			ent.WithDisplayName(fmt.Sprintf("%s Permission %s", resource.DisplayName, titleCaser.String(pName))),
			ent.WithDescription(fmt.Sprintf("%s Bill.com Permission", resource.DisplayName)),
		))
	}

	return rv, "", nil, nil
}

func (o *roleResourceType) Grants(ctx context.Context, resource *v2.Resource, token *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	page, err := handlePageToken(token)
	if err != nil {
		return nil, "", nil, err
	}

	// parse the role id from profile
	roleTrait, err := rs.GetRoleTrait(resource)
	if err != nil {
		return nil, "", nil, err
	}

	roleId, ok := rs.GetProfileStringValue(roleTrait.Profile, "role_id")
	if !ok {
		return nil, "", nil, fmt.Errorf("bill-connector: failed to get role id from profile")
	}

	// get all users and add membership grants for each user with the corresponding role
	users, nextPage, err := o.client.GetUsers(ctx, bill.PaginationParams{
		Start: page,
		Max:   ResourcesPageSize,
	})
	if err != nil {
		return nil, "", nil, fmt.Errorf("bill-connector: failed to get users: %w", err)
	}

	var rv []*v2.Grant
	for _, user := range users {
		// skip if user does not have the role
		if user.RoleId != roleId {
			continue
		}

		userCopy := user
		ur, err := userResource(ctx, &userCopy, nil)
		if err != nil {
			return nil, "", nil, err
		}

		// Create a new grant for the user membership on role.
		rv = append(rv, grant.NewGrant(
			resource,
			roleMember,
			ur.Id,
		))
	}

	return rv, strconv.Itoa(nextPage), nil, nil
}

func roleBuilder(client *bill.Client) *roleResourceType {
	return &roleResourceType{
		resourceType: resourceTypeRole,
		client:       client,
	}
}
