package httpapi

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/faber-numeris/sem/api/auth"
	"github.com/faber-numeris/sem/api/auth/headlessidentityapidemo"
	"github.com/faber-numeris/sem/internal/core/domain"
	"github.com/faber-numeris/sem/internal/mocks"
	"github.com/stretchr/testify/mock"
)

func TestStrictHandler_TokenCharacterization(t *testing.T) {
	svc := mocks.NewMockEmulatorService(t)
	r := NewRouter(svc, nil)

	t.Run("JSON body with charset", func(t *testing.T) {
		svc.EXPECT().IssueToken(mock.Anything, mock.MatchedBy(func(req domain.TokenRequest) bool {
			return req.GrantType == "password" &&
				req.ClientID == "json-client" &&
				req.ClientSecret == "json-secret" &&
				req.Username == "user" &&
				req.Password == "pass" &&
				req.Code == "code" &&
				req.RedirectURI == "http://localhost/cb" &&
				req.RefreshToken == "refresh" &&
				req.Assertion == "jwt" &&
				req.Scope == "full" &&
				req.InstanceURL == "https://example.com"
		})).Return(&domain.OAuthToken{
			AccessToken: "at",
			InstanceURL: "https://example.com",
			IDURL:       "https://example.com/id/1/1",
			TokenType:   "Bearer",
			IssuedAt:    123,
			Signature:   "sig",
			Scope:       "full",
		}, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "https://example.com/services/oauth2/token", strings.NewReader(`{
			"grant_type":"password",
			"client_id":"json-client",
			"client_secret":"json-secret",
			"username":"user",
			"password":"pass",
			"code":"code",
			"redirect_uri":"http://localhost/cb",
			"refresh_token":"refresh",
			"assertion":"jwt",
			"scope":"full"
		}`))
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if got := w.Header().Get("Content-Type"); got != jsonContentType {
			t.Fatalf("unexpected content type %q", got)
		}
		var body map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body["issued_at"] != "123" {
			t.Fatalf("issued_at must be a string, got %#v", body["issued_at"])
		}
		if _, ok := body["refresh_token"]; ok {
			t.Fatalf("refresh_token must be omitted when empty: %s", w.Body.String())
		}
		for _, key := range []string{"access_token", "instance_url", "id", "token_type", "issued_at", "signature", "scope"} {
			if _, ok := body[key]; !ok {
				t.Fatalf("missing %q in %s", key, w.Body.String())
			}
		}
	})

	t.Run("form body with charset and Basic fallback", func(t *testing.T) {
		svc.EXPECT().IssueToken(mock.Anything, mock.MatchedBy(func(req domain.TokenRequest) bool {
			return req.GrantType == "client_credentials" &&
				req.ClientID == "basic-client" &&
				req.ClientSecret == "basic-secret" &&
				req.InstanceURL == "https://example.com"
		})).Return(&domain.OAuthToken{
			AccessToken:  "at2",
			RefreshToken: "rt2",
			InstanceURL:  "https://example.com",
			IDURL:        "https://example.com/id/1/1",
			TokenType:    "Bearer",
			IssuedAt:     456,
			Signature:    "sig2",
			Scope:        "api",
		}, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "https://example.com/services/oauth2/token", strings.NewReader("grant_type=client_credentials"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		req.SetBasicAuth("basic-client", "basic-secret")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body["refresh_token"] != "rt2" {
			t.Fatalf("expected refresh_token, got %s", w.Body.String())
		}
	})

	t.Run("no content type uses Basic auth", func(t *testing.T) {
		svc.EXPECT().IssueToken(mock.Anything, mock.MatchedBy(func(req domain.TokenRequest) bool {
			return req.GrantType == "" && req.ClientID == "basic-only" && req.ClientSecret == "secret"
		})).Return(&domain.OAuthToken{InstanceURL: "http://example.com", IssuedAt: 1}, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/services/oauth2/token", nil)
		req.SetBasicAuth("basic-only", "secret")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("malformed JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/services/oauth2/token", strings.NewReader(`{`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
		assertOAuthError(t, w.Body.Bytes(), "invalid_request")
	})

	t.Run("empty JSON body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/services/oauth2/token", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
		assertOAuthError(t, w.Body.Bytes(), "invalid_request")
	})
}

func TestStrictHandler_AuthErrorsAndUnimplemented(t *testing.T) {
	svc := mocks.NewMockEmulatorService(t)
	r := NewRouter(svc, nil)

	t.Run("token service error", func(t *testing.T) {
		svc.EXPECT().IssueToken(mock.Anything, mock.Anything).Return(nil, domain.ErrInvalidGrant).Once()
		req := httptest.NewRequest(http.MethodPost, "/services/oauth2/token", strings.NewReader("grant_type=bad"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
		assertOAuthError(t, w.Body.Bytes(), "invalid_grant")
	})

	t.Run("revoke query takes precedence and errors", func(t *testing.T) {
		svc.EXPECT().RevokeToken(mock.Anything, "query-token").Return(nil).Once()
		req := httptest.NewRequest(http.MethodPost, "/services/oauth2/revoke?token=query-token", strings.NewReader("token=body-token"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		svc.EXPECT().RevokeToken(mock.Anything, "query-token").Return(domain.ErrInvalidToken).Once()
		w = httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
		assertOAuthError(t, w.Body.Bytes(), "invalid_token")
	})

	t.Run("introspect form body fallback", func(t *testing.T) {
		svc.EXPECT().IntrospectToken(mock.Anything, "form-token").Return(&domain.IntrospectResult{
			Active:    true,
			Scope:     "full",
			ClientID:  "cid",
			Username:  "user",
			TokenType: "access_token",
			Exp:       100,
			Iat:       50,
			Sub:       "sub",
		}, nil).Once()
		req := httptest.NewRequest(http.MethodPost, "/services/oauth2/introspect", strings.NewReader("token=form-token"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var body map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body["active"] != true || body["exp"] != float64(100) || body["iat"] != float64(50) {
			t.Fatalf("unexpected introspection body: %s", w.Body.String())
		}

		svc.EXPECT().IntrospectToken(mock.Anything, "").Return(nil, nil).Once()
		w = httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/services/oauth2/introspect", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var empty map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &empty); err != nil {
			t.Fatal(err)
		}
		if _, ok := empty["exp"]; ok {
			t.Fatalf("zero exp must be omitted: %s", w.Body.String())
		}
	})

	unimplemented := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/services/oauth2/register"},
		{http.MethodGet, "/.well-known/auth-configuration"},
		{http.MethodGet, "/.well-known/openid-configuration"},
		{http.MethodGet, "/id/keys"},
		{http.MethodGet, "/id/00D/005"},
		{http.MethodPost, "/services/oauth2/authorize"},
		{http.MethodPost, "/services/auth/headless/init/registration"},
		{http.MethodPost, "/services/auth/headless/forgot_password"},
		{http.MethodPost, "/services/auth/headless/init/passwordless/login"},
		{http.MethodPost, "/services/data/v60.0/commerce/payments/payment-methods"},
	}
	for _, tc := range unimplemented {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
			if w.Code != http.StatusNotImplemented {
				t.Fatalf("expected 501, got %d: %s", w.Code, w.Body.String())
			}
			assertOAuthError(t, w.Body.Bytes(), "not_implemented")
		})
	}

	t.Run("strict error handlers", func(t *testing.T) {
		w := httptest.NewRecorder()
		strictRequestErrorHandler(w, nil, context.Canceled)
		if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "invalid_request") {
			t.Fatalf("unexpected request error response: %d %s", w.Code, w.Body.String())
		}

		w = httptest.NewRecorder()
		strictResponseErrorHandler(w, nil, errNotImplemented)
		if w.Code != http.StatusNotImplemented {
			t.Fatalf("expected 501, got %d", w.Code)
		}

		w = httptest.NewRecorder()
		strictResponseErrorHandler(w, nil, errors.New("boom"))
		if w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), "server_error") {
			t.Fatalf("unexpected response error response: %d %s", w.Code, w.Body.String())
		}
	})
}

func TestStrictHandler_UserInfoAndData(t *testing.T) {
	svc := mocks.NewMockEmulatorService(t)
	r := NewRouter(svc, nil)

	t.Run("data versions", func(t *testing.T) {
		svc.EXPECT().ListVersions().Return([]domain.APIVersion{
			{Label: "Winter '26", URL: "/services/data/v66.0", Version: "66.0"},
		}).Once()
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if got := w.Header().Get("Content-Type"); got != jsonContentType {
			t.Fatalf("unexpected content type %q", got)
		}
		var versions []map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &versions); err != nil {
			t.Fatal(err)
		}
		if len(versions) != 1 || versions[0]["url"] != "/services/data/v66.0" {
			t.Fatalf("unexpected versions: %s", w.Body.String())
		}
	})

	t.Run("userinfo success", func(t *testing.T) {
		svc.EXPECT().GetUserInfo(mock.Anything, "Bearer at", "http://example.com").Return(&domain.UserInfoResult{
			Sub:               "sub",
			UserID:            "005",
			OrganizationID:    "00D",
			PreferredUsername: "user@example.com",
			Nickname:          "user",
			Name:              "User Name",
			Email:             "user@example.com",
			EmailVerified:     true,
			GivenName:         "User",
			FamilyName:        "Name",
			Zoneinfo:          "America/New_York",
			Photos:            map[string]string{"picture": "p"},
			Profile:           "profile",
			Picture:           "picture",
			Urls:              map[string]string{"rest": "rest"},
			Active:            true,
			UserType:          "STANDARD",
			Language:          "en_US",
			Locale:            "en_US",
			UTCOffset:         -300,
			UpdatedAt:         "2026-01-01T00:00:00Z",
		}, nil).Once()
		req := httptest.NewRequest(http.MethodGet, "http://example.com/services/oauth2/userinfo", nil)
		req.Header.Set("Authorization", "Bearer at")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if got := w.Header().Get("Content-Type"); got != jsonContentType {
			t.Fatalf("unexpected content type %q", got)
		}
		var body map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body["user_id"] != "005" || body["utcOffset"] != float64(-300) {
			t.Fatalf("unexpected userinfo body: %s", w.Body.String())
		}
	})

	t.Run("userinfo errors", func(t *testing.T) {
		cases := []struct {
			err    error
			status int
			code   string
		}{
			{domain.ErrInvalidToken, http.StatusUnauthorized, "INVALID_SESSION_ID"},
			{domain.ErrNotFound, http.StatusNotFound, "NOT_FOUND"},
			{errors.New("boom"), http.StatusBadRequest, "INVALID_REQUEST"},
		}
		for _, tc := range cases {
			svc.EXPECT().GetUserInfo(mock.Anything, "", mock.Anything).Return(nil, tc.err).Once()
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/oauth2/userinfo", nil))
			if w.Code != tc.status {
				t.Fatalf("%v: expected %d, got %d", tc.err, tc.status, w.Code)
			}
			var body []map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if len(body) != 1 || body[0]["errorCode"] != tc.code {
				t.Fatalf("%v: unexpected body %s", tc.err, w.Body.String())
			}
		}
	})
}

func TestRequestContextMiddleware(t *testing.T) {
	if got := RequestContextFrom(context.Background()); got != (RequestContext{}) {
		t.Fatalf("unexpected empty context: %#v", got)
	}

	cases := []struct {
		name      string
		host      string
		tls       bool
		forwarded string
		username  string
		password  string
		want      RequestContext
	}{
		{
			name: "http without basic",
			host: "example.com",
			want: RequestContext{InstanceURL: "http://example.com", UserInfoBaseURL: "http://example.com"},
		},
		{
			name:     "https via TLS and basic",
			host:     "example.com",
			tls:      true,
			username: "cid",
			password: "secret",
			want: RequestContext{
				BasicUsername:   "cid",
				BasicPassword:   "secret",
				BasicAuthSet:    true,
				InstanceURL:     "https://example.com",
				UserInfoBaseURL: "http://example.com",
			},
		},
		{
			name:      "https via forwarded proto",
			host:      "example.com",
			forwarded: "https",
			want:      RequestContext{InstanceURL: "https://example.com", UserInfoBaseURL: "http://example.com"},
		},
		{
			name: "empty host",
			want: RequestContext{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got RequestContext
			handler := RequestContextMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				got = RequestContextFrom(r.Context())
			}))
			req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
			req.Host = tc.host
			if tc.tls {
				req.TLS = &tls.ConnectionState{}
			}
			if tc.forwarded != "" {
				req.Header.Set("X-Forwarded-Proto", tc.forwarded)
			}
			if tc.username != "" {
				req.SetBasicAuth(tc.username, tc.password)
			}
			handler.ServeHTTP(httptest.NewRecorder(), req)
			if got != tc.want {
				t.Fatalf("unexpected context: got %#v want %#v", got, tc.want)
			}
		})
	}
}

func TestAuthorizeLocationAndMapperHelpers(t *testing.T) {
	cases := []struct {
		name string
		res  *domain.AuthorizeResult
		want string
	}{
		{"code with query", &domain.AuthorizeResult{Code: "c1", RedirectURI: "http://localhost/cb?foo=1", State: "s"}, "http://localhost/cb?foo=1&code=c1&state=s"},
		{"code without query", &domain.AuthorizeResult{Code: "c1", RedirectURI: "http://localhost/cb"}, "http://localhost/cb?code=c1"},
		{"implicit token", &domain.AuthorizeResult{AccessToken: "at", RedirectURI: "http://localhost/cb"}, "http://localhost/cb#access_token=at&token_type=Bearer"},
		{"empty", &domain.AuthorizeResult{RedirectURI: "http://localhost/cb"}, "http://localhost/cb"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := authorizeLocation(tc.res); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}

	if got := stringValue(nil); got != "" {
		t.Fatalf("unexpected nil value %q", got)
	}
	value := "value"
	if got := stringValue(&value); got != value {
		t.Fatalf("unexpected value %q", got)
	}
	if got := stringPointer(""); got != nil {
		t.Fatalf("expected nil pointer, got %#v", got)
	}
	if got := stringPointer(value); got == nil || *got != value {
		t.Fatalf("unexpected pointer %#v", got)
	}
	if got := int64Pointer(0); got != nil {
		t.Fatalf("expected nil pointer, got %#v", got)
	}
	if got := int64Pointer(7); got == nil || *got != 7 {
		t.Fatalf("unexpected pointer %#v", got)
	}
	if got := apiVersionsToAPI(nil); got != nil {
		t.Fatalf("expected nil versions, got %#v", got)
	}

	zero := introspectResultToAPI(&domain.IntrospectResult{})
	if zero.Scope != nil || zero.Exp != nil || zero.Iat != nil {
		t.Fatalf("zero values must map to nil pointers: %#v", zero)
	}
	full := introspectResultToAPI(&domain.IntrospectResult{Scope: "s", ClientID: "c", Username: "u", TokenType: "t", Exp: 1, Iat: 2, Sub: "sub"})
	if full.Scope == nil || full.ClientId == nil || full.Username == nil || full.TokenType == nil || full.Exp == nil || full.Iat == nil || full.Sub == nil {
		t.Fatalf("expected all pointers set: %#v", full)
	}

	errorCases := []struct {
		err    error
		status int
		code   string
	}{
		{domain.ErrNotFound, http.StatusNotFound, "NOT_FOUND"},
		{domain.ErrRequiredFieldMissing, http.StatusBadRequest, "REQUIRED_FIELD_MISSING"},
		{domain.ErrMalformedQuery, http.StatusBadRequest, "MALFORMED_QUERY"},
		{domain.ErrMalformedSearch, http.StatusBadRequest, "MALFORMED_SEARCH"},
		{domain.ErrInvalidToken, http.StatusUnauthorized, "INVALID_SESSION_ID"},
		{domain.ErrInvalidClient, http.StatusUnauthorized, "INVALID_SESSION_ID"},
		{domain.ErrInvalidGrant, http.StatusUnauthorized, "INVALID_SESSION_ID"},
		{errors.New("boom"), http.StatusBadRequest, "INVALID_REQUEST"},
	}
	for _, tc := range errorCases {
		status, code := salesforceErrorStatus(tc.err)
		if status != tc.status || code != tc.code {
			t.Fatalf("%v: got %d %s, want %d %s", tc.err, status, code, tc.status, tc.code)
		}
	}
}

func assertOAuthError(t *testing.T, body []byte, want string) {
	t.Helper()
	var payload map[string]string
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["error"] != want {
		t.Fatalf("expected error %q, got %s", want, body)
	}
}

func TestAuthMappersCoverGeneratedFields(t *testing.T) {
	tokenFields := map[string]bool{
		"GrantType": true, "ClientId": true, "ClientSecret": true, "Username": true,
		"Password": true, "Code": true, "RedirectUri": true, "RefreshToken": true,
		"Assertion": true, "Scope": true,
	}
	tokenType := reflect.TypeOf(auth.TokenRequest{})
	for i := 0; i < tokenType.NumField(); i++ {
		name := tokenType.Field(i).Name
		if !tokenFields[name] {
			t.Errorf("generated token field %q is not mapped by tokenRequestFromAPI", name)
		}
	}

	authorizeFields := map[string]bool{
		"ResponseType": true, "ClientId": true, "RedirectUri": true, "Scope": true, "State": true,
	}
	authorizeType := reflect.TypeOf(headlessidentityapidemo.GetServicesOauth2AuthorizeParams{})
	for i := 0; i < authorizeType.NumField(); i++ {
		name := authorizeType.Field(i).Name
		if !authorizeFields[name] {
			t.Errorf("generated authorize field %q is not mapped by authorizeRequestFromAPI", name)
		}
	}
}
