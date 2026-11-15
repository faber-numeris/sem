package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/faber-numeris/sem/api/auth"
	"github.com/faber-numeris/sem/api/auth/headlessidentityapidemo"
	"github.com/faber-numeris/sem/api/subscriptionmanagement/queryorgpostmanrunner"
	"github.com/faber-numeris/sem/internal/core/domain"
)

const jsonContentType = "application/json; charset=UTF-8"

func tokenRequestFromAPI(body *auth.TokenRequest, reqCtx RequestContext) domain.TokenRequest {
	req := domain.TokenRequest{
		InstanceURL: reqCtx.InstanceURL,
	}
	if body != nil {
		if body.GrantType != nil {
			req.GrantType = string(*body.GrantType)
		}
		req.ClientID = stringValue(body.ClientId)
		req.ClientSecret = stringValue(body.ClientSecret)
		req.Username = stringValue(body.Username)
		req.Password = stringValue(body.Password)
		req.Code = stringValue(body.Code)
		req.RedirectURI = stringValue(body.RedirectUri)
		req.RefreshToken = stringValue(body.RefreshToken)
		req.Assertion = stringValue(body.Assertion)
		req.Scope = stringValue(body.Scope)
	}
	if req.ClientID == "" && reqCtx.BasicAuthSet {
		req.ClientID = reqCtx.BasicUsername
		req.ClientSecret = reqCtx.BasicPassword
	}
	return req
}

func tokenResponseToAPI(tok *domain.OAuthToken) auth.TokenResponse {
	resp := auth.TokenResponse{
		AccessToken: tok.AccessToken,
		InstanceUrl: tok.InstanceURL,
		Id:          tok.IDURL,
		TokenType:   tok.TokenType,
		IssuedAt:    strconv.FormatInt(tok.IssuedAt, 10),
		Signature:   tok.Signature,
		Scope:       tok.Scope,
	}
	if tok.RefreshToken != "" {
		resp.RefreshToken = stringPointer(tok.RefreshToken)
	}
	return resp
}

func authorizeRequestFromAPI(params headlessidentityapidemo.GetServicesOauth2AuthorizeParams) domain.AuthorizeRequest {
	return domain.AuthorizeRequest{
		ResponseType: stringValue(params.ResponseType),
		ClientID:     stringValue(params.ClientId),
		RedirectURI:  stringValue(params.RedirectUri),
		Scope:        stringValue(params.Scope),
		State:        stringValue(params.State),
	}
}

func authorizeResponseToAPI(res *domain.AuthorizeResult, location string) headlessidentityapidemo.AuthorizeResponse {
	return headlessidentityapidemo.AuthorizeResponse{
		Code:        res.Code,
		AccessToken: res.AccessToken,
		RedirectUri: location,
		State:       res.State,
	}
}

func introspectResultToAPI(res *domain.IntrospectResult) auth.IntrospectResponse {
	return auth.IntrospectResponse{
		Active:    res.Active,
		Scope:     stringPointer(res.Scope),
		ClientId:  stringPointer(res.ClientID),
		Username:  stringPointer(res.Username),
		TokenType: stringPointer(res.TokenType),
		Exp:       int64Pointer(res.Exp),
		Iat:       int64Pointer(res.Iat),
		Sub:       stringPointer(res.Sub),
	}
}

func userInfoToAPI(info *domain.UserInfoResult) queryorgpostmanrunner.UserInfoResponse {
	return queryorgpostmanrunner.UserInfoResponse{
		Sub:               info.Sub,
		UserId:            info.UserID,
		OrganizationId:    info.OrganizationID,
		PreferredUsername: info.PreferredUsername,
		Nickname:          info.Nickname,
		Name:              info.Name,
		Email:             info.Email,
		EmailVerified:     info.EmailVerified,
		GivenName:         info.GivenName,
		FamilyName:        info.FamilyName,
		Zoneinfo:          info.Zoneinfo,
		Photos:            info.Photos,
		Profile:           info.Profile,
		Picture:           info.Picture,
		Urls:              info.Urls,
		Active:            info.Active,
		UserType:          info.UserType,
		Language:          info.Language,
		Locale:            info.Locale,
		UtcOffset:         info.UTCOffset,
		UpdatedAt:         info.UpdatedAt,
	}
}

func apiVersionsToAPI(versions []domain.APIVersion) []queryorgpostmanrunner.APIVersion {
	if versions == nil {
		return nil
	}
	out := make([]queryorgpostmanrunner.APIVersion, 0, len(versions))
	for _, version := range versions {
		out = append(out, queryorgpostmanrunner.APIVersion{
			Label:   version.Label,
			Url:     version.URL,
			Version: version.Version,
		})
	}
	return out
}

func salesforceError(err error) (queryorgpostmanrunner.SalesforceError, int) {
	status, code := salesforceErrorStatus(err)
	return queryorgpostmanrunner.SalesforceError{{
		Message:   err.Error(),
		ErrorCode: code,
		Fields:    []string{},
	}}, status
}

func salesforceErrorStatus(err error) (int, string) {
	status := http.StatusBadRequest
	code := "INVALID_REQUEST"

	switch {
	case errors.Is(err, domain.ErrNotFound), errors.Is(err, domain.ErrInvalidVersion), errors.Is(err, domain.ErrUnsupportedSObject):
		status = http.StatusNotFound
		code = "NOT_FOUND"
	case errors.Is(err, domain.ErrRequiredFieldMissing):
		status = http.StatusBadRequest
		code = "REQUIRED_FIELD_MISSING"
	case errors.Is(err, domain.ErrMalformedQuery):
		status = http.StatusBadRequest
		code = "MALFORMED_QUERY"
	case errors.Is(err, domain.ErrMalformedSearch):
		status = http.StatusBadRequest
		code = "MALFORMED_SEARCH"
	case errors.Is(err, domain.ErrInvalidToken), errors.Is(err, domain.ErrInvalidClient), errors.Is(err, domain.ErrInvalidGrant):
		status = http.StatusUnauthorized
		code = "INVALID_SESSION_ID"
	}
	return status, code
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func stringPointer(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func int64Pointer(value int64) *int64 {
	if value == 0 {
		return nil
	}
	return &value
}
