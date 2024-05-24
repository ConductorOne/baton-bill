package bill

import (
	"net/url"
	"strconv"
)

type RequestOption interface {
	Apply(*url.Values)
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

func IsInvalidResponse[T any](response BaseResponse[T]) bool {
	return response.Status == 1 || response.Message == "Error"
}
