package bill

type BaseResource struct {
	Id string `json:"id"`
}

type LoginData struct {
	SessionId string `json:"sessionId"`
	OrgId     string `json:"orgId"`
}

type SessionDetails struct {
	OrgId  string `json:"organizationId"`
	UserId string `json:"userId"`
}

type User struct {
	BaseResource
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	IsActive  bool   `json:"isActive"`
	RoleId    string `json:"profileId"`
}

type Organization struct {
	Id   string `json:"orgId"`
	Name string `json:"orgName"`
}

type UserRoleProfile struct {
	BaseResource
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type BaseResponse[T any] struct {
	Status  int    `json:"response_status"`
	Message string `json:"response_message"`
	Data    T      `json:"response_data"`
}

type PaginationParams struct {
	Max   int `json:"max"`
	Start int `json:"start"`
}

type Filter struct {
	Field string      `json:"field"`
	Op    string      `json:"op"`
	Value interface{} `json:"value"`
}

type Sort struct {
	Field string `json:"field"`
	Asc   bool   `json:"asc"`
}

type SearchParams struct {
	Id        string   `json:"id"`
	Filters   []Filter `json:"filters"`
	Sort      []Sort   `json:"sort"`
	Nested    bool     `json:"nested"`
	ShowAudit bool     `json:"showAudit"`
}
