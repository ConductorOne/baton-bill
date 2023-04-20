package bill

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
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

type LoginResponse struct {
	BaseResponse[LoginData]
}

type UsersResponse struct {
	BaseResponse[[]User]
}

type SessionDetailsResponse struct {
	BaseResponse[SessionDetails]
}

type OrganizationsResponse struct {
	BaseResponse[[]Organization]
}

type UserRoleProfileResponse struct {
	BaseResponse[UserRoleProfile]
}

type UserRoleProfilesResponse struct {
	BaseResponse[[]UserRoleProfile]
}

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

	if loginResponse.Status == 1 || loginResponse.Message == "Error" {
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

	if organizationsResponse.Status == 1 || organizationsResponse.Message == "Error" {
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

	if sessionDetailsResponse.Status == 1 || sessionDetailsResponse.Message == "Error" {
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

	if usersResponse.Status == 1 || usersResponse.Message == "Error" {
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

	if userRoleProfilesResponse.Status == 1 || userRoleProfilesResponse.Message == "Error" {
		return nil, 0, status.Error(400, "Request failed")
	}

	return userRoleProfilesResponse.Data, getUserRoleProfilesVars.Start + getUserRoleProfilesVars.Max, nil
}

// GetUserRoleProfile returns detail information about the user role under provided id
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

	if userRoleProfileResponse.Status == 1 || userRoleProfileResponse.Message == "Error" {
		return UserRoleProfile{}, status.Error(400, "Request failed")
	}

	return userRoleProfileResponse.Data, nil
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

// Method Apply for Credentials struct adds credentials to the request body.
func (credentials Credentials) Apply(body *url.Values) {
	// add username (required for login)
	if credentials.Username != "" {
		body.Set("userName", credentials.Username)
	}

	// add password (required for login)
	if credentials.Password != "" {
		body.Set("password", credentials.Password)
	}

	// add organization id (required for login)
	if credentials.OrganizationId != "" {
		body.Set("orgId", credentials.OrganizationId)
	}

	// add developer key (required for login)
	if credentials.DeveloperKey != "" {
		body.Set("devKey", credentials.DeveloperKey)
	}

	// add session id
	if credentials.SessionId != "" {
		body.Set("sessionId", credentials.SessionId)
	}
}

// Method Apply for PaginationParams struct adds pagination parameters to the request body.
func (pagination PaginationParams) Apply(body *url.Values) {
	data := url.Values{}

	// add max reference
	if pagination.Max != 0 {
		data.Set("max", strconv.Itoa(pagination.Max))
	}

	// add start reference
	if pagination.Start != 0 {
		data.Set("start", strconv.Itoa(pagination.Start))
	}

	body.Set("data", data.Encode())
}

// Method Apply for SearchParams struct adds search parameters (like id of the resource) to the request body.
// In case of Bill.com API, it uses the data field to pass the search parameters as well as the pagination parameters
// So this function handles both cases.
func (searchParams SearchParams) Apply(body *url.Values) {
	// add Id if provided
	if searchParams.Id != "" {
		// check if the data field is already set
		if body.Has("data") {
			data, err := url.ParseQuery(body.Get("data"))
			if err != nil {
				return
			}

			data.Set("id", searchParams.Id)
			body.Set("data", data.Encode())
		} else {
			data := url.Values{}

			data.Set("id", searchParams.Id)
			body.Set("data", data.Encode())
		}
	}

	// TODO: add filtering and sorting?
}
