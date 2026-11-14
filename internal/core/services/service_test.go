package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/faber-numeris/sem/internal/core/domain"
	"github.com/faber-numeris/sem/internal/mocks"
	"github.com/stretchr/testify/mock"
)

func TestService_VersionsAndResources(t *testing.T) {
	repo := mocks.NewMockRepository(t)
	svc := NewService(repo, "")
	if len(svc.ListVersions()) != 4 {
		t.Fatal("expected 4 versions")
	}
	if err := svc.ValidateVersion("v61.0"); err != nil {
		t.Fatal(err)
	}
	if err := svc.ValidateVersion("99.0"); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	if res, err := svc.ListResources("v61.0"); err != nil || res["sobjects"] == "" {
		t.Fatal(err)
	}
	if _, err := svc.ListResources("v1.0"); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
}

func TestService_Auth(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockRepository(t)
	svc := NewService(repo, "http://localhost:8080/")

	defUser := &domain.SalesforceUser{ID: "005A", OrganizationID: "00DA", Username: "admin@example.com", IsActive: true}

	// client_credentials: success, bad client, secret mismatch, default user error, create token error
	repo.EXPECT().GetClient(ctx, "default_client_id").Return(&domain.OAuthClient{ClientID: "default_client_id", ClientSecret: "sec", IsActive: true}, nil).Once()
	repo.EXPECT().GetDefaultUser(ctx).Return(defUser, nil).Once()
	repo.EXPECT().CreateToken(ctx, mock.Anything).Return(&domain.OAuthToken{AccessToken: "at1"}, nil).Once()
	if tok, err := svc.IssueToken(ctx, domain.TokenRequest{GrantType: "client_credentials"}); err != nil || tok.AccessToken != "at1" {
		t.Fatal(err)
	}
	repo.EXPECT().GetClient(ctx, "c1").Return(&domain.OAuthClient{ClientID: "c1", ClientSecret: "sec", IsActive: true}, nil).Once()
	if _, err := svc.IssueToken(ctx, domain.TokenRequest{GrantType: "client_credentials", ClientID: "c1", ClientSecret: "wrong"}); !errors.Is(err, domain.ErrInvalidClient) {
		t.Fatal("expected ErrInvalidClient")
	}
	repo.EXPECT().GetClient(ctx, "c1").Return(&domain.OAuthClient{ClientID: "c1", IsActive: true}, nil).Once()
	repo.EXPECT().GetDefaultUser(ctx).Return(nil, errors.New("no user")).Once()
	if _, err := svc.IssueToken(ctx, domain.TokenRequest{GrantType: "client_credentials", ClientID: "c1"}); err == nil {
		t.Fatal("expected err")
	}

	// password grant: with username (ok & fail) and without username (ok & fail)
	repo.EXPECT().GetUserByUsername(ctx, "admin@example.com").Return(defUser, nil).Once()
	repo.EXPECT().CreateToken(ctx, mock.Anything).Return(&domain.OAuthToken{AccessToken: "at2"}, nil).Once()
	if tok, err := svc.IssueToken(ctx, domain.TokenRequest{GrantType: "password", Username: "admin@example.com", InstanceURL: "https://org.my.salesforce.com"}); err != nil || tok.AccessToken != "at2" {
		t.Fatal(err)
	}
	repo.EXPECT().GetUserByUsername(ctx, "missing").Return(nil, domain.ErrNotFound).Once()
	if _, err := svc.IssueToken(ctx, domain.TokenRequest{GrantType: "password", Username: "missing"}); !errors.Is(err, domain.ErrInvalidGrant) {
		t.Fatal("expected ErrInvalidGrant")
	}
	repo.EXPECT().GetDefaultUser(ctx).Return(defUser, nil).Once()
	repo.EXPECT().CreateToken(ctx, mock.Anything).Return(&domain.OAuthToken{AccessToken: "at3"}, nil).Once()
	if _, err := svc.IssueToken(ctx, domain.TokenRequest{GrantType: "password"}); err != nil {
		t.Fatal(err)
	}
	repo.EXPECT().GetDefaultUser(ctx).Return(nil, errors.New("err")).Once()
	if _, err := svc.IssueToken(ctx, domain.TokenRequest{GrantType: "password"}); err == nil {
		t.Fatal("expected err")
	}

	// authorization_code grant
	if _, err := svc.IssueToken(ctx, domain.TokenRequest{GrantType: "authorization_code"}); !errors.Is(err, domain.ErrInvalidGrant) {
		t.Fatal("expected ErrInvalidGrant")
	}
	repo.EXPECT().GetAuthCode(ctx, "used").Return(&domain.OAuthAuthCode{Code: "used", IsUsed: true, ExpiresAt: time.Now().Add(time.Hour)}, nil).Once()
	if _, err := svc.IssueToken(ctx, domain.TokenRequest{GrantType: "authorization_code", Code: "used"}); !errors.Is(err, domain.ErrInvalidGrant) {
		t.Fatal("expected ErrInvalidGrant")
	}
	repo.EXPECT().GetAuthCode(ctx, "c1").Return(&domain.OAuthAuthCode{Code: "c1", ClientID: "cli", UserID: "005A", Scope: "api", ExpiresAt: time.Now().Add(time.Hour)}, nil).Once()
	repo.EXPECT().MarkAuthCodeUsed(ctx, "c1").Return(errors.New("db err")).Once()
	if _, err := svc.IssueToken(ctx, domain.TokenRequest{GrantType: "authorization_code", Code: "c1"}); err == nil {
		t.Fatal("expected err")
	}
	repo.EXPECT().GetAuthCode(ctx, "c1").Return(&domain.OAuthAuthCode{Code: "c1", ClientID: "cli", UserID: "005A", Scope: "api", ExpiresAt: time.Now().Add(time.Hour)}, nil).Once()
	repo.EXPECT().MarkAuthCodeUsed(ctx, "c1").Return(nil).Once()
	repo.EXPECT().GetUserByID(ctx, "005A").Return(nil, errors.New("err")).Once()
	if _, err := svc.IssueToken(ctx, domain.TokenRequest{GrantType: "authorization_code", Code: "c1"}); err == nil {
		t.Fatal("expected err")
	}
	repo.EXPECT().GetAuthCode(ctx, "c1").Return(&domain.OAuthAuthCode{Code: "c1", ClientID: "cli", UserID: "005A", Scope: "api", ExpiresAt: time.Now().Add(time.Hour)}, nil).Once()
	repo.EXPECT().MarkAuthCodeUsed(ctx, "c1").Return(nil).Once()
	repo.EXPECT().GetUserByID(ctx, "005A").Return(defUser, nil).Once()
	repo.EXPECT().CreateToken(ctx, mock.Anything).Return(&domain.OAuthToken{AccessToken: "at4"}, nil).Once()
	if _, err := svc.IssueToken(ctx, domain.TokenRequest{GrantType: "authorization_code", Code: "c1"}); err != nil {
		t.Fatal(err)
	}

	// refresh_token grant
	if _, err := svc.IssueToken(ctx, domain.TokenRequest{GrantType: "refresh_token"}); !errors.Is(err, domain.ErrInvalidGrant) {
		t.Fatal("expected ErrInvalidGrant")
	}
	repo.EXPECT().GetTokenByRefresh(ctx, "rev").Return(&domain.OAuthToken{IsRevoked: true}, nil).Once()
	if _, err := svc.IssueToken(ctx, domain.TokenRequest{GrantType: "refresh_token", RefreshToken: "rev"}); !errors.Is(err, domain.ErrInvalidGrant) {
		t.Fatal("expected ErrInvalidGrant")
	}
	repo.EXPECT().GetTokenByRefresh(ctx, "rt1").Return(&domain.OAuthToken{ClientID: "cli", UserID: "005A", Scope: "api"}, nil).Once()
	repo.EXPECT().GetUserByID(ctx, "005A").Return(nil, errors.New("err")).Once()
	if _, err := svc.IssueToken(ctx, domain.TokenRequest{GrantType: "refresh_token", RefreshToken: "rt1"}); err == nil {
		t.Fatal("expected err")
	}
	repo.EXPECT().GetTokenByRefresh(ctx, "rt1").Return(&domain.OAuthToken{ClientID: "cli", UserID: "005A", Scope: "api"}, nil).Once()
	repo.EXPECT().GetUserByID(ctx, "005A").Return(defUser, nil).Once()
	repo.EXPECT().CreateToken(ctx, mock.Anything).Return(&domain.OAuthToken{AccessToken: "at5"}, nil).Once()
	if _, err := svc.IssueToken(ctx, domain.TokenRequest{GrantType: "refresh_token", RefreshToken: "rt1"}); err != nil {
		t.Fatal(err)
	}

	// jwt-bearer grant
	if _, err := svc.IssueToken(ctx, domain.TokenRequest{GrantType: "urn:ietf:params:oauth:grant-type:jwt-bearer"}); !errors.Is(err, domain.ErrInvalidGrant) {
		t.Fatal("expected ErrInvalidGrant")
	}
	repo.EXPECT().GetDefaultUser(ctx).Return(nil, errors.New("err")).Once()
	if _, err := svc.IssueToken(ctx, domain.TokenRequest{GrantType: "urn:ietf:params:oauth:grant-type:jwt-bearer", Assertion: "jwt"}); err == nil {
		t.Fatal("expected err")
	}
	repo.EXPECT().GetDefaultUser(ctx).Return(defUser, nil).Once()
	repo.EXPECT().CreateToken(ctx, mock.Anything).Return(&domain.OAuthToken{AccessToken: "at6"}, nil).Once()
	if _, err := svc.IssueToken(ctx, domain.TokenRequest{GrantType: "urn:ietf:params:oauth:grant-type:jwt-bearer", Assertion: "jwt"}); err != nil {
		t.Fatal(err)
	}

	// unsupported grant
	if _, err := svc.IssueToken(ctx, domain.TokenRequest{GrantType: "unknown"}); !errors.Is(err, domain.ErrInvalidGrant) {
		t.Fatal("expected ErrInvalidGrant")
	}

	// Authorize: missing client_id, invalid client, user error, code flow (ok + err), token flow (ok + err), unsupported response_type
	if _, err := svc.Authorize(ctx, domain.AuthorizeRequest{}); !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatal("expected ErrInvalidRequest")
	}
	repo.EXPECT().GetClient(ctx, "bad").Return(nil, domain.ErrInvalidClient).Once()
	if _, err := svc.Authorize(ctx, domain.AuthorizeRequest{ClientID: "bad"}); !errors.Is(err, domain.ErrInvalidClient) {
		t.Fatal("expected ErrInvalidClient")
	}
	repo.EXPECT().GetClient(ctx, "c1").Return(&domain.OAuthClient{ClientID: "c1", IsActive: true, RedirectURI: "http://cb", Scopes: "api"}, nil).Once()
	repo.EXPECT().GetDefaultUser(ctx).Return(nil, errors.New("err")).Once()
	if _, err := svc.Authorize(ctx, domain.AuthorizeRequest{ClientID: "c1"}); err == nil {
		t.Fatal("expected err")
	}
	repo.EXPECT().GetClient(ctx, "c1").Return(&domain.OAuthClient{ClientID: "c1", IsActive: true, RedirectURI: "http://cb", Scopes: "api"}, nil).Once()
	repo.EXPECT().GetDefaultUser(ctx).Return(defUser, nil).Once()
	repo.EXPECT().CreateAuthCode(ctx, mock.Anything).Return(nil, errors.New("err")).Once()
	if _, err := svc.Authorize(ctx, domain.AuthorizeRequest{ClientID: "c1", ResponseType: "code"}); err == nil {
		t.Fatal("expected err")
	}
	repo.EXPECT().GetClient(ctx, "c1").Return(&domain.OAuthClient{ClientID: "c1", IsActive: true, RedirectURI: "http://cb", Scopes: "api"}, nil).Once()
	repo.EXPECT().GetDefaultUser(ctx).Return(defUser, nil).Once()
	repo.EXPECT().CreateAuthCode(ctx, mock.Anything).Return(&domain.OAuthAuthCode{Code: "code1"}, nil).Once()
	if res, err := svc.Authorize(ctx, domain.AuthorizeRequest{ClientID: "c1", ResponseType: "code"}); err != nil || res.Code == "" {
		t.Fatal(err)
	}
	repo.EXPECT().GetClient(ctx, "c1").Return(&domain.OAuthClient{ClientID: "c1", IsActive: true, RedirectURI: "http://cb", Scopes: "api"}, nil).Once()
	repo.EXPECT().GetDefaultUser(ctx).Return(defUser, nil).Once()
	repo.EXPECT().GetClient(ctx, "c1").Return(&domain.OAuthClient{ClientID: "c1", IsActive: true}, nil).Once()
	repo.EXPECT().GetDefaultUser(ctx).Return(defUser, nil).Once()
	repo.EXPECT().CreateToken(ctx, mock.Anything).Return(&domain.OAuthToken{AccessToken: "at_implicit"}, nil).Once()
	if res, err := svc.Authorize(ctx, domain.AuthorizeRequest{ClientID: "c1", ResponseType: "token"}); err != nil || res.AccessToken != "at_implicit" {
		t.Fatal(err)
	}
	repo.EXPECT().GetClient(ctx, "c1").Return(&domain.OAuthClient{ClientID: "c1", IsActive: true}, nil).Once()
	repo.EXPECT().GetDefaultUser(ctx).Return(defUser, nil).Once()
	repo.EXPECT().GetClient(ctx, "c1").Return(nil, domain.ErrInvalidClient).Once()
	if _, err := svc.Authorize(ctx, domain.AuthorizeRequest{ClientID: "c1", ResponseType: "token"}); err == nil {
		t.Fatal("expected err")
	}
	repo.EXPECT().GetClient(ctx, "c1").Return(&domain.OAuthClient{ClientID: "c1", IsActive: true}, nil).Once()
	repo.EXPECT().GetDefaultUser(ctx).Return(defUser, nil).Once()
	if _, err := svc.Authorize(ctx, domain.AuthorizeRequest{ClientID: "c1", ResponseType: "invalid"}); !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatal("expected ErrInvalidRequest")
	}

	// RevokeToken
	if err := svc.RevokeToken(ctx, "  "); !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatal("expected ErrInvalidRequest")
	}
	repo.EXPECT().GetTokenByAny(ctx, "bad").Return(nil, domain.ErrInvalidToken).Once()
	if err := svc.RevokeToken(ctx, "bad"); !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatal("expected ErrInvalidToken")
	}
	repo.EXPECT().GetTokenByAny(ctx, "tok").Return(&domain.OAuthToken{AccessToken: "tok"}, nil).Once()
	repo.EXPECT().RevokeToken(ctx, "tok").Return(nil).Once()
	if err := svc.RevokeToken(ctx, "tok"); err != nil {
		t.Fatal(err)
	}

	// IntrospectToken
	if res, _ := svc.IntrospectToken(ctx, ""); res.Active {
		t.Fatal("expected inactive")
	}
	repo.EXPECT().GetTokenByAny(ctx, "rev").Return(&domain.OAuthToken{IsRevoked: true}, nil).Once()
	if res, _ := svc.IntrospectToken(ctx, "rev"); res.Active {
		t.Fatal("expected inactive")
	}
	repo.EXPECT().GetTokenByAny(ctx, "good").Return(&domain.OAuthToken{
		AccessToken: "good",
		UserID:      "005A",
		ExpiresAt:   time.Now().Add(time.Hour),
	}, nil).Once()
	repo.EXPECT().GetUserByID(ctx, "005A").Return(defUser, nil).Once()
	if res, _ := svc.IntrospectToken(ctx, "good"); !res.Active || res.Username != "admin@example.com" {
		t.Fatal("expected active")
	}

	// GetUserInfo
	repo.EXPECT().GetTokenByAccess(ctx, "rev").Return(&domain.OAuthToken{IsRevoked: true}, nil).Once()
	if _, err := svc.GetUserInfo(ctx, "Bearer rev", ""); !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatal("expected ErrInvalidToken")
	}
	repo.EXPECT().GetTokenByAccess(ctx, "good").Return(&domain.OAuthToken{UserID: "005A", ExpiresAt: time.Now().Add(time.Hour)}, nil).Once()
	repo.EXPECT().GetUserByID(ctx, "005A").Return(defUser, nil).Once()
	if info, err := svc.GetUserInfo(ctx, "Bearer good", "https://custom.my.salesforce.com"); err != nil || info.UserID != "005A" {
		t.Fatal(err)
	}
	repo.EXPECT().GetDefaultUser(ctx).Return(defUser, nil).Once()
	if info, err := svc.GetUserInfo(ctx, "", ""); err != nil || info.UserID != "005A" {
		t.Fatal(err)
	}
	repo.EXPECT().GetDefaultUser(ctx).Return(nil, errors.New("err")).Once()
	if _, err := svc.GetUserInfo(ctx, "", ""); err == nil {
		t.Fatal("expected err")
	}
}

func TestService_SObject(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockRepository(t)
	svc := NewService(repo, "http://localhost:8080")

	if _, err := svc.GetLimits(ctx, "v1.0"); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	repo.EXPECT().ListOrgLimits(ctx).Return(nil, errors.New("err")).Once()
	if _, err := svc.GetLimits(ctx, "v61.0"); err == nil {
		t.Fatal("expected err")
	}
	repo.EXPECT().ListOrgLimits(ctx).Return([]domain.OrgLimit{{Name: "DailyApiRequests", Max: 100, Remaining: 99}}, nil).Once()
	if lims, err := svc.GetLimits(ctx, "v61.0"); err != nil || lims["DailyApiRequests"].Max != 100 {
		t.Fatal(err)
	}

	if _, err := svc.DescribeGlobal(ctx, "v1.0", false); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	repo.EXPECT().ListSObjectDefinitions(ctx, false).Return(nil, errors.New("err")).Once()
	if _, err := svc.DescribeGlobal(ctx, "v61.0", false); err == nil {
		t.Fatal("expected err")
	}
	repo.EXPECT().ListSObjectDefinitions(ctx, true).Return([]domain.SObjectDefinition{{Name: "ApexClass", IsTooling: true}}, nil).Once()
	if dg, err := svc.DescribeGlobal(ctx, "v61.0", true); err != nil || dg["maxBatchSize"] != 200 {
		t.Fatal(err)
	}

	if _, err := svc.GetSObjectBasicInfo(ctx, "v1.0", "Account"); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	repo.EXPECT().GetSObjectDescribe(ctx, "Missing").Return(nil, domain.ErrNotFound).Once()
	if _, err := svc.GetSObjectBasicInfo(ctx, "v61.0", "Missing"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}
	repo.EXPECT().GetSObjectDescribe(ctx, "Account").Return(&domain.SObjectDescribe{Definition: domain.SObjectDefinition{Name: "Account"}}, nil).Once()
	if info, err := svc.GetSObjectBasicInfo(ctx, "v61.0", "Account"); err != nil || info["objectDescribe"] == nil {
		t.Fatal(err)
	}

	if _, err := svc.DescribeSObject(ctx, "v1.0", "Account", false); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	repo.EXPECT().GetSObjectDescribe(ctx, "Missing").Return(nil, domain.ErrNotFound).Once()
	if _, err := svc.DescribeSObject(ctx, "v61.0", "Missing", false); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}
	repo.EXPECT().GetSObjectDescribe(ctx, "Contact").Return(&domain.SObjectDescribe{
		Definition: domain.SObjectDefinition{Name: "Contact"},
		Fields: []domain.SObjectFieldDefinition{
			{Name: "Id"},
			{Name: "AccountId", ReferenceTo: "Account"},
		},
	}, nil).Once()
	if desc, err := svc.DescribeSObject(ctx, "v61.0", "Contact", false); err != nil || desc["fields"] == nil {
		t.Fatal(err)
	}

	if _, _, err := svc.CreateRecord(ctx, "v1.0", "Account", nil); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	repo.EXPECT().CreateRecord(ctx, "Account", mock.Anything).Return(&domain.SaveResult{ID: "001A", Success: true}, map[string]any{"Id": "001A"}, nil).Once()
	if res, _, err := svc.CreateRecord(ctx, "v61.0", "Account", map[string]any{"Name": "Acme"}); err != nil || res.ID != "001A" {
		t.Fatal(err)
	}

	if _, err := svc.GetRecord(ctx, "v1.0", "Account", "1", nil); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	repo.EXPECT().GetRecord(ctx, "61.0", "Account", "missing").Return(nil, domain.ErrNotFound).Once()
	if _, err := svc.GetRecord(ctx, "v61.0", "Account", "missing", nil); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}
	repo.EXPECT().GetRecord(ctx, "61.0", "Account", "1").Return(map[string]any{"Id": "1", "Name": "Acme", "Phone": "123"}, nil).Twice()
	if rec, err := svc.GetRecord(ctx, "v61.0", "Account", "1", nil); err != nil || rec["Phone"] != "123" {
		t.Fatal(err)
	}
	if rec, err := svc.GetRecord(ctx, "v61.0", "Account", "1", []string{"", "Name"}); err != nil || rec["Phone"] != nil || rec["Name"] != "Acme" {
		t.Fatal("expected projected fields only")
	}

	if _, err := svc.UpdateRecord(ctx, "v1.0", "Account", "1", nil); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	repo.EXPECT().UpdateRecord(ctx, "Account", "1", mock.Anything).Return(map[string]any{"Id": "1"}, nil).Once()
	if _, err := svc.UpdateRecord(ctx, "v61.0", "Account", "1", map[string]any{"Name": "New"}); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.UpsertRecord(ctx, "v1.0", "Account", "External_ID__c", "E1", nil); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	repo.EXPECT().UpsertRecordByExternalID(ctx, "Account", "External_ID__c", "E1", mock.Anything).Return(&domain.SaveResult{ID: "1", Success: true}, nil).Once()
	if _, err := svc.UpsertRecord(ctx, "v61.0", "Account", "External_ID__c", "E1", map[string]any{"Name": "New"}); err != nil {
		t.Fatal(err)
	}

	if err := svc.DeleteRecord(ctx, "v1.0", "Account", "1"); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	repo.EXPECT().DeleteRecord(ctx, "Account", "1").Return(nil).Once()
	if err := svc.DeleteRecord(ctx, "v61.0", "Account", "1"); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.GetContentVersionData(ctx, "v1.0", "1"); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	repo.EXPECT().GetContentVersionData(ctx, "1").Return([]byte("abc"), nil).Once()
	if b, err := svc.GetContentVersionData(ctx, "v61.0", "1"); err != nil || string(b) != "abc" {
		t.Fatal(err)
	}
}

func TestService_QueryAndSOSL(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockRepository(t)
	svc := NewService(repo, "http://localhost:8080")

	// Helper coverage for pointer types in stringifyFieldVal & operators
	sVal := "Acme"
	fVal := 100.5
	iVal := 10
	bVal := true
	var nilStr *string
	var nilFloat *float64
	var nilInt *int
	var nilBool *bool
	rec := map[string]any{
		"Name":              &sVal,
		"AnnualRevenue":     &fVal,
		"NumberOfEmployees": &iVal,
		"IsDeleted":         &bVal,
		"NilStr":            nilStr,
		"NilFloat":          nilFloat,
		"NilInt":            nilInt,
		"NilBool":           nilBool,
		"Industry":          "Tech",
	}
	for _, expr := range []string{
		"(Name = 'Acme') AND AnnualRevenue > 50 AND AnnualRevenue >= 100.5 AND NumberOfEmployees < 20 AND NumberOfEmployees <= 10",
		"Name != 'Other' AND NilStr = null AND IsDeleted = true",
		"Industry IN ('Tech', 'Finance') AND Name LIKE 'Ac%'",
		"Industry IN ('Retail') OR Industry = 'Tech'",
		"Industry < 'Z' AND Industry > 'A' AND Industry <= 'Z' AND Industry >= 'A'",
		"BareTokenWithoutOperator",
	} {
		if !matchesWhere(rec, expr) {
			t.Fatalf("expected match for %q", expr)
		}
	}
	if matchesWhere(rec, "Industry IN ('Retail') AND Name = 'Nope'") {
		t.Fatal("expected non-match")
	}
	if stringifyFieldVal(nil) != "" || stringifyFieldVal(nilStr) != "" || stringifyFieldVal(nilFloat) != "" || stringifyFieldVal(nilInt) != "" || stringifyFieldVal(nilBool) != "" {
		t.Fatal("expected empty string for nil pointers")
	}
	if _, err := parseSOQL("SELECT , FROM Account"); !errors.Is(err, domain.ErrMalformedQuery) {
		t.Fatal("expected ErrMalformedQuery for empty field list")
	}

	// ExecuteSOQL validation & errors
	if _, err := svc.ExecuteSOQL(ctx, "v1.0", "SELECT Id FROM Account", false, 0); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	if _, err := svc.ExecuteSOQL(ctx, "v61.0", "INVALID QUERY", false, 0); !errors.Is(err, domain.ErrMalformedQuery) {
		t.Fatal("expected ErrMalformedQuery")
	}
	repo.EXPECT().ListRecordsForSObject(ctx, "61.0", "Account", false).Return(nil, errors.New("db err")).Once()
	if _, err := svc.ExecuteSOQL(ctx, "v61.0", "SELECT Id FROM Account", false, 0); err == nil {
		t.Fatal("expected err")
	}

	// ExecuteSOQL COUNT(), ORDER BY ASC/DESC, LIMIT, OFFSET, FIELDS(ALL), and pagination
	sampleRecs := []map[string]any{
		{"attributes": domain.RecordAttributes{Type: "Account"}, "Id": "001A", "Name": "Beta"},
		{"attributes": domain.RecordAttributes{Type: "Account"}, "Id": "001B", "Name": "Alpha"},
		{"attributes": domain.RecordAttributes{Type: "Account"}, "Id": "001C", "Name": "Gamma"},
	}
	repo.EXPECT().ListRecordsForSObject(ctx, "61.0", "Account", false).Return(sampleRecs, nil).Times(6)

	if res, err := svc.ExecuteSOQL(ctx, "v61.0", "SELECT COUNT() FROM Account", false, 0); err != nil || res.TotalSize != 3 {
		t.Fatal(err)
	}
	if res, err := svc.ExecuteSOQL(ctx, "v61.0", "SELECT FIELDS(ALL) FROM Account WHERE MissingField = null ORDER BY Name ASC LIMIT 1 OFFSET 1", false, 0); err != nil || len(res.Records) != 1 {
		t.Fatal(err)
	}
	if res, err := svc.ExecuteSOQL(ctx, "v61.0", "SELECT Id, Name FROM Account ORDER BY Name DESC OFFSET 10", false, 0); err != nil || len(res.Records) != 0 {
		t.Fatal(err)
	}
	// Pagination creation error & success
	repo.EXPECT().CreateQueryCursor(ctx, mock.Anything).Return(nil, errors.New("cursor err")).Once()
	if _, err := svc.ExecuteSOQL(ctx, "v61.0", "SELECT Id, Name FROM Account", false, 1); err == nil {
		t.Fatal("expected cursor err")
	}
	repo.EXPECT().CreateQueryCursor(ctx, mock.Anything).Return(&domain.QueryCursor{ID: "01g123"}, nil).Once()
	resPage, err := svc.ExecuteSOQL(ctx, "v61.0", "SELECT Id, Name FROM Account", false, 1)
	if err != nil || resPage.Done || resPage.NextRecordsURL == "" {
		t.Fatal("expected paginated result")
	}
	// Default batchSize
	if res, err := svc.ExecuteSOQL(ctx, "v61.0", "SELECT Id, Name FROM Account", false, 0); err != nil || !res.Done {
		t.Fatal(err)
	}

	// QueryNextBatch: invalid version, missing cursor, bad cursor query, list error, middle page (ok + update err), final page
	if _, err := svc.QueryNextBatch(ctx, "v1.0", "01g123-1"); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	repo.EXPECT().GetQueryCursor(ctx, "01g123").Return(nil, domain.ErrNotFound).Once()
	if _, err := svc.QueryNextBatch(ctx, "v61.0", "01g123-1"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}
	repo.EXPECT().GetQueryCursor(ctx, "01g123").Return(&domain.QueryCursor{ID: "01g123", SOQLQuery: "BAD"}, nil).Once()
	if _, err := svc.QueryNextBatch(ctx, "v61.0", "01g123-1"); err == nil {
		t.Fatal("expected malformed query err")
	}
	repo.EXPECT().GetQueryCursor(ctx, "01g123").Return(&domain.QueryCursor{ID: "01g123", SObjectName: "Account", SOQLQuery: "SELECT Id FROM Account"}, nil).Once()
	repo.EXPECT().ListRecordsForSObject(ctx, "61.0", "Account", false).Return(nil, errors.New("err")).Once()
	if _, err := svc.QueryNextBatch(ctx, "v61.0", "01g123-1"); err == nil {
		t.Fatal("expected err")
	}

	// Middle page: OffsetValue=1, BatchSize=1, Total=3 -> next offset 2
	repo.EXPECT().GetQueryCursor(ctx, "01g123").Return(&domain.QueryCursor{ID: "01g123", SObjectName: "Account", SOQLQuery: "SELECT Id FROM Account", OffsetValue: 1, BatchSize: 1}, nil).Twice()
	repo.EXPECT().ListRecordsForSObject(ctx, "61.0", "Account", false).Return(sampleRecs, nil).Twice()
	repo.EXPECT().UpdateQueryCursorOffset(ctx, "01g123", 2).Return(nil, errors.New("upd err")).Once()
	if _, err := svc.QueryNextBatch(ctx, "v61.0", "01g123-1"); err == nil {
		t.Fatal("expected upd err")
	}
	repo.EXPECT().UpdateQueryCursorOffset(ctx, "01g123", 2).Return(&domain.QueryCursor{ID: "01g123", OffsetValue: 2}, nil).Once()
	if next, err := svc.QueryNextBatch(ctx, "v61.0", "01g123-1"); err != nil || next.Done {
		t.Fatal("expected Done=false")
	}

	// Final page (including start > totalSize edge case)
	repo.EXPECT().GetQueryCursor(ctx, "01g123").Return(&domain.QueryCursor{ID: "01g123", SObjectName: "Account", SOQLQuery: "SELECT Id FROM Account", OffsetValue: 10, BatchSize: 2}, nil).Once()
	repo.EXPECT().ListRecordsForSObject(ctx, "61.0", "Account", false).Return(sampleRecs, nil).Once()
	repo.EXPECT().DeleteQueryCursor(ctx, "01g123").Return(nil).Once()
	if final, err := svc.QueryNextBatch(ctx, "v61.0", "01g123"); err != nil || !final.Done {
		t.Fatal("expected Done=true")
	}

	// ExecuteSOSL
	if _, err := svc.ExecuteSOSL(ctx, "v1.0", "FIND {Acme}"); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	if _, err := svc.ExecuteSOSL(ctx, "v61.0", "BAD SOSL"); !errors.Is(err, domain.ErrMalformedSearch) {
		t.Fatal("expected ErrMalformedSearch")
	}
	// With RETURNING (projected and unprojected) + error
	repo.EXPECT().SearchRecordsForSObject(ctx, "61.0", "Account", "Acme").Return(sampleRecs[:1], nil).Once()
	repo.EXPECT().SearchRecordsForSObject(ctx, "61.0", "Contact", "Acme").Return(sampleRecs[:1], nil).Once()
	if sr, err := svc.ExecuteSOSL(ctx, "v61.0", "FIND {Acme} IN NAME FIELDS RETURNING Account(Id, Name), Contact"); err != nil || len(sr.SearchRecords) != 2 {
		t.Fatal(err)
	}
	repo.EXPECT().SearchRecordsForSObject(ctx, "61.0", "Account", "Acme").Return(nil, errors.New("err")).Once()
	if _, err := svc.ExecuteSOSL(ctx, "v61.0", "FIND {Acme} RETURNING Account"); err == nil {
		t.Fatal("expected err")
	}
	// Default targets when RETURNING omitted
	repo.EXPECT().SearchRecordsForSObject(ctx, "61.0", "Account", "Acme").Return(nil, nil).Once()
	repo.EXPECT().SearchRecordsForSObject(ctx, "61.0", "Contact", "Acme").Return(nil, nil).Once()
	repo.EXPECT().SearchRecordsForSObject(ctx, "61.0", "Lead", "Acme").Return(nil, nil).Once()
	if sr, err := svc.ExecuteSOSL(ctx, "v61.0", "FIND {Acme}"); err != nil || len(sr.SearchRecords) != 0 {
		t.Fatal(err)
	}
}

func TestService_BulkAndTooling(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockRepository(t)
	svc := NewService(repo, "http://localhost:8080")

	// CreateIngestJob
	if _, err := svc.CreateIngestJob(ctx, "v1.0", "insert", "Account", "", ""); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	if _, err := svc.CreateIngestJob(ctx, "v61.0", "invalid_op", "Account", "", ""); !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatal("expected ErrInvalidRequest")
	}
	if _, err := svc.CreateIngestJob(ctx, "v61.0", "insert", "", "", ""); !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatal("expected ErrInvalidRequest")
	}
	repo.EXPECT().CreateBulkIngestJob(ctx, mock.Anything).Return(&domain.BulkIngestJob{ID: "750A", State: "Open"}, nil).Once()
	if j, err := svc.CreateIngestJob(ctx, "v61.0", "insert", "Account", "", ""); err != nil || j.ID != "750A" {
		t.Fatal(err)
	}

	// UploadIngestBatch
	if err := svc.UploadIngestBatch(ctx, "v1.0", "750A", "Name\nAcme"); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	repo.EXPECT().GetBulkIngestJob(ctx, "missing").Return(nil, domain.ErrNotFound).Once()
	if err := svc.UploadIngestBatch(ctx, "v61.0", "missing", "Name\nAcme"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}
	repo.EXPECT().GetBulkIngestJob(ctx, "closed").Return(&domain.BulkIngestJob{ID: "closed", State: "JobComplete"}, nil).Once()
	if err := svc.UploadIngestBatch(ctx, "v61.0", "closed", "Name\nAcme"); !errors.Is(err, domain.ErrInvalidJobState) {
		t.Fatal("expected ErrInvalidJobState")
	}
	repo.EXPECT().GetBulkIngestJob(ctx, "750A").Return(&domain.BulkIngestJob{ID: "750A", State: "Open"}, nil).Once()
	repo.EXPECT().CreateBulkIngestBatch(ctx, "750A", "Name\nAcme").Return(&domain.BulkIngestBatch{ID: "b1"}, nil).Once()
	if err := svc.UploadIngestBatch(ctx, "v61.0", "750A", "Name\nAcme"); err != nil {
		t.Fatal(err)
	}

	// CloseOrAbortIngestJob: invalid version, missing job, Aborted, Invalid state, UploadComplete (insert, update, upsert, delete + batch error)
	if _, err := svc.CloseOrAbortIngestJob(ctx, "v1.0", "750A", "Aborted"); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	repo.EXPECT().GetBulkIngestJob(ctx, "missing").Return(nil, domain.ErrNotFound).Once()
	if _, err := svc.CloseOrAbortIngestJob(ctx, "v61.0", "missing", "Aborted"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}
	repo.EXPECT().GetBulkIngestJob(ctx, "750A").Return(&domain.BulkIngestJob{ID: "750A"}, nil).Once()
	repo.EXPECT().UpdateBulkIngestJobState(ctx, "750A", "Aborted").Return(&domain.BulkIngestJob{ID: "750A", State: "Aborted"}, nil).Once()
	if j, err := svc.CloseOrAbortIngestJob(ctx, "v61.0", "750A", "Aborted"); err != nil || j.State != "Aborted" {
		t.Fatal(err)
	}
	repo.EXPECT().GetBulkIngestJob(ctx, "750A").Return(&domain.BulkIngestJob{ID: "750A"}, nil).Once()
	if _, err := svc.CloseOrAbortIngestJob(ctx, "v61.0", "750A", "InvalidState"); !errors.Is(err, domain.ErrInvalidJobState) {
		t.Fatal("expected ErrInvalidJobState")
	}
	repo.EXPECT().GetBulkIngestJob(ctx, "750A").Return(&domain.BulkIngestJob{ID: "750A", Operation: "insert"}, nil).Once()
	repo.EXPECT().ListBulkIngestBatches(ctx, "750A").Return(nil, errors.New("err")).Once()
	if _, err := svc.CloseOrAbortIngestJob(ctx, "v61.0", "750A", "UploadComplete"); err == nil {
		t.Fatal("expected err")
	}

	// UploadComplete for insert (1 ok, 1 fail, 1 header-only batch)
	repo.EXPECT().GetBulkIngestJob(ctx, "750A").Return(&domain.BulkIngestJob{ID: "750A", Operation: "insert", ObjectName: "Account"}, nil).Once()
	repo.EXPECT().ListBulkIngestBatches(ctx, "750A").Return([]domain.BulkIngestBatch{
		{CSVContent: "OnlyHeader"},
		{CSVContent: "Name\nAcme\n\"\""},
	}, nil).Once()
	repo.EXPECT().CreateRecord(ctx, "Account", map[string]any{"Name": "Acme"}).Return(&domain.SaveResult{ID: "001A", Created: true}, nil, nil).Once()
	repo.EXPECT().CreateRecord(ctx, "Account", map[string]any{"Name": ""}).Return(nil, nil, domain.ErrRequiredFieldMissing).Once()
	repo.EXPECT().CreateBulkIngestResult(ctx, mock.Anything).Return(&domain.BulkIngestResult{}, nil).Twice()
	repo.EXPECT().UpdateBulkIngestJobStats(ctx, "750A", "JobComplete", 2, 1).Return(&domain.BulkIngestJob{ID: "750A", State: "JobComplete"}, nil).Once()
	if _, err := svc.CloseOrAbortIngestJob(ctx, "v61.0", "750A", "UploadComplete"); err != nil {
		t.Fatal(err)
	}

	// UploadComplete for update (missing Id, update error, update success)
	repo.EXPECT().GetBulkIngestJob(ctx, "750U").Return(&domain.BulkIngestJob{ID: "750U", Operation: "update", ObjectName: "Account"}, nil).Once()
	repo.EXPECT().ListBulkIngestBatches(ctx, "750U").Return([]domain.BulkIngestBatch{
		{CSVContent: "Id,Name\n,NoId\nbad,Fail\n001A,Ok"},
	}, nil).Once()
	repo.EXPECT().UpdateRecord(ctx, "Account", "bad", mock.Anything).Return(nil, domain.ErrNotFound).Once()
	repo.EXPECT().UpdateRecord(ctx, "Account", "001A", mock.Anything).Return(map[string]any{"Id": "001A"}, nil).Once()
	repo.EXPECT().CreateBulkIngestResult(ctx, mock.Anything).Return(&domain.BulkIngestResult{}, nil).Times(3)
	repo.EXPECT().UpdateBulkIngestJobStats(ctx, "750U", "JobComplete", 3, 2).Return(&domain.BulkIngestJob{ID: "750U", State: "JobComplete"}, nil).Once()
	if _, err := svc.CloseOrAbortIngestJob(ctx, "v61.0", "750U", "UploadComplete"); err != nil {
		t.Fatal(err)
	}

	// UploadComplete for upsert (ok & fail)
	repo.EXPECT().GetBulkIngestJob(ctx, "750P").Return(&domain.BulkIngestJob{ID: "750P", Operation: "upsert", ObjectName: "Account"}, nil).Once()
	repo.EXPECT().ListBulkIngestBatches(ctx, "750P").Return([]domain.BulkIngestBatch{
		{CSVContent: "External_ID__c,Name\nE1,Ok\nE2,Fail"},
	}, nil).Once()
	repo.EXPECT().UpsertRecordByExternalID(ctx, "Account", "External_ID__c", "E1", mock.Anything).Return(&domain.SaveResult{ID: "001A", Created: true}, nil).Once()
	repo.EXPECT().UpsertRecordByExternalID(ctx, "Account", "External_ID__c", "E2", mock.Anything).Return(nil, errors.New("err")).Once()
	repo.EXPECT().CreateBulkIngestResult(ctx, mock.Anything).Return(&domain.BulkIngestResult{}, nil).Twice()
	repo.EXPECT().UpdateBulkIngestJobStats(ctx, "750P", "JobComplete", 2, 1).Return(&domain.BulkIngestJob{ID: "750P", State: "JobComplete"}, nil).Once()
	if _, err := svc.CloseOrAbortIngestJob(ctx, "v61.0", "750P", "UploadComplete"); err != nil {
		t.Fatal(err)
	}

	// UploadComplete for delete (missing Id, delete error, delete success)
	repo.EXPECT().GetBulkIngestJob(ctx, "750D").Return(&domain.BulkIngestJob{ID: "750D", Operation: "delete", ObjectName: "Account"}, nil).Once()
	repo.EXPECT().ListBulkIngestBatches(ctx, "750D").Return([]domain.BulkIngestBatch{
		{CSVContent: "Id\n\"\"\nbad\n001A"},
	}, nil).Once()
	repo.EXPECT().DeleteRecord(ctx, "Account", "bad").Return(domain.ErrNotFound).Once()
	repo.EXPECT().DeleteRecord(ctx, "Account", "001A").Return(nil).Once()
	repo.EXPECT().CreateBulkIngestResult(ctx, mock.Anything).Return(&domain.BulkIngestResult{}, nil).Times(3)
	repo.EXPECT().UpdateBulkIngestJobStats(ctx, "750D", "JobComplete", 3, 2).Return(&domain.BulkIngestJob{ID: "750D", State: "JobComplete"}, nil).Once()
	if _, err := svc.CloseOrAbortIngestJob(ctx, "v61.0", "750D", "UploadComplete"); err != nil {
		t.Fatal(err)
	}

	// GetIngestJob & ListIngestJobs & GetIngestResultsCSV
	if _, err := svc.GetIngestJob(ctx, "v1.0", "750A"); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	repo.EXPECT().GetBulkIngestJob(ctx, "750A").Return(&domain.BulkIngestJob{ID: "750A"}, nil).Once()
	if _, err := svc.GetIngestJob(ctx, "v61.0", "750A"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ListIngestJobs(ctx, "v1.0"); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	repo.EXPECT().ListBulkIngestJobs(ctx).Return([]domain.BulkIngestJob{{ID: "750A"}}, nil).Once()
	if _, err := svc.ListIngestJobs(ctx, "v61.0"); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.GetIngestResultsCSV(ctx, "v1.0", "750A", true); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	repo.EXPECT().GetBulkIngestJob(ctx, "missing").Return(nil, domain.ErrNotFound).Once()
	if _, err := svc.GetIngestResultsCSV(ctx, "v61.0", "missing", true); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}
	repo.EXPECT().GetBulkIngestJob(ctx, "750A").Return(&domain.BulkIngestJob{ID: "750A"}, nil).Once()
	repo.EXPECT().ListBulkIngestBatches(ctx, "750A").Return(nil, errors.New("err")).Once()
	if _, err := svc.GetIngestResultsCSV(ctx, "v61.0", "750A", true); err == nil {
		t.Fatal("expected err")
	}
	repo.EXPECT().GetBulkIngestJob(ctx, "750A").Return(&domain.BulkIngestJob{ID: "750A"}, nil).Once()
	repo.EXPECT().ListBulkIngestBatches(ctx, "750A").Return([]domain.BulkIngestBatch{{CSVContent: "Name\nAcme"}}, nil).Once()
	repo.EXPECT().ListBulkIngestResults(ctx, "750A", true).Return(nil, errors.New("err")).Once()
	if _, err := svc.GetIngestResultsCSV(ctx, "v61.0", "750A", true); err == nil {
		t.Fatal("expected err")
	}
	repo.EXPECT().GetBulkIngestJob(ctx, "750A").Return(&domain.BulkIngestJob{ID: "750A"}, nil).Twice()
	repo.EXPECT().ListBulkIngestBatches(ctx, "750A").Return([]domain.BulkIngestBatch{{CSVContent: "Name\nAcme"}}, nil).Twice()
	repo.EXPECT().ListBulkIngestResults(ctx, "750A", true).Return([]domain.BulkIngestResult{{SFID: "001A", SFCreated: true, RowCSV: "Acme"}}, nil).Once()
	if out, err := svc.GetIngestResultsCSV(ctx, "v61.0", "750A", true); err != nil || !strings.Contains(out, "sf__Created") {
		t.Fatal(err)
	}
	repo.EXPECT().ListBulkIngestResults(ctx, "750A", false).Return([]domain.BulkIngestResult{{SFID: "", SFError: "missing field", RowCSV: "Bad"}}, nil).Once()
	if out, err := svc.GetIngestResultsCSV(ctx, "v61.0", "750A", false); err != nil || !strings.Contains(out, "sf__Error") {
		t.Fatal(err)
	}

	// Bulk Query Jobs
	if _, err := svc.CreateQueryJob(ctx, "v1.0", "query", "SELECT Id FROM Account"); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	if _, err := svc.CreateQueryJob(ctx, "v61.0", "invalid_op", "SELECT Id FROM Account"); !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatal("expected ErrInvalidRequest")
	}
	if _, err := svc.CreateQueryJob(ctx, "v61.0", "", "BAD QUERY"); !errors.Is(err, domain.ErrMalformedQuery) {
		t.Fatal("expected ErrMalformedQuery")
	}
	repo.EXPECT().CreateBulkQueryJob(ctx, mock.Anything).Return(&domain.BulkQueryJob{ID: "750Q"}, nil).Once()
	if j, err := svc.CreateQueryJob(ctx, "v61.0", "", "SELECT Id, Name FROM Account"); err != nil || j.ID != "750Q" {
		t.Fatal(err)
	}

	if _, err := svc.GetQueryJob(ctx, "v1.0", "750Q"); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	repo.EXPECT().GetBulkQueryJob(ctx, "750Q").Return(&domain.BulkQueryJob{ID: "750Q"}, nil).Once()
	if _, err := svc.GetQueryJob(ctx, "v61.0", "750Q"); err != nil {
		t.Fatal(err)
	}

	if _, _, err := svc.GetQueryJobResultsCSV(ctx, "v1.0", "750Q", 0); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	repo.EXPECT().GetBulkQueryJob(ctx, "missing").Return(nil, domain.ErrNotFound).Once()
	if _, _, err := svc.GetQueryJobResultsCSV(ctx, "v61.0", "missing", 0); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}
	repo.EXPECT().GetBulkQueryJob(ctx, "badq").Return(&domain.BulkQueryJob{ID: "badq", QueryText: "BAD"}, nil).Once()
	if _, _, err := svc.GetQueryJobResultsCSV(ctx, "v61.0", "badq", 0); err == nil {
		t.Fatal("expected err")
	}
	repo.EXPECT().GetBulkQueryJob(ctx, "750Q").Return(&domain.BulkQueryJob{ID: "750Q", Operation: "query", QueryText: "SELECT * FROM Account"}, nil).Twice()
	repo.EXPECT().ListRecordsForSObject(ctx, "61.0", "Account", false).Return(nil, errors.New("err")).Once()
	if _, _, err := svc.GetQueryJobResultsCSV(ctx, "v61.0", "750Q", 0); err == nil {
		t.Fatal("expected err")
	}
	repo.EXPECT().ListRecordsForSObject(ctx, "61.0", "Account", false).Return([]map[string]any{
		{"Id": "001A", "Name": "Acme"},
		{"Id": "001B", "Name": "Beta"},
	}, nil).Once()
	if csvOut, count, err := svc.GetQueryJobResultsCSV(ctx, "v61.0", "750Q", 1); err != nil || count != 1 || !strings.Contains(csvOut, "Acme") {
		t.Fatal(err)
	}

	// Tooling Service
	if _, err := svc.ExecuteToolingQuery(ctx, "v1.0", "SELECT Id FROM ApexClass"); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	if _, err := svc.ExecuteToolingQuery(ctx, "v61.0", "BAD"); !errors.Is(err, domain.ErrMalformedQuery) {
		t.Fatal("expected ErrMalformedQuery")
	}
	repo.EXPECT().ListToolingRecords(ctx, "61.0", "ApexClass").Return(nil, errors.New("err")).Once()
	if _, err := svc.ExecuteToolingQuery(ctx, "v61.0", "SELECT Id FROM ApexClass"); err == nil {
		t.Fatal("expected err")
	}
	repo.EXPECT().ListToolingRecords(ctx, "61.0", "ApexClass").Return([]map[string]any{{"Id": "01pA", "Name": "MyClass"}}, nil).Twice()
	if res, err := svc.ExecuteToolingQuery(ctx, "v61.0", "SELECT COUNT() FROM ApexClass"); err != nil || res.TotalSize != 1 {
		t.Fatal(err)
	}
	if res, err := svc.ExecuteToolingQuery(ctx, "v61.0", "SELECT Id, Name FROM ApexClass"); err != nil || len(res.Records) != 1 {
		t.Fatal(err)
	}

	if _, err := svc.ExecuteAnonymousApex(ctx, "v1.0", "System.debug('hi');"); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	if res, err := svc.ExecuteAnonymousApex(ctx, "v61.0", ""); err != nil || res.Compiled {
		t.Fatal("expected compile failure")
	}
	if res, err := svc.ExecuteAnonymousApex(ctx, "v61.0", "throw new CalloutException();"); err != nil || !res.Compiled || res.Success {
		t.Fatal("expected runtime exception")
	}
	if res, err := svc.ExecuteAnonymousApex(ctx, "v61.0", "System.debug('ok');"); err != nil || !res.Success {
		t.Fatal("expected success")
	}

	repo.EXPECT().GetSObjectDescribe(ctx, "ApexClass").Return(&domain.SObjectDescribe{Definition: domain.SObjectDefinition{Name: "ApexClass"}}, nil).Once()
	if _, err := svc.DescribeToolingSObject(ctx, "v61.0", "ApexClass"); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.RunTestsAsynchronous(ctx, "v1.0", "", ""); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	repo.EXPECT().CreateApexTestRun(ctx, "AccountServiceTest", "RunSpecifiedTests").Return(nil, errors.New("err")).Once()
	if _, err := svc.RunTestsAsynchronous(ctx, "v61.0", "", ""); err == nil {
		t.Fatal("expected err")
	}
	repo.EXPECT().CreateApexTestRun(ctx, "AccountServiceTest", "RunSpecifiedTests").Return(&domain.ApexTestRun{ID: "707A"}, nil).Once()
	if id, err := svc.RunTestsAsynchronous(ctx, "v61.0", "", ""); err != nil || id != "707A" {
		t.Fatal(err)
	}

	// ExecuteApexRest
	repo.EXPECT().GetApexRestRoute(ctx, "/v1/integrations/status", "GET").Return(&domain.ApexRestRoute{ResponseStatus: 200, ResponseBody: `{"status":"UP"}`}, nil).Once()
	if st, body, err := svc.ExecuteApexRest(ctx, "v1/integrations/status", "", ""); err != nil || st != 200 || body == "" {
		t.Fatal(err)
	}
	repo.EXPECT().GetApexRestRoute(ctx, "/v1/custom", "POST").Return(nil, domain.ErrNotFound).Once()
	if st, _, err := svc.ExecuteApexRest(ctx, "/v1/custom", "POST", `{"hello":"world"}`); err != nil || st != 200 {
		t.Fatal(err)
	}
	repo.EXPECT().GetApexRestRoute(ctx, "/other", "GET").Return(nil, domain.ErrNotFound).Once()
	if _, _, err := svc.ExecuteApexRest(ctx, "/other", "GET", ""); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	// ProcessCometD
	if _, err := svc.ProcessCometD(ctx, "v1.0", nil); !errors.Is(err, domain.ErrInvalidVersion) {
		t.Fatal("expected ErrInvalidVersion")
	}
	// Handshake error
	repo.EXPECT().UpsertCometDClient(ctx, mock.Anything).Return(nil, errors.New("err")).Once()
	if _, err := svc.ProcessCometD(ctx, "v61.0", []domain.BayeuxMessage{{Channel: "/meta/handshake"}}); err == nil {
		t.Fatal("expected err")
	}
	// Full CometD lifecycle
	repo.EXPECT().UpsertCometDClient(ctx, mock.Anything).Return(&domain.CometDClient{ClientID: "c1"}, nil).Once()
	repo.EXPECT().GetCometDClient(ctx, "unknown").Return(nil, domain.ErrNotFound).Once()
	repo.EXPECT().GetCometDClient(ctx, "c1").Return(&domain.CometDClient{ClientID: "c1"}, nil).Once()
	repo.EXPECT().ListCometDSubscriptions(ctx, "c1").Return([]domain.CometDSubscription{{ID: 1, ClientID: "c1", ChannelName: "/event/Order__e", LastReplayID: 0}}, nil).Once()
	repo.EXPECT().ListPlatformEventsSince(ctx, "/event/Order__e", int64(0)).Return([]domain.PlatformEvent{{ID: "e1", ReplayID: 5, Payload: `{"k":"v"}`}}, nil).Once()
	repo.EXPECT().UpdateCometDSubscriptionReplayID(ctx, int64(1), int64(5)).Return(nil).Once()
	repo.EXPECT().UpsertCometDSubscription(ctx, "c1", "/event/Order__e", int64(2)).Return(&domain.CometDSubscription{ID: 1}, nil).Once()
	repo.EXPECT().DeleteCometDSubscription(ctx, "c1", "/event/Order__e").Return(nil).Once()
	repo.EXPECT().DisconnectCometDClient(ctx, "c1").Return(nil).Once()
	repo.EXPECT().CreatePlatformEvent(ctx, "/event/Order__e", "Order__e", mock.Anything).Return(&domain.PlatformEvent{ID: "e2"}, nil).Once()

	msgs := []domain.BayeuxMessage{
		{Channel: "/meta/handshake", ID: "1"},
		{Channel: "/meta/connect", ID: "2"},
		{Channel: "/meta/connect", ID: "3", ClientID: "unknown"},
		{Channel: "/meta/connect", ID: "4", ClientID: "c1"},
		{Channel: "/meta/subscribe", ID: "5", ClientID: "c1", Subscription: "/event/Order__e", Ext: map[string]any{"replay": map[string]any{"/event/Order__e": float64(2)}}},
		{Channel: "/meta/unsubscribe", ID: "6", ClientID: "c1", Subscription: "/event/Order__e"},
		{Channel: "/meta/disconnect", ID: "7", ClientID: "c1"},
		{Channel: "/event/Order__e", ID: "8", ClientID: "c1", Data: map[string]any{"Order_Number__c": "100"}},
		{Channel: "/unsupported/channel", ID: "9", ClientID: "c1"},
	}
	if resps, err := svc.ProcessCometD(ctx, "v61.0", msgs); err != nil || len(resps) != 10 {
		t.Fatalf("unexpected cometd responses: %d, err=%v", len(resps), err)
	}

	// CometD error branches
	repo.EXPECT().GetCometDClient(ctx, "c1").Return(&domain.CometDClient{ClientID: "c1"}, nil).Once()
	repo.EXPECT().ListCometDSubscriptions(ctx, "c1").Return(nil, errors.New("sub err")).Once()
	if _, err := svc.ProcessCometD(ctx, "v61.0", []domain.BayeuxMessage{{Channel: "/meta/connect", ClientID: "c1"}}); err == nil {
		t.Fatal("expected err")
	}
	repo.EXPECT().GetCometDClient(ctx, "c1").Return(&domain.CometDClient{ClientID: "c1"}, nil).Once()
	repo.EXPECT().ListCometDSubscriptions(ctx, "c1").Return([]domain.CometDSubscription{{ID: 1, ChannelName: "/event/Order__e"}}, nil).Once()
	repo.EXPECT().ListPlatformEventsSince(ctx, "/event/Order__e", int64(0)).Return(nil, errors.New("ev err")).Once()
	if _, err := svc.ProcessCometD(ctx, "v61.0", []domain.BayeuxMessage{{Channel: "/meta/connect", ClientID: "c1"}}); err == nil {
		t.Fatal("expected err")
	}
	repo.EXPECT().UpsertCometDSubscription(ctx, "c1", "/event/Order__e", int64(0)).Return(nil, errors.New("err")).Once()
	if _, err := svc.ProcessCometD(ctx, "v61.0", []domain.BayeuxMessage{{Channel: "/meta/subscribe", ClientID: "c1", Subscription: "/event/Order__e"}}); err == nil {
		t.Fatal("expected err")
	}
	repo.EXPECT().DeleteCometDSubscription(ctx, "c1", "/event/Order__e").Return(errors.New("err")).Once()
	if _, err := svc.ProcessCometD(ctx, "v61.0", []domain.BayeuxMessage{{Channel: "/meta/unsubscribe", ClientID: "c1", Subscription: "/event/Order__e"}}); err == nil {
		t.Fatal("expected err")
	}
	repo.EXPECT().DisconnectCometDClient(ctx, "c1").Return(errors.New("err")).Once()
	if _, err := svc.ProcessCometD(ctx, "v61.0", []domain.BayeuxMessage{{Channel: "/meta/disconnect", ClientID: "c1"}}); err == nil {
		t.Fatal("expected err")
	}
	repo.EXPECT().CreatePlatformEvent(ctx, "/event/Order__e", "Order__e", mock.Anything).Return(nil, errors.New("err")).Once()
	if _, err := svc.ProcessCometD(ctx, "v61.0", []domain.BayeuxMessage{{Channel: "/event/Order__e", ClientID: "c1"}}); err == nil {
		t.Fatal("expected err")
	}
}
