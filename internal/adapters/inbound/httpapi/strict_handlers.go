package httpapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/faber-numeris/sem/api/auth"
	"github.com/faber-numeris/sem/api/auth/headlessidentityapidemo"
	"github.com/faber-numeris/sem/api/subscriptionmanagement/queryorgpostmanrunner"
	"github.com/faber-numeris/sem/internal/core/domain"
	"github.com/faber-numeris/sem/internal/ports/inbound"
)

var (
	_                 auth.StrictServerInterface                    = (*strictHandler)(nil)
	_                 headlessidentityapidemo.StrictServerInterface = (*strictHandler)(nil)
	_                 queryorgpostmanrunner.StrictServerInterface   = (*strictHandler)(nil)
	errNotImplemented                                               = errors.New("not implemented")
)

type strictHandler struct {
	svc inbound.EmulatorService
}

func strictRequestErrorHandler(w http.ResponseWriter, _ *http.Request, err error) {
	if errors.Is(err, io.EOF) {
		err = errors.New("request body is empty")
	}
	writeJSON(w, http.StatusBadRequest, map[string]string{
		"error":             "invalid_request",
		"error_description": err.Error(),
	})
}

func strictResponseErrorHandler(w http.ResponseWriter, _ *http.Request, err error) {
	if errors.Is(err, errNotImplemented) {
		writeJSON(w, http.StatusNotImplemented, map[string]string{
			"error":             "not_implemented",
			"error_description": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{
		"error":             "server_error",
		"error_description": err.Error(),
	})
}

func (s *strictHandler) PostServicesOauth2Token(ctx context.Context, request auth.PostServicesOauth2TokenRequestObject) (auth.PostServicesOauth2TokenResponseObject, error) {
	var body *auth.TokenRequest
	switch {
	case request.JSONBody != nil:
		body = request.JSONBody
	case request.FormdataBody != nil:
		body = request.FormdataBody
	}
	tok, err := s.svc.IssueToken(ctx, tokenRequestFromAPI(body, RequestContextFrom(ctx)))
	if err != nil {
		return auth.PostServicesOauth2Token400JSONResponse{
			Body:    authOAuthError("invalid_grant", err),
			Headers: auth.PostServicesOauth2Token400ResponseHeaders{ContentType: stringPointer(jsonContentType)},
		}, nil
	}
	return auth.PostServicesOauth2Token200JSONResponse{
		Body:    tokenResponseToAPI(tok),
		Headers: auth.PostServicesOauth2Token200ResponseHeaders{ContentType: stringPointer(jsonContentType)},
	}, nil
}

func (s *strictHandler) PostServicesOauth2Revoke(ctx context.Context, request auth.PostServicesOauth2RevokeRequestObject) (auth.PostServicesOauth2RevokeResponseObject, error) {
	token := ""
	if request.Params.Token != nil {
		token = *request.Params.Token
	} else if request.Body != nil {
		token = stringValue(request.Body.Token)
	}
	if err := s.svc.RevokeToken(ctx, token); err != nil {
		return auth.PostServicesOauth2Revoke400JSONResponse{
			Body:    authOAuthError("invalid_token", err),
			Headers: auth.PostServicesOauth2Revoke400ResponseHeaders{ContentType: stringPointer(jsonContentType)},
		}, nil
	}
	return auth.PostServicesOauth2Revoke200Response{}, nil
}

func (s *strictHandler) PostServicesOauth2Introspect(ctx context.Context, request auth.PostServicesOauth2IntrospectRequestObject) (auth.PostServicesOauth2IntrospectResponseObject, error) {
	token := ""
	if request.Params.Token != nil {
		token = *request.Params.Token
	} else if request.Body != nil {
		token = stringValue(request.Body.Token)
	}
	res, _ := s.svc.IntrospectToken(ctx, token)
	if res == nil {
		res = &domain.IntrospectResult{}
	}
	return auth.PostServicesOauth2Introspect200JSONResponse{
		Body:    introspectResultToAPI(res),
		Headers: auth.PostServicesOauth2Introspect200ResponseHeaders{ContentType: stringPointer(jsonContentType)},
	}, nil
}

func (s *strictHandler) PostServicesOauth2Register(context.Context, auth.PostServicesOauth2RegisterRequestObject) (auth.PostServicesOauth2RegisterResponseObject, error) {
	return nil, errNotImplemented
}

func (s *strictHandler) GetWellKnownAuthConfiguration(context.Context, auth.GetWellKnownAuthConfigurationRequestObject) (auth.GetWellKnownAuthConfigurationResponseObject, error) {
	return nil, errNotImplemented
}

func (s *strictHandler) GetWellKnownOpenidConfiguration(context.Context, auth.GetWellKnownOpenidConfigurationRequestObject) (auth.GetWellKnownOpenidConfigurationResponseObject, error) {
	return nil, errNotImplemented
}

func (s *strictHandler) GetIdKeys(context.Context, auth.GetIdKeysRequestObject) (auth.GetIdKeysResponseObject, error) {
	return nil, errNotImplemented
}

func (s *strictHandler) GetIdOrgIdUserId(context.Context, auth.GetIdOrgIdUserIdRequestObject) (auth.GetIdOrgIdUserIdResponseObject, error) {
	return nil, errNotImplemented
}

func (s *strictHandler) GetServicesOauth2Authorize(ctx context.Context, request headlessidentityapidemo.GetServicesOauth2AuthorizeRequestObject) (headlessidentityapidemo.GetServicesOauth2AuthorizeResponseObject, error) {
	res, err := s.svc.Authorize(ctx, authorizeRequestFromAPI(request.Params))
	if err != nil {
		return headlessidentityapidemo.GetServicesOauth2Authorize400JSONResponse{
			Body:    headlessOAuthError("invalid_request", err),
			Headers: headlessidentityapidemo.GetServicesOauth2Authorize400ResponseHeaders{ContentType: stringPointer(jsonContentType)},
		}, nil
	}
	location := authorizeLocation(res)
	return headlessidentityapidemo.GetServicesOauth2Authorize302JSONResponse{
		Body: authorizeResponseToAPI(res, location),
		Headers: headlessidentityapidemo.GetServicesOauth2Authorize302ResponseHeaders{
			ContentType: stringPointer(jsonContentType),
			Location:    stringPointer(location),
		},
	}, nil
}

func (s *strictHandler) PostServicesOauth2Authorize(context.Context, headlessidentityapidemo.PostServicesOauth2AuthorizeRequestObject) (headlessidentityapidemo.PostServicesOauth2AuthorizeResponseObject, error) {
	return nil, errNotImplemented
}

func (s *strictHandler) PostServicesAuthHeadlessInitRegistration(context.Context, headlessidentityapidemo.PostServicesAuthHeadlessInitRegistrationRequestObject) (headlessidentityapidemo.PostServicesAuthHeadlessInitRegistrationResponseObject, error) {
	return nil, errNotImplemented
}

func (s *strictHandler) PostServicesAuthHeadlessForgotPassword(context.Context, headlessidentityapidemo.PostServicesAuthHeadlessForgotPasswordRequestObject) (headlessidentityapidemo.PostServicesAuthHeadlessForgotPasswordResponseObject, error) {
	return nil, errNotImplemented
}

func (s *strictHandler) PostServicesAuthHeadlessInitPasswordlessLogin(context.Context, headlessidentityapidemo.PostServicesAuthHeadlessInitPasswordlessLoginRequestObject) (headlessidentityapidemo.PostServicesAuthHeadlessInitPasswordlessLoginResponseObject, error) {
	return nil, errNotImplemented
}

func (s *strictHandler) GetServicesData(context.Context, queryorgpostmanrunner.GetServicesDataRequestObject) (queryorgpostmanrunner.GetServicesDataResponseObject, error) {
	return queryorgpostmanrunner.GetServicesData200JSONResponse{
		Body:    apiVersionsToAPI(s.svc.ListVersions()),
		Headers: queryorgpostmanrunner.GetServicesData200ResponseHeaders{ContentType: stringPointer(jsonContentType)},
	}, nil
}

func (s *strictHandler) PostServicesDataVVersionCommercePaymentsPaymentMethods(context.Context, queryorgpostmanrunner.PostServicesDataVVersionCommercePaymentsPaymentMethodsRequestObject) (queryorgpostmanrunner.PostServicesDataVVersionCommercePaymentsPaymentMethodsResponseObject, error) {
	return nil, errNotImplemented
}

func (s *strictHandler) GetServicesOauth2Userinfo(ctx context.Context, request queryorgpostmanrunner.GetServicesOauth2UserinfoRequestObject) (queryorgpostmanrunner.GetServicesOauth2UserinfoResponseObject, error) {
	info, err := s.svc.GetUserInfo(ctx, stringValue(request.Params.Authorization), RequestContextFrom(ctx).UserInfoBaseURL)
	if err != nil {
		payload, status := salesforceError(err)
		switch status {
		case http.StatusUnauthorized:
			return queryorgpostmanrunner.GetServicesOauth2Userinfo401JSONResponse{
				Body:    payload,
				Headers: queryorgpostmanrunner.GetServicesOauth2Userinfo401ResponseHeaders{ContentType: stringPointer(jsonContentType)},
			}, nil
		case http.StatusNotFound:
			return queryorgpostmanrunner.GetServicesOauth2Userinfo404JSONResponse{
				Body:    payload,
				Headers: queryorgpostmanrunner.GetServicesOauth2Userinfo404ResponseHeaders{ContentType: stringPointer(jsonContentType)},
			}, nil
		default:
			return queryorgpostmanrunner.GetServicesOauth2Userinfo400JSONResponse{
				Body:    payload,
				Headers: queryorgpostmanrunner.GetServicesOauth2Userinfo400ResponseHeaders{ContentType: stringPointer(jsonContentType)},
			}, nil
		}
	}
	return queryorgpostmanrunner.GetServicesOauth2Userinfo200JSONResponse{
		Body:    userInfoToAPI(info),
		Headers: queryorgpostmanrunner.GetServicesOauth2Userinfo200ResponseHeaders{ContentType: stringPointer(jsonContentType)},
	}, nil
}

func authorizeLocation(res *domain.AuthorizeResult) string {
	location := res.RedirectURI
	if res.Code != "" {
		separator := "?"
		if strings.Contains(location, "?") {
			separator = "&"
		}
		location = fmt.Sprintf("%s%scode=%s", location, separator, res.Code)
		if res.State != "" {
			location += "&state=" + res.State
		}
	} else if res.AccessToken != "" {
		location = fmt.Sprintf("%s#access_token=%s&token_type=Bearer", location, res.AccessToken)
	}
	return location
}

func authOAuthError(code string, err error) auth.OAuthError {
	description := err.Error()
	return auth.OAuthError{Error: code, ErrorDescription: &description}
}

func headlessOAuthError(code string, err error) headlessidentityapidemo.OAuthError {
	description := err.Error()
	return headlessidentityapidemo.OAuthError{Error: code, ErrorDescription: &description}
}
