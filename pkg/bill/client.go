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

const SandboxBaseURL = "https://api-sandbox.bill.com/api/v2"
const BaseURL = "https://api.bill.com/api/v2"

// usual format: <API_Base_URL>/Crud/<Operation>/<Entity>.json (Read more at https://developer.bill.com/docs/api-request-format)
const LoginBaseURL = BaseURL + "/Login.json"
const UsersBaseURL = BaseURL + "/List/User.json"
const OrganizationsBaseURL = BaseURL + "/ListOrgs.json"
const UserRoleProfileBaseURL = BaseURL + "/Crud/Read/Profile.json"
const UserRoleProfilesBaseURL = BaseURL + "/List/Profile.json"
const UserRolePermissionsBaseURL = BaseURL + "/GetProfilePermissions.json"
const ApiSessionBaseURL = BaseURL + "/GetSessionInfo.json"

type Credentials struct {
	Username       string
	Password       string
	OrganizationId string
	DeveloperKey   string
	SessionId      string
}

type Client struct {
	httpClient *http.Client
	Credentials
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

func NewClient(httpClient *http.Client, credentials Credentials) *Client {
	return &Client{
		httpClient:  httpClient,
		Credentials: credentials,
	}
}

// Login logs the user into specific organization and returns the session id and organization id.
func (c *Client) Login(ctx context.Context, organizationId string) error {
	var loginResponse LoginResponse

	// Setup required organization id for client
	c.Credentials.OrganizationId = organizationId

	err := c.doRequest(ctx, UsersBaseURL, &loginResponse, c.Credentials, nil, nil)

	if err != nil {
		return err
	}

	if IsInvalidResponse(loginResponse) {
		return status.Error(400, "Request failed")
	}

	// modify the client credentials to involve new session id and organization id
	c.Credentials.SessionId = loginResponse.Data.SessionId
	c.Credentials.OrganizationId = loginResponse.Data.OrgId

	return nil
}

// GetOrganization returns detail information about the organization.
// This operation does not require Login to be called first.
func (c *Client) GetOrganizations(ctx context.Context) ([]Organization, error) {
	var organizationsResponse OrganizationsResponse

	err := c.doRequest(
		ctx,
		OrganizationsBaseURL,
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
		UsersBaseURL,
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
		UsersBaseURL,
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
		UserRoleProfilesBaseURL,
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
		UserRoleProfileBaseURL,
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
		UserRolePermissionsBaseURL,
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
		return status.Error(codes.Code(rawResponse.StatusCode), "Request failed")
	}

	// TODO: check if this works because in case of error, the users in data field
	//  	 won't be present, ther will be an error data instead
	if err := json.NewDecoder(rawResponse.Body).Decode(&resourceResponse); err != nil {
		return err
	}

	return nil
}
