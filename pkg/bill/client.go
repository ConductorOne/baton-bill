package bill

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const DefaultBaseURL = "https://api.bill.com/api/v2"

type Credentials struct {
	Username       string
	Password       string
	OrganizationId string
	DeveloperKey   string
	SessionId      string
}

type Client struct {
	httpClient *http.Client
	baseURL    string
	Credentials
}

func (c *Client) usersURL() string {
	return c.baseURL + "/List/User.json"
}

func (c *Client) organizationsURL() string {
	return c.baseURL + "/ListOrgs.json"
}

func (c *Client) userRoleProfileURL() string {
	return c.baseURL + "/Crud/Read/Profile.json"
}

func (c *Client) userRoleProfilesURL() string {
	return c.baseURL + "/List/Profile.json"
}

func (c *Client) userRolePermissionsURL() string {
	return c.baseURL + "/GetProfilePermissions.json"
}

type LoginResponse = BaseResponse[LoginData]
type UsersResponse = BaseResponse[[]User]
type SessionDetailsResponse = BaseResponse[SessionDetails]
type OrganizationsResponse = BaseResponse[[]Organization]
type UserRoleProfileResponse = BaseResponse[UserRoleProfile]
type UserRoleProfilesResponse = BaseResponse[[]UserRoleProfile]
type UserRolePermissionsResponse = BaseResponse[map[string]bool]

type UserParams struct {
	PaginationParams
	SearchParams
}

func NewClient(httpClient *http.Client, credentials Credentials, baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		httpClient:  httpClient,
		baseURL:     baseURL,
		Credentials: credentials,
	}
}

// Login logs the user into specific organization and returns the session id and organization id.
func (c *Client) Login(ctx context.Context, organizationId string) error {
	var loginResponse LoginResponse

	// Build the request credentials with the target org id without mutating
	// the client; we only commit the new org/session to c on success so a
	// failed login does not leave the client half-updated.
	reqCreds := c.Credentials
	reqCreds.OrganizationId = organizationId

	err := c.doRequest(ctx, c.usersURL(), &loginResponse, reqCreds, nil, nil)

	if err != nil {
		return err
	}

	if IsInvalidResponse(loginResponse) {
		return status.Error(400, "Request failed")
	}

	c.SessionId = loginResponse.Data.SessionId
	c.OrganizationId = loginResponse.Data.OrgId

	return nil
}

// GetOrganization returns detail information about the organization.
// This operation does not require Login to be called first.
func (c *Client) GetOrganizations(ctx context.Context) ([]Organization, error) {
	var organizationsResponse OrganizationsResponse

	err := c.doRequest(
		ctx,
		c.organizationsURL(),
		&organizationsResponse,
		Credentials{
			DeveloperKey: c.DeveloperKey,
			Username:     c.Username,
			Password:     c.Password,
		},
		nil,
		nil,
	)

	if err != nil {
		return nil, err
	}

	if IsInvalidResponse(organizationsResponse) {
		return nil, status.Error(400, "Request failed")
	}

	return organizationsResponse.Data, nil
}

// GetSessionDetails returns details regarding session of currently signed in user and organization.
func (c *Client) GetSessionDetails(ctx context.Context) (SessionDetails, error) {
	var sessionDetailsResponse SessionDetailsResponse

	err := c.doRequest(
		ctx,
		c.usersURL(),
		&sessionDetailsResponse,
		Credentials{
			DeveloperKey: c.DeveloperKey,
			SessionId:    c.SessionId,
		},
		nil,
		nil,
	)

	if err != nil {
		return SessionDetails{}, err
	}

	if IsInvalidResponse(sessionDetailsResponse) {
		return SessionDetails{}, status.Error(400, "Request failed")
	}

	return sessionDetailsResponse.Data, nil
}

// GetUsers returns all users under the organization account.
func (c *Client) GetUsers(ctx context.Context, getUsersVars PaginationParams) ([]User, int, error) {
	var usersResponse UsersResponse

	err := c.doRequest(
		ctx,
		c.usersURL(),
		&usersResponse,
		Credentials{
			DeveloperKey: c.DeveloperKey,
			SessionId:    c.SessionId,
		},
		getUsersVars,
		nil,
	)

	if err != nil {
		return nil, 0, err
	}

	if IsInvalidResponse(usersResponse) {
		return nil, 0, status.Error(400, "Request failed")
	}

	return usersResponse.Data, getUsersVars.Start + getUsersVars.Max, nil
}

// GetUserRoleProfiles returns all user roles available in the organization.
func (c *Client) GetUserRoleProfiles(ctx context.Context, getUserRoleProfilesVars PaginationParams) ([]UserRoleProfile, int, error) {
	var userRoleProfilesResponse UserRoleProfilesResponse

	err := c.doRequest(
		ctx,
		c.userRoleProfilesURL(),
		&userRoleProfilesResponse,
		Credentials{
			DeveloperKey: c.DeveloperKey,
			SessionId:    c.SessionId,
		},
		getUserRoleProfilesVars,
		nil,
	)

	if err != nil {
		return nil, 0, err
	}

	if IsInvalidResponse(userRoleProfilesResponse) {
		return nil, 0, status.Error(400, "Request failed")
	}

	return userRoleProfilesResponse.Data, getUserRoleProfilesVars.Start + getUserRoleProfilesVars.Max, nil
}

// GetUserRoleProfile returns detail information about the user role under provided id.
func (c *Client) GetUserRoleProfile(ctx context.Context, roleId string) (UserRoleProfile, error) {
	var userRoleProfileResponse UserRoleProfileResponse

	err := c.doRequest(
		ctx,
		c.userRoleProfileURL(),
		&userRoleProfileResponse,
		Credentials{
			DeveloperKey: c.DeveloperKey,
			SessionId:    c.SessionId,
		},
		nil,
		SearchParams{Id: roleId},
	)

	if err != nil {
		return UserRoleProfile{}, err
	}

	if IsInvalidResponse(userRoleProfileResponse) {
		return UserRoleProfile{}, status.Error(400, "Request failed")
	}

	return userRoleProfileResponse.Data, nil
}

// GetUserRolePermissions returns map of permissions under the provided user role.
func (c *Client) GetUserRolePermissions(ctx context.Context, roleId string) (map[string]bool, error) {
	var userRolePermissionsResponse UserRolePermissionsResponse

	err := c.doRequest(
		ctx,
		c.userRolePermissionsURL(),
		&userRolePermissionsResponse,
		Credentials{
			DeveloperKey: c.DeveloperKey,
			SessionId:    c.SessionId,
		},
		nil,
		SearchParams{Id: roleId},
	)

	if err != nil {
		return nil, err
	}

	if IsInvalidResponse(userRolePermissionsResponse) {
		return nil, status.Error(400, "Request failed")
	}

	return userRolePermissionsResponse.Data, nil
}

func (c *Client) doRequest(
	ctx context.Context,
	urlAddress string,
	resourceResponse interface{},
	requestOptions ...RequestOption,
) error {
	requestBody := url.Values{}

	for _, option := range requestOptions {
		if option != nil {
			option.Apply(&requestBody)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlAddress, strings.NewReader(requestBody.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	rawResponse, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer rawResponse.Body.Close()

	if rawResponse.StatusCode >= 300 {
		return status.Error(codes.Code(uint32(rawResponse.StatusCode)), "Request failed") //nolint:gosec // StatusCode is always valid HTTP code
	}

	// TODO: check if this works because in case of error, the users in data field
	//  	 won't be present, ther will be an error data instead
	if err := json.NewDecoder(rawResponse.Body).Decode(&resourceResponse); err != nil {
		return err
	}

	return nil
}
