package domain

import "time"

type SalesforceUser struct {
	ID               string
	OrganizationID   string
	Username         string
	Email            string
	PasswordHash     string
	Name             string
	FirstName        string
	LastName         string
	Nickname         string
	PictureURL       string
	ProfileID        string
	UserType         string
	Language         string
	Locale           string
	UTCOffset        int
	IsActive         bool
	CreatedDate      time.Time
	LastModifiedDate time.Time
}

type OAuthClient struct {
	ID           string
	ClientID     string
	ClientSecret string
	Name         string
	RedirectURI  string
	Scopes       string
	IsActive     bool
}

type OAuthAuthCode struct {
	Code        string
	ClientID    string
	UserID      string
	RedirectURI string
	Scope       string
	IsUsed      bool
	ExpiresAt   time.Time
}

type OAuthToken struct {
	ID           string
	AccessToken  string
	RefreshToken string
	TokenType    string
	ClientID     string
	UserID       string
	Scope        string
	InstanceURL  string
	IDURL        string
	Signature    string
	IssuedAt     int64
	ExpiresAt    time.Time
	IsRevoked    bool
}

type TokenRequest struct {
	GrantType    string
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
	Code         string
	RedirectURI  string
	RefreshToken string
	Assertion    string
	Scope        string
	InstanceURL  string
}

type AuthorizeRequest struct {
	ResponseType string
	ClientID     string
	RedirectURI  string
	Scope        string
	State        string
}

type AuthorizeResult struct {
	Code        string
	AccessToken string
	RedirectURI string
	State       string
}

type IntrospectResult struct {
	Active    bool   `json:"active"`
	Scope     string `json:"scope,omitempty"`
	ClientID  string `json:"client_id,omitempty"`
	Username  string `json:"username,omitempty"`
	TokenType string `json:"token_type,omitempty"`
	Exp       int64  `json:"exp,omitempty"`
	Iat       int64  `json:"iat,omitempty"`
	Sub       string `json:"sub,omitempty"`
}

type UserInfoResult struct {
	Sub               string            `json:"sub"`
	UserID            string            `json:"user_id"`
	OrganizationID    string            `json:"organization_id"`
	PreferredUsername string            `json:"preferred_username"`
	Nickname          string            `json:"nickname"`
	Name              string            `json:"name"`
	Email             string            `json:"email"`
	EmailVerified     bool              `json:"email_verified"`
	GivenName         string            `json:"given_name"`
	FamilyName        string            `json:"family_name"`
	Zoneinfo          string            `json:"zoneinfo"`
	Photos            map[string]string `json:"photos"`
	Profile           string            `json:"profile"`
	Picture           string            `json:"picture"`
	Urls              map[string]string `json:"urls"`
	Active            bool              `json:"active"`
	UserType          string            `json:"user_type"`
	Language          string            `json:"language"`
	Locale            string            `json:"locale"`
	UTCOffset         int               `json:"utcOffset"`
	UpdatedAt         string            `json:"updated_at"`
}
