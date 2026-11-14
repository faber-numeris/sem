package postgres

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/faber-numeris/sem/internal/adapters/outbound/postgres/gen"
	"github.com/faber-numeris/sem/internal/core/domain"
	"github.com/faber-numeris/sem/internal/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
)

type fakeDBTXWithPing struct {
	err error
}

func (f *fakeDBTXWithPing) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (f *fakeDBTXWithPing) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, nil
}
func (f *fakeDBTXWithPing) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row { return nil }
func (f *fakeDBTXWithPing) Ping(ctx context.Context) error                                { return f.err }

type fakeDBTXWithoutPing struct{}

func (f *fakeDBTXWithoutPing) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (f *fakeDBTXWithoutPing) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, nil
}
func (f *fakeDBTXWithoutPing) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return nil
}

func TestRepository_PingAndHelpers(t *testing.T) {
	r1 := NewRepository(&fakeDBTXWithPing{err: nil})
	if !r1.Ping() {
		t.Fatal("expected Ping true")
	}
	r2 := NewRepository(&fakeDBTXWithPing{err: errors.New("down")})
	if r2.Ping() {
		t.Fatal("expected Ping false")
	}
	r3 := NewRepository(&fakeDBTXWithoutPing{})
	if !r3.Ping() {
		t.Fatal("expected Ping true when Pinger not implemented")
	}
	q := mocks.NewMockQuerier(t)
	r4 := NewRepositoryWithQuerier(q, nil)
	if !r4.Ping() {
		t.Fatal("expected Ping true when pingFn nil")
	}
	r5 := NewRepositoryWithQuerier(q, func() bool { return false })
	if r5.Ping() {
		t.Fatal("expected Ping false from custom pingFn")
	}

	now := time.Date(2026, 2, 2, 12, 0, 0, 0, time.UTC)
	if formatSFTimePtr(nil) != nil {
		t.Fatal("expected nil for nil time ptr")
	}
	if formatSFTimePtr(&now) == nil {
		t.Fatal("expected formatted string")
	}
	attr := makeAttributes("", "Account", "001xx")
	if attr.URL != "/services/data/v61.0/sobjects/Account/001xx" {
		t.Fatalf("unexpected url: %s", attr.URL)
	}
	tAttr := makeToolingAttributes("", "ApexClass", "01pxx")
	if tAttr.URL != "/services/data/v61.0/tooling/sobjects/ApexClass/01pxx" {
		t.Fatalf("unexpected tooling url: %s", tAttr.URL)
	}

	m := map[string]any{
		"str":      "hello",
		"intVal":   123,
		"floatVal": 45.67,
		"numStr":   "89.5",
		"intStr":   "42",
		"boolVal":  true,
		"boolStr":  "true",
		"timeVal":  now,
		"timeStr":  "2026-02-02",
		"badType":  struct{}{},
		"nilVal":   nil,
	}
	if getString(m, "STR", "") != "hello" {
		t.Fatal("expected case-insensitive lookup")
	}
	if getString(m, "intVal", "") != "123" {
		t.Fatal("expected stringified int")
	}
	if getString(m, "missing", "def") != "def" || getString(m, "nilVal", "def") != "def" {
		t.Fatal("expected default")
	}
	if getStringPtr(m, "missing") != nil || getStringPtr(m, "str") == nil {
		t.Fatal("unexpected getStringPtr")
	}
	if getFloatPtr(m, "floatVal") == nil || getFloatPtr(m, "intVal") == nil || getFloatPtr(m, "numStr") == nil || getFloatPtr(m, "badType") != nil || getFloatPtr(m, "missing") != nil {
		t.Fatal("unexpected getFloatPtr")
	}
	if getIntPtr(m, "intVal") == nil || getIntPtr(m, "floatVal") == nil || getIntPtr(m, "intStr") == nil || getIntPtr(m, "badType") != nil || getIntPtr(m, "missing") != nil {
		t.Fatal("unexpected getIntPtr")
	}
	if getBoolPtr(m, "boolVal") == nil || getBoolPtr(m, "boolStr") == nil || getBoolPtr(m, "badType") != nil || getBoolPtr(m, "missing") != nil {
		t.Fatal("unexpected getBoolPtr")
	}
	if getTimePtr(m, "timeVal") == nil || getTimePtr(m, "timeStr") == nil || getTimePtr(m, "badType") != nil || getTimePtr(m, "missing") != nil {
		t.Fatal("unexpected getTimePtr")
	}
}

func TestRepository_Auth(t *testing.T) {
	ctx := context.Background()
	q := mocks.NewMockQuerier(t)
	r := NewRepositoryWithQuerier(q, nil)

	q.EXPECT().GetOAuthClientByClientID(ctx, "c1").Return(gen.SalesforceOauthClient{ID: "1", ClientID: "c1"}, nil).Once()
	if c, err := r.GetClient(ctx, "c1"); err != nil || c.ClientID != "c1" {
		t.Fatalf("unexpected: %v %v", c, err)
	}
	q.EXPECT().GetOAuthClientByClientID(ctx, "bad").Return(gen.SalesforceOauthClient{}, errors.New("err")).Once()
	if _, err := r.GetClient(ctx, "bad"); !errors.Is(err, domain.ErrInvalidClient) {
		t.Fatalf("expected ErrInvalidClient, got %v", err)
	}

	q.EXPECT().GetUserByUsername(ctx, "u1").Return(gen.SalesforceSalesforceUser{ID: "005", Username: "u1"}, nil).Once()
	if u, err := r.GetUserByUsername(ctx, "u1"); err != nil || u.ID != "005" {
		t.Fatal(err)
	}
	q.EXPECT().GetUserByUsername(ctx, "bad").Return(gen.SalesforceSalesforceUser{}, errors.New("err")).Once()
	if _, err := r.GetUserByUsername(ctx, "bad"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().GetUserByID(ctx, "005").Return(gen.SalesforceSalesforceUser{ID: "005"}, nil).Once()
	if u, err := r.GetUserByID(ctx, "005"); err != nil || u.ID != "005" {
		t.Fatal(err)
	}
	q.EXPECT().GetUserByID(ctx, "bad").Return(gen.SalesforceSalesforceUser{}, errors.New("err")).Once()
	if _, err := r.GetUserByID(ctx, "bad"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().GetDefaultUser(ctx).Return(gen.SalesforceSalesforceUser{ID: "005"}, nil).Once()
	if u, err := r.GetDefaultUser(ctx); err != nil || u.ID != "005" {
		t.Fatal(err)
	}
	q.EXPECT().GetDefaultUser(ctx).Return(gen.SalesforceSalesforceUser{}, errors.New("err")).Once()
	if _, err := r.GetDefaultUser(ctx); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().CreateAuthCode(ctx, mock.Anything).Return(gen.SalesforceOauthAuthCode{Code: "code1"}, nil).Once()
	if c, err := r.CreateAuthCode(ctx, &domain.OAuthAuthCode{Code: "code1"}); err != nil || c.Code != "code1" {
		t.Fatal(err)
	}
	q.EXPECT().CreateAuthCode(ctx, mock.Anything).Return(gen.SalesforceOauthAuthCode{}, errors.New("err")).Once()
	if _, err := r.CreateAuthCode(ctx, &domain.OAuthAuthCode{Code: "code1"}); err == nil {
		t.Fatal("expected error")
	}

	q.EXPECT().GetAuthCode(ctx, "code1").Return(gen.SalesforceOauthAuthCode{Code: "code1"}, nil).Once()
	if c, err := r.GetAuthCode(ctx, "code1"); err != nil || c.Code != "code1" {
		t.Fatal(err)
	}
	q.EXPECT().GetAuthCode(ctx, "bad").Return(gen.SalesforceOauthAuthCode{}, errors.New("err")).Once()
	if _, err := r.GetAuthCode(ctx, "bad"); !errors.Is(err, domain.ErrInvalidGrant) {
		t.Fatal("expected ErrInvalidGrant")
	}

	q.EXPECT().MarkAuthCodeUsed(ctx, "code1").Return(nil).Once()
	if err := r.MarkAuthCodeUsed(ctx, "code1"); err != nil {
		t.Fatal(err)
	}

	rtVal := "rt1"
	q.EXPECT().CreateOAuthToken(ctx, mock.Anything).Return(gen.SalesforceOauthToken{AccessToken: "at1", RefreshToken: &rtVal}, nil).Once()
	if tok, err := r.CreateToken(ctx, &domain.OAuthToken{AccessToken: "at1", RefreshToken: "rt1"}); err != nil || tok.AccessToken != "at1" {
		t.Fatal(err)
	}
	q.EXPECT().CreateOAuthToken(ctx, mock.Anything).Return(gen.SalesforceOauthToken{}, errors.New("err")).Once()
	if _, err := r.CreateToken(ctx, &domain.OAuthToken{}); err == nil {
		t.Fatal("expected error")
	}

	q.EXPECT().GetOAuthTokenByAccessToken(ctx, "at1").Return(gen.SalesforceOauthToken{AccessToken: "at1"}, nil).Once()
	if tok, err := r.GetTokenByAccess(ctx, "at1"); err != nil || tok.AccessToken != "at1" {
		t.Fatal(err)
	}
	q.EXPECT().GetOAuthTokenByAccessToken(ctx, "bad").Return(gen.SalesforceOauthToken{}, errors.New("err")).Once()
	if _, err := r.GetTokenByAccess(ctx, "bad"); !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatal("expected ErrInvalidToken")
	}

	q.EXPECT().GetOAuthTokenByRefreshToken(ctx, "rt1").Return(gen.SalesforceOauthToken{AccessToken: "at1"}, nil).Once()
	if tok, err := r.GetTokenByRefresh(ctx, "rt1"); err != nil || tok.AccessToken != "at1" {
		t.Fatal(err)
	}
	q.EXPECT().GetOAuthTokenByRefreshToken(ctx, "bad").Return(gen.SalesforceOauthToken{}, errors.New("err")).Once()
	if _, err := r.GetTokenByRefresh(ctx, "bad"); !errors.Is(err, domain.ErrInvalidGrant) {
		t.Fatal("expected ErrInvalidGrant")
	}

	q.EXPECT().GetOAuthTokenByAnyToken(ctx, "any").Return(gen.SalesforceOauthToken{AccessToken: "at1"}, nil).Once()
	if tok, err := r.GetTokenByAny(ctx, "any"); err != nil || tok.AccessToken != "at1" {
		t.Fatal(err)
	}
	q.EXPECT().GetOAuthTokenByAnyToken(ctx, "bad").Return(gen.SalesforceOauthToken{}, errors.New("err")).Once()
	if _, err := r.GetTokenByAny(ctx, "bad"); !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatal("expected ErrInvalidToken")
	}

	q.EXPECT().RevokeOAuthToken(ctx, "any").Return(nil).Once()
	if err := r.RevokeToken(ctx, "any"); err != nil {
		t.Fatal(err)
	}
}

func TestRepository_Metadata(t *testing.T) {
	ctx := context.Background()
	q := mocks.NewMockQuerier(t)
	r := NewRepositoryWithQuerier(q, nil)

	q.EXPECT().ListOrgLimits(ctx).Return([]gen.SalesforceOrgLimit{{LimitName: "DailyApiRequests", MaxLimit: 100, RemainingLimit: 99}}, nil).Once()
	if lims, err := r.ListOrgLimits(ctx); err != nil || len(lims) != 1 {
		t.Fatal(err)
	}
	q.EXPECT().ListOrgLimits(ctx).Return(nil, errors.New("err")).Once()
	if _, err := r.ListOrgLimits(ctx); err == nil {
		t.Fatal("expected error")
	}

	q.EXPECT().ListSObjectDefinitions(ctx, false).Return([]gen.SalesforceSobjectDefinition{{Name: "Account"}}, nil).Once()
	if defs, err := r.ListSObjectDefinitions(ctx, false); err != nil || len(defs) != 1 {
		t.Fatal(err)
	}
	q.EXPECT().ListSObjectDefinitions(ctx, false).Return(nil, errors.New("err")).Once()
	if _, err := r.ListSObjectDefinitions(ctx, false); err == nil {
		t.Fatal("expected error")
	}

	ref := "Account"
	q.EXPECT().GetSObjectDefinition(ctx, "Contact").Return(gen.SalesforceSobjectDefinition{Name: "Contact"}, nil).Once()
	q.EXPECT().ListSObjectFieldDefinitions(ctx, "Contact").Return([]gen.SalesforceSobjectFieldDefinition{
		{Name: "Id", ReferenceTo: nil},
		{Name: "AccountId", ReferenceTo: &ref},
	}, nil).Once()
	if desc, err := r.GetSObjectDescribe(ctx, "Contact"); err != nil || len(desc.Fields) != 2 {
		t.Fatal(err)
	}

	q.EXPECT().GetSObjectDefinition(ctx, "Missing").Return(gen.SalesforceSobjectDefinition{}, errors.New("err")).Once()
	if _, err := r.GetSObjectDescribe(ctx, "Missing"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().GetSObjectDefinition(ctx, "Contact").Return(gen.SalesforceSobjectDefinition{Name: "Contact"}, nil).Once()
	q.EXPECT().ListSObjectFieldDefinitions(ctx, "Contact").Return(nil, errors.New("err")).Once()
	if _, err := r.GetSObjectDescribe(ctx, "Contact"); err == nil {
		t.Fatal("expected error")
	}
}

func TestRepository_CreateGetUpdateDeleteUpsertListSearch(t *testing.T) {
	ctx := context.Background()
	q := mocks.NewMockQuerier(t)
	r := NewRepositoryWithQuerier(q, nil)

	// CreateRecord: Account validation & success & error
	if _, _, err := r.CreateRecord(ctx, "Account", map[string]any{}); !errors.Is(err, domain.ErrRequiredFieldMissing) {
		t.Fatal("expected ErrRequiredFieldMissing")
	}
	q.EXPECT().CreateAccount(ctx, mock.Anything).Return(gen.SalesforceAccount{ID: "001A", Name: "Acme"}, nil).Once()
	if res, _, err := r.CreateRecord(ctx, "Account", map[string]any{"Name": "Acme"}); err != nil || res.ID != "001A" {
		t.Fatal(err)
	}
	q.EXPECT().CreateAccount(ctx, mock.Anything).Return(gen.SalesforceAccount{}, errors.New("db err")).Once()
	if _, _, err := r.CreateRecord(ctx, "Account", map[string]any{"Name": "Acme"}); err == nil {
		t.Fatal("expected err")
	}

	// Contact validation & FirstName+LastName & LastName-only & error
	if _, _, err := r.CreateRecord(ctx, "Contact", map[string]any{}); !errors.Is(err, domain.ErrRequiredFieldMissing) {
		t.Fatal("expected ErrRequiredFieldMissing")
	}
	q.EXPECT().CreateContact(ctx, mock.Anything).Return(gen.SalesforceContact{ID: "003A", LastName: "Smith"}, nil).Once()
	if res, _, err := r.CreateRecord(ctx, "Contact", map[string]any{"FirstName": "Jane", "LastName": "Smith"}); err != nil || res.ID != "003A" {
		t.Fatal(err)
	}
	q.EXPECT().CreateContact(ctx, mock.Anything).Return(gen.SalesforceContact{ID: "003B", LastName: "Doe"}, nil).Once()
	if res, _, err := r.CreateRecord(ctx, "Contact", map[string]any{"LastName": "Doe"}); err != nil || res.ID != "003B" {
		t.Fatal(err)
	}
	q.EXPECT().CreateContact(ctx, mock.Anything).Return(gen.SalesforceContact{}, errors.New("db err")).Once()
	if _, _, err := r.CreateRecord(ctx, "Contact", map[string]any{"LastName": "Doe"}); err == nil {
		t.Fatal("expected err")
	}

	// Lead validation & FirstName+LastName & LastName-only & error
	if _, _, err := r.CreateRecord(ctx, "Lead", map[string]any{}); !errors.Is(err, domain.ErrRequiredFieldMissing) {
		t.Fatal("expected ErrRequiredFieldMissing")
	}
	q.EXPECT().CreateLead(ctx, mock.Anything).Return(gen.SalesforceLead{ID: "00QA"}, nil).Once()
	if res, _, err := r.CreateRecord(ctx, "Lead", map[string]any{"FirstName": "Bob", "LastName": "Lee"}); err != nil || res.ID != "00QA" {
		t.Fatal(err)
	}
	q.EXPECT().CreateLead(ctx, mock.Anything).Return(gen.SalesforceLead{ID: "00QB"}, nil).Once()
	if res, _, err := r.CreateRecord(ctx, "Lead", map[string]any{"LastName": "Lee"}); err != nil || res.ID != "00QB" {
		t.Fatal(err)
	}
	q.EXPECT().CreateLead(ctx, mock.Anything).Return(gen.SalesforceLead{}, errors.New("db err")).Once()
	if _, _, err := r.CreateRecord(ctx, "Lead", map[string]any{"LastName": "Lee"}); err == nil {
		t.Fatal("expected err")
	}

	// Case success & error
	q.EXPECT().CreateCase(ctx, mock.Anything).Return(gen.SalesforceCase{ID: "500A"}, nil).Once()
	if res, _, err := r.CreateRecord(ctx, "Case", map[string]any{"Subject": "Help"}); err != nil || res.ID != "500A" {
		t.Fatal(err)
	}
	q.EXPECT().CreateCase(ctx, mock.Anything).Return(gen.SalesforceCase{}, errors.New("db err")).Once()
	if _, _, err := r.CreateRecord(ctx, "Case", map[string]any{"Subject": "Help"}); err == nil {
		t.Fatal("expected err")
	}

	// Custom_Object__c success & error
	q.EXPECT().CreateCustomObject(ctx, mock.Anything).Return(gen.SalesforceCustomObjectC{ID: "a00A"}, nil).Once()
	if res, _, err := r.CreateRecord(ctx, "Custom_Object__c", map[string]any{"Name": "C1"}); err != nil || res.ID != "a00A" {
		t.Fatal(err)
	}
	q.EXPECT().CreateCustomObject(ctx, mock.Anything).Return(gen.SalesforceCustomObjectC{}, errors.New("db err")).Once()
	if _, _, err := r.CreateRecord(ctx, "Custom_Object__c", map[string]any{"Name": "C1"}); err == nil {
		t.Fatal("expected err")
	}

	// ContentDocument success & error
	q.EXPECT().CreateContentDocument(ctx, mock.Anything).Return(gen.SalesforceContentDocument{ID: "069A"}, nil).Once()
	if res, _, err := r.CreateRecord(ctx, "ContentDocument", map[string]any{"Title": "Doc"}); err != nil || res.ID != "069A" {
		t.Fatal(err)
	}
	q.EXPECT().CreateContentDocument(ctx, mock.Anything).Return(gen.SalesforceContentDocument{}, errors.New("db err")).Once()
	if _, _, err := r.CreateRecord(ctx, "ContentDocument", map[string]any{"Title": "Doc"}); err == nil {
		t.Fatal("expected err")
	}

	// ContentVersion auto-doc creation & explicit docID & errors
	q.EXPECT().CreateContentDocument(ctx, mock.Anything).Return(gen.SalesforceContentDocument{ID: "069A"}, nil).Once()
	q.EXPECT().CreateContentVersion(ctx, mock.Anything).Return(gen.SalesforceContentVersion{ID: "068A", ContentDocumentID: "069A"}, nil).Once()
	if res, _, err := r.CreateRecord(ctx, "ContentVersion", map[string]any{"Title": "Ver"}); err != nil || res.ID != "068A" {
		t.Fatal(err)
	}
	q.EXPECT().CreateContentDocument(ctx, mock.Anything).Return(gen.SalesforceContentDocument{}, errors.New("doc err")).Once()
	if _, _, err := r.CreateRecord(ctx, "ContentVersion", map[string]any{"Title": "Ver"}); err == nil {
		t.Fatal("expected err")
	}
	q.EXPECT().CreateContentVersion(ctx, mock.Anything).Return(gen.SalesforceContentVersion{}, errors.New("cv err")).Once()
	if _, _, err := r.CreateRecord(ctx, "ContentVersion", map[string]any{"Title": "Ver", "ContentDocumentId": "069A"}); err == nil {
		t.Fatal("expected err")
	}

	// ContentDocumentLink success & error
	q.EXPECT().CreateContentDocumentLink(ctx, mock.Anything).Return(gen.SalesforceContentDocumentLink{ID: "06AA"}, nil).Once()
	if res, _, err := r.CreateRecord(ctx, "ContentDocumentLink", map[string]any{"ContentDocumentId": "069A", "LinkedEntityId": "001A"}); err != nil || res.ID != "06AA" {
		t.Fatal(err)
	}
	q.EXPECT().CreateContentDocumentLink(ctx, mock.Anything).Return(gen.SalesforceContentDocumentLink{}, errors.New("cdl err")).Once()
	if _, _, err := r.CreateRecord(ctx, "ContentDocumentLink", map[string]any{}); err == nil {
		t.Fatal("expected err")
	}

	// Platform Event (__e) success & error & invalid sobject
	q.EXPECT().CreatePlatformEvent(ctx, mock.Anything).Return(gen.SalesforcePlatformEvent{ID: "e00A", ReplayID: 1}, nil).Once()
	if res, _, err := r.CreateRecord(ctx, "Order_Event__e", map[string]any{"Order_Number__c": "100"}); err != nil || res.ID != "e00A" {
		t.Fatal(err)
	}
	q.EXPECT().CreatePlatformEvent(ctx, mock.Anything).Return(gen.SalesforcePlatformEvent{}, errors.New("pe err")).Once()
	if _, _, err := r.CreateRecord(ctx, "Order_Event__e", map[string]any{}); err == nil {
		t.Fatal("expected err")
	}
	if _, _, err := r.CreateRecord(ctx, "UnknownSObject", map[string]any{}); !errors.Is(err, domain.ErrInvalidSObject) {
		t.Fatal("expected ErrInvalidSObject")
	}

	// GetRecord for all types (success + error)
	q.EXPECT().GetAccountByID(ctx, "1").Return(gen.SalesforceAccount{ID: "1"}, nil).Once()
	if _, err := r.GetRecord(ctx, "v61.0", "Account", "1"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().GetAccountByID(ctx, "2").Return(gen.SalesforceAccount{}, errors.New("err")).Once()
	if _, err := r.GetRecord(ctx, "v61.0", "Account", "2"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().GetContactByID(ctx, "1").Return(gen.SalesforceContact{ID: "1"}, nil).Once()
	if _, err := r.GetRecord(ctx, "v61.0", "Contact", "1"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().GetContactByID(ctx, "2").Return(gen.SalesforceContact{}, errors.New("err")).Once()
	if _, err := r.GetRecord(ctx, "v61.0", "Contact", "2"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().GetLeadByID(ctx, "1").Return(gen.SalesforceLead{ID: "1"}, nil).Once()
	if _, err := r.GetRecord(ctx, "v61.0", "Lead", "1"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().GetLeadByID(ctx, "2").Return(gen.SalesforceLead{}, errors.New("err")).Once()
	if _, err := r.GetRecord(ctx, "v61.0", "Lead", "2"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().SelectCase(ctx, "1").Return(gen.SalesforceCase{ID: "1"}, nil).Once()
	if _, err := r.GetRecord(ctx, "v61.0", "Case", "1"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().SelectCase(ctx, "2").Return(gen.SalesforceCase{}, errors.New("err")).Once()
	if _, err := r.GetRecord(ctx, "v61.0", "Case", "2"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().GetCustomObjectByID(ctx, "1").Return(gen.SalesforceCustomObjectC{ID: "1"}, nil).Once()
	if _, err := r.GetRecord(ctx, "v61.0", "Custom_Object__c", "1"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().GetCustomObjectByID(ctx, "2").Return(gen.SalesforceCustomObjectC{}, errors.New("err")).Once()
	if _, err := r.GetRecord(ctx, "v61.0", "Custom_Object__c", "2"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().GetContentDocument(ctx, "1").Return(gen.SalesforceContentDocument{ID: "1"}, nil).Once()
	if _, err := r.GetRecord(ctx, "v61.0", "ContentDocument", "1"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().GetContentDocument(ctx, "2").Return(gen.SalesforceContentDocument{}, errors.New("err")).Once()
	if _, err := r.GetRecord(ctx, "v61.0", "ContentDocument", "2"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().GetContentVersion(ctx, "1").Return(gen.SalesforceContentVersion{ID: "1"}, nil).Once()
	if _, err := r.GetRecord(ctx, "v61.0", "ContentVersion", "1"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().GetContentVersion(ctx, "2").Return(gen.SalesforceContentVersion{}, errors.New("err")).Once()
	if _, err := r.GetRecord(ctx, "v61.0", "ContentVersion", "2"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().GetContentDocumentLink(ctx, "1").Return(gen.SalesforceContentDocumentLink{ID: "1"}, nil).Once()
	if _, err := r.GetRecord(ctx, "v61.0", "ContentDocumentLink", "1"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().GetContentDocumentLink(ctx, "2").Return(gen.SalesforceContentDocumentLink{}, errors.New("err")).Once()
	if _, err := r.GetRecord(ctx, "v61.0", "ContentDocumentLink", "2"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	if _, err := r.GetRecord(ctx, "v61.0", "Unknown", "1"); !errors.Is(err, domain.ErrInvalidSObject) {
		t.Fatal("expected ErrInvalidSObject")
	}

	// UpdateRecord for all types (success + error)
	q.EXPECT().UpdateAccount(ctx, mock.Anything).Return(gen.SalesforceAccount{ID: "1"}, nil).Once()
	if _, err := r.UpdateRecord(ctx, "Account", "1", map[string]any{"Name": "N"}); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().UpdateAccount(ctx, mock.Anything).Return(gen.SalesforceAccount{}, errors.New("err")).Once()
	if _, err := r.UpdateRecord(ctx, "Account", "2", map[string]any{}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().UpdateContact(ctx, mock.Anything).Return(gen.SalesforceContact{ID: "1"}, nil).Once()
	if _, err := r.UpdateRecord(ctx, "Contact", "1", map[string]any{}); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().UpdateContact(ctx, mock.Anything).Return(gen.SalesforceContact{}, errors.New("err")).Once()
	if _, err := r.UpdateRecord(ctx, "Contact", "2", map[string]any{}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().UpdateLead(ctx, mock.Anything).Return(gen.SalesforceLead{ID: "1"}, nil).Once()
	if _, err := r.UpdateRecord(ctx, "Lead", "1", map[string]any{}); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().UpdateLead(ctx, mock.Anything).Return(gen.SalesforceLead{}, errors.New("err")).Once()
	if _, err := r.UpdateRecord(ctx, "Lead", "2", map[string]any{}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().UpdateCase(ctx, mock.Anything).Return(gen.SalesforceCase{ID: "1"}, nil).Once()
	if _, err := r.UpdateRecord(ctx, "Case", "1", map[string]any{}); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().UpdateCase(ctx, mock.Anything).Return(gen.SalesforceCase{}, errors.New("err")).Once()
	if _, err := r.UpdateRecord(ctx, "Case", "2", map[string]any{}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().UpdateCustomObject(ctx, mock.Anything).Return(gen.SalesforceCustomObjectC{ID: "1"}, nil).Once()
	if _, err := r.UpdateRecord(ctx, "Custom_Object__c", "1", map[string]any{}); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().UpdateCustomObject(ctx, mock.Anything).Return(gen.SalesforceCustomObjectC{}, errors.New("err")).Once()
	if _, err := r.UpdateRecord(ctx, "Custom_Object__c", "2", map[string]any{}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().UpdateContentDocument(ctx, mock.Anything).Return(gen.SalesforceContentDocument{ID: "1"}, nil).Once()
	if _, err := r.UpdateRecord(ctx, "ContentDocument", "1", map[string]any{}); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().UpdateContentDocument(ctx, mock.Anything).Return(gen.SalesforceContentDocument{}, errors.New("err")).Once()
	if _, err := r.UpdateRecord(ctx, "ContentDocument", "2", map[string]any{}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	if _, err := r.UpdateRecord(ctx, "Unknown", "1", map[string]any{}); !errors.Is(err, domain.ErrInvalidSObject) {
		t.Fatal("expected ErrInvalidSObject")
	}

	// UpsertRecordByExternalID: by Id (update & create) and by External_ID__c for each type
	q.EXPECT().UpdateAccount(ctx, mock.Anything).Return(gen.SalesforceAccount{ID: "001A"}, nil).Once()
	if res, err := r.UpsertRecordByExternalID(ctx, "Account", "Id", "001A", map[string]any{"Name": "N"}); err != nil || res.Created {
		t.Fatal("expected update via Id")
	}
	q.EXPECT().UpdateAccount(ctx, mock.Anything).Return(gen.SalesforceAccount{}, errors.New("not found")).Once()
	q.EXPECT().CreateAccount(ctx, mock.Anything).Return(gen.SalesforceAccount{ID: "001B"}, nil).Once()
	if res, err := r.UpsertRecordByExternalID(ctx, "Account", "Id", "001B", map[string]any{"Name": "N"}); err != nil || !res.Created {
		t.Fatal("expected create via Id fallback")
	}

	q.EXPECT().GetAccountByExternalID(ctx, "EXT1").Return(gen.SalesforceAccount{ID: "001A"}, nil).Once()
	q.EXPECT().UpdateAccount(ctx, mock.Anything).Return(gen.SalesforceAccount{ID: "001A"}, nil).Once()
	if res, err := r.UpsertRecordByExternalID(ctx, "Account", "External_ID__c", "EXT1", map[string]any{"Name": "N"}); err != nil || res.Created {
		t.Fatal(err)
	}
	q.EXPECT().GetAccountByExternalID(ctx, "EXT1").Return(gen.SalesforceAccount{ID: "001A"}, nil).Once()
	q.EXPECT().UpdateAccount(ctx, mock.Anything).Return(gen.SalesforceAccount{}, errors.New("upd err")).Once()
	if _, err := r.UpsertRecordByExternalID(ctx, "Account", "External_ID__c", "EXT1", map[string]any{"Name": "N"}); err == nil {
		t.Fatal("expected error on update failure")
	}
	q.EXPECT().GetAccountByExternalID(ctx, "EXT2").Return(gen.SalesforceAccount{}, errors.New("missing")).Once()
	q.EXPECT().CreateAccount(ctx, mock.Anything).Return(gen.SalesforceAccount{ID: "001C"}, nil).Once()
	if res, err := r.UpsertRecordByExternalID(ctx, "Account", "External_ID__c", "EXT2", map[string]any{"Name": "N"}); err != nil || !res.Created {
		t.Fatal(err)
	}

	q.EXPECT().GetContactByExternalID(ctx, "EXT1").Return(gen.SalesforceContact{ID: "003A"}, nil).Once()
	q.EXPECT().UpdateContact(ctx, mock.Anything).Return(gen.SalesforceContact{ID: "003A"}, nil).Once()
	if _, err := r.UpsertRecordByExternalID(ctx, "Contact", "External_ID__c", "EXT1", map[string]any{"LastName": "L"}); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().GetLeadByExternalID(ctx, "EXT1").Return(gen.SalesforceLead{ID: "00QA"}, nil).Once()
	q.EXPECT().UpdateLead(ctx, mock.Anything).Return(gen.SalesforceLead{ID: "00QA"}, nil).Once()
	if _, err := r.UpsertRecordByExternalID(ctx, "Lead", "External_ID__c", "EXT1", map[string]any{"LastName": "L"}); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().GetCaseByExternalID(ctx, "EXT1").Return(gen.SalesforceCase{ID: "500A"}, nil).Once()
	q.EXPECT().UpdateCase(ctx, mock.Anything).Return(gen.SalesforceCase{ID: "500A"}, nil).Once()
	if _, err := r.UpsertRecordByExternalID(ctx, "Case", "External_ID__c", "EXT1", map[string]any{"Subject": "S"}); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().GetCustomObjectByExternalID(ctx, "EXT1").Return(gen.SalesforceCustomObjectC{ID: "a00A"}, nil).Once()
	q.EXPECT().UpdateCustomObject(ctx, mock.Anything).Return(gen.SalesforceCustomObjectC{ID: "a00A"}, nil).Once()
	if _, err := r.UpsertRecordByExternalID(ctx, "Custom_Object__c", "External_ID__c", "EXT1", map[string]any{"Name": "N"}); err != nil {
		t.Fatal(err)
	}
	if _, err := r.UpsertRecordByExternalID(ctx, "Unknown", "External_ID__c", "EXT1", map[string]any{}); !errors.Is(err, domain.ErrInvalidSObject) {
		t.Fatal("expected ErrInvalidSObject")
	}

	// DeleteRecord for all types
	q.EXPECT().SoftDeleteAccount(ctx, "1").Return(int64(1), nil).Once()
	if err := r.DeleteRecord(ctx, "Account", "1"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().SoftDeleteContact(ctx, "1").Return(int64(1), nil).Once()
	if err := r.DeleteRecord(ctx, "Contact", "1"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().SoftDeleteLead(ctx, "1").Return(int64(1), nil).Once()
	if err := r.DeleteRecord(ctx, "Lead", "1"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().SoftDeleteCase(ctx, "1").Return(int64(1), nil).Once()
	if err := r.DeleteRecord(ctx, "Case", "1"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().SoftDeleteCustomObject(ctx, "1").Return(int64(1), nil).Once()
	if err := r.DeleteRecord(ctx, "Custom_Object__c", "1"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().SoftDeleteContentDocument(ctx, "1").Return(int64(1), nil).Once()
	if err := r.DeleteRecord(ctx, "ContentDocument", "1"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().SoftDeleteContentVersion(ctx, "1").Return(int64(1), nil).Once()
	if err := r.DeleteRecord(ctx, "ContentVersion", "1"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().SoftDeleteContentDocumentLink(ctx, "1").Return(int64(1), nil).Once()
	if err := r.DeleteRecord(ctx, "ContentDocumentLink", "1"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().SoftDeleteAccount(ctx, "0").Return(int64(0), nil).Once()
	if err := r.DeleteRecord(ctx, "Account", "0"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}
	q.EXPECT().SoftDeleteAccount(ctx, "err").Return(int64(0), errors.New("db err")).Once()
	if err := r.DeleteRecord(ctx, "Account", "err"); err == nil {
		t.Fatal("expected db err")
	}
	if err := r.DeleteRecord(ctx, "Unknown", "1"); !errors.Is(err, domain.ErrInvalidSObject) {
		t.Fatal("expected ErrInvalidSObject")
	}

	// GetContentVersionData: base64, raw string, not found
	b64 := base64.StdEncoding.EncodeToString([]byte("hello world"))
	q.EXPECT().GetContentVersion(ctx, "1").Return(gen.SalesforceContentVersion{ID: "1", VersionData: b64}, nil).Once()
	if data, err := r.GetContentVersionData(ctx, "1"); err != nil || string(data) != "hello world" {
		t.Fatalf("unexpected data: %s %v", string(data), err)
	}
	q.EXPECT().GetContentVersion(ctx, "2").Return(gen.SalesforceContentVersion{ID: "2", VersionData: "!!!not-base64!!!"}, nil).Once()
	if data, err := r.GetContentVersionData(ctx, "2"); err != nil || string(data) != "!!!not-base64!!!" {
		t.Fatal("expected raw bytes fallback")
	}
	q.EXPECT().GetContentVersion(ctx, "3").Return(gen.SalesforceContentVersion{}, errors.New("err")).Once()
	if _, err := r.GetContentVersionData(ctx, "3"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	// ListRecordsForSObject (includeDeleted false/true + errors)
	q.EXPECT().ListAccounts(ctx).Return([]gen.SalesforceAccount{{ID: "1"}}, nil).Once()
	if list, err := r.ListRecordsForSObject(ctx, "v61.0", "Account", false); err != nil || len(list) != 1 {
		t.Fatal(err)
	}
	q.EXPECT().ListAccountsAll(ctx).Return([]gen.SalesforceAccount{{ID: "1"}}, nil).Once()
	if _, err := r.ListRecordsForSObject(ctx, "v61.0", "Account", true); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().ListAccounts(ctx).Return(nil, errors.New("err")).Once()
	if _, err := r.ListRecordsForSObject(ctx, "v61.0", "Account", false); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().ListContacts(ctx).Return([]gen.SalesforceContact{{ID: "1"}}, nil).Once()
	if _, err := r.ListRecordsForSObject(ctx, "v61.0", "Contact", false); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().ListContactsAll(ctx).Return([]gen.SalesforceContact{{ID: "1"}}, nil).Once()
	if _, err := r.ListRecordsForSObject(ctx, "v61.0", "Contact", true); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().ListContacts(ctx).Return(nil, errors.New("err")).Once()
	if _, err := r.ListRecordsForSObject(ctx, "v61.0", "Contact", false); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().ListLeads(ctx).Return([]gen.SalesforceLead{{ID: "1"}}, nil).Once()
	if _, err := r.ListRecordsForSObject(ctx, "v61.0", "Lead", false); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().ListLeadsAll(ctx).Return([]gen.SalesforceLead{{ID: "1"}}, nil).Once()
	if _, err := r.ListRecordsForSObject(ctx, "v61.0", "Lead", true); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().ListLeads(ctx).Return(nil, errors.New("err")).Once()
	if _, err := r.ListRecordsForSObject(ctx, "v61.0", "Lead", false); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().ListCases(ctx).Return([]gen.SalesforceCase{{ID: "1"}}, nil).Once()
	if _, err := r.ListRecordsForSObject(ctx, "v61.0", "Case", false); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().ListCasesAll(ctx).Return([]gen.SalesforceCase{{ID: "1"}}, nil).Once()
	if _, err := r.ListRecordsForSObject(ctx, "v61.0", "Case", true); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().ListCases(ctx).Return(nil, errors.New("err")).Once()
	if _, err := r.ListRecordsForSObject(ctx, "v61.0", "Case", false); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().ListCustomObjects(ctx).Return([]gen.SalesforceCustomObjectC{{ID: "1"}}, nil).Once()
	if _, err := r.ListRecordsForSObject(ctx, "v61.0", "Custom_Object__c", false); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().ListCustomObjectsAll(ctx).Return([]gen.SalesforceCustomObjectC{{ID: "1"}}, nil).Once()
	if _, err := r.ListRecordsForSObject(ctx, "v61.0", "Custom_Object__c", true); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().ListCustomObjects(ctx).Return(nil, errors.New("err")).Once()
	if _, err := r.ListRecordsForSObject(ctx, "v61.0", "Custom_Object__c", false); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().ListContentDocuments(ctx).Return([]gen.SalesforceContentDocument{{ID: "1"}}, nil).Once()
	if _, err := r.ListRecordsForSObject(ctx, "v61.0", "ContentDocument", false); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().ListContentDocuments(ctx).Return(nil, errors.New("err")).Once()
	if _, err := r.ListRecordsForSObject(ctx, "v61.0", "ContentDocument", false); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().ListContentVersions(ctx).Return([]gen.SalesforceContentVersion{{ID: "1"}}, nil).Once()
	if _, err := r.ListRecordsForSObject(ctx, "v61.0", "ContentVersion", false); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().ListContentVersions(ctx).Return(nil, errors.New("err")).Once()
	if _, err := r.ListRecordsForSObject(ctx, "v61.0", "ContentVersion", false); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().ListContentDocumentLinks(ctx).Return([]gen.SalesforceContentDocumentLink{{ID: "1"}}, nil).Once()
	if _, err := r.ListRecordsForSObject(ctx, "v61.0", "ContentDocumentLink", false); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().ListContentDocumentLinks(ctx).Return(nil, errors.New("err")).Once()
	if _, err := r.ListRecordsForSObject(ctx, "v61.0", "ContentDocumentLink", false); err == nil {
		t.Fatal("expected err")
	}

	if _, err := r.ListRecordsForSObject(ctx, "v61.0", "Unknown", false); !errors.Is(err, domain.ErrInvalidSObject) {
		t.Fatal("expected ErrInvalidSObject")
	}

	// SearchRecordsForSObject
	q.EXPECT().SearchAccounts(ctx, "Acme").Return([]gen.SalesforceAccount{{ID: "1"}}, nil).Once()
	if _, err := r.SearchRecordsForSObject(ctx, "v61.0", "Account", "Acme"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().SearchAccounts(ctx, "Acme").Return(nil, errors.New("err")).Once()
	if _, err := r.SearchRecordsForSObject(ctx, "v61.0", "Account", "Acme"); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().SearchContacts(ctx, "Acme").Return([]gen.SalesforceContact{{ID: "1"}}, nil).Once()
	if _, err := r.SearchRecordsForSObject(ctx, "v61.0", "Contact", "Acme"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().SearchContacts(ctx, "Acme").Return(nil, errors.New("err")).Once()
	if _, err := r.SearchRecordsForSObject(ctx, "v61.0", "Contact", "Acme"); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().SearchLeads(ctx, "Acme").Return([]gen.SalesforceLead{{ID: "1"}}, nil).Once()
	if _, err := r.SearchRecordsForSObject(ctx, "v61.0", "Lead", "Acme"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().SearchLeads(ctx, "Acme").Return(nil, errors.New("err")).Once()
	if _, err := r.SearchRecordsForSObject(ctx, "v61.0", "Lead", "Acme"); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().SearchCases(ctx, "Acme").Return([]gen.SalesforceCase{{ID: "1"}}, nil).Once()
	if _, err := r.SearchRecordsForSObject(ctx, "v61.0", "Case", "Acme"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().SearchCases(ctx, "Acme").Return(nil, errors.New("err")).Once()
	if _, err := r.SearchRecordsForSObject(ctx, "v61.0", "Case", "Acme"); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().SearchCustomObjects(ctx, "Acme").Return([]gen.SalesforceCustomObjectC{{ID: "1"}}, nil).Once()
	if _, err := r.SearchRecordsForSObject(ctx, "v61.0", "Custom_Object__c", "Acme"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().SearchCustomObjects(ctx, "Acme").Return(nil, errors.New("err")).Once()
	if _, err := r.SearchRecordsForSObject(ctx, "v61.0", "Custom_Object__c", "Acme"); err == nil {
		t.Fatal("expected err")
	}

	if _, err := r.SearchRecordsForSObject(ctx, "v61.0", "Unknown", "Acme"); !errors.Is(err, domain.ErrInvalidSObject) {
		t.Fatal("expected ErrInvalidSObject")
	}
}

func TestRepository_QueryBulkTooling(t *testing.T) {
	ctx := context.Background()
	q := mocks.NewMockQuerier(t)
	r := NewRepositoryWithQuerier(q, nil)

	// Query cursors
	q.EXPECT().CreateQueryCursor(ctx, mock.Anything).Return(gen.SalesforceQueryCursor{ID: "cur1"}, nil).Once()
	if c, err := r.CreateQueryCursor(ctx, &domain.QueryCursor{ID: "cur1"}); err != nil || c.ID != "cur1" {
		t.Fatal(err)
	}
	q.EXPECT().CreateQueryCursor(ctx, mock.Anything).Return(gen.SalesforceQueryCursor{}, errors.New("err")).Once()
	if _, err := r.CreateQueryCursor(ctx, &domain.QueryCursor{ID: "cur1"}); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().GetQueryCursor(ctx, "cur1").Return(gen.SalesforceQueryCursor{ID: "cur1"}, nil).Once()
	if c, err := r.GetQueryCursor(ctx, "cur1"); err != nil || c.ID != "cur1" {
		t.Fatal(err)
	}
	q.EXPECT().GetQueryCursor(ctx, "bad").Return(gen.SalesforceQueryCursor{}, errors.New("err")).Once()
	if _, err := r.GetQueryCursor(ctx, "bad"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().UpdateQueryCursorOffset(ctx, mock.Anything).Return(gen.SalesforceQueryCursor{ID: "cur1", OffsetValue: 10}, nil).Once()
	if c, err := r.UpdateQueryCursorOffset(ctx, "cur1", 10); err != nil || c.OffsetValue != 10 {
		t.Fatal(err)
	}
	q.EXPECT().UpdateQueryCursorOffset(ctx, mock.Anything).Return(gen.SalesforceQueryCursor{}, errors.New("err")).Once()
	if _, err := r.UpdateQueryCursorOffset(ctx, "cur1", 10); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().DeleteQueryCursor(ctx, "cur1").Return(nil).Once()
	if err := r.DeleteQueryCursor(ctx, "cur1"); err != nil {
		t.Fatal(err)
	}

	// Bulk Ingest & Query
	extField := "External_ID__c"
	q.EXPECT().CreateBulkIngestJob(ctx, mock.Anything).Return(gen.SalesforceBulkIngestJob{ID: "750A", ExternalIDFieldName: &extField}, nil).Once()
	if j, err := r.CreateBulkIngestJob(ctx, &domain.BulkIngestJob{Operation: "insert", ExternalIDFieldName: "External_ID__c"}); err != nil || j.ID != "750A" {
		t.Fatal(err)
	}
	q.EXPECT().CreateBulkIngestJob(ctx, mock.Anything).Return(gen.SalesforceBulkIngestJob{}, errors.New("err")).Once()
	if _, err := r.CreateBulkIngestJob(ctx, &domain.BulkIngestJob{}); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().GetBulkIngestJob(ctx, "750A").Return(gen.SalesforceBulkIngestJob{ID: "750A"}, nil).Once()
	if j, err := r.GetBulkIngestJob(ctx, "750A"); err != nil || j.ID != "750A" {
		t.Fatal(err)
	}
	q.EXPECT().GetBulkIngestJob(ctx, "bad").Return(gen.SalesforceBulkIngestJob{}, errors.New("err")).Once()
	if _, err := r.GetBulkIngestJob(ctx, "bad"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().UpdateBulkIngestJobState(ctx, mock.Anything).Return(gen.SalesforceBulkIngestJob{ID: "750A", State: "Aborted"}, nil).Once()
	if j, err := r.UpdateBulkIngestJobState(ctx, "750A", "Aborted"); err != nil || j.State != "Aborted" {
		t.Fatal(err)
	}
	q.EXPECT().UpdateBulkIngestJobState(ctx, mock.Anything).Return(gen.SalesforceBulkIngestJob{}, errors.New("err")).Once()
	if _, err := r.UpdateBulkIngestJobState(ctx, "bad", "Aborted"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().UpdateBulkIngestJobStats(ctx, mock.Anything).Return(gen.SalesforceBulkIngestJob{ID: "750A", State: "JobComplete"}, nil).Once()
	if j, err := r.UpdateBulkIngestJobStats(ctx, "750A", "JobComplete", 5, 1); err != nil || j.State != "JobComplete" {
		t.Fatal(err)
	}
	q.EXPECT().UpdateBulkIngestJobStats(ctx, mock.Anything).Return(gen.SalesforceBulkIngestJob{}, errors.New("err")).Once()
	if _, err := r.UpdateBulkIngestJobStats(ctx, "bad", "JobComplete", 5, 1); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().ListBulkIngestJobs(ctx).Return([]gen.SalesforceBulkIngestJob{{ID: "750A"}}, nil).Once()
	if jobs, err := r.ListBulkIngestJobs(ctx); err != nil || len(jobs) != 1 {
		t.Fatal(err)
	}
	q.EXPECT().ListBulkIngestJobs(ctx).Return(nil, errors.New("err")).Once()
	if _, err := r.ListBulkIngestJobs(ctx); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().CreateBulkIngestBatch(ctx, mock.Anything).Return(gen.SalesforceBulkIngestBatch{ID: "751A"}, nil).Once()
	if b, err := r.CreateBulkIngestBatch(ctx, "750A", "Name\nAcme"); err != nil || b.ID != "751A" {
		t.Fatal(err)
	}
	q.EXPECT().CreateBulkIngestBatch(ctx, mock.Anything).Return(gen.SalesforceBulkIngestBatch{}, errors.New("err")).Once()
	if _, err := r.CreateBulkIngestBatch(ctx, "750A", ""); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().ListBulkIngestBatchesByJob(ctx, "750A").Return([]gen.SalesforceBulkIngestBatch{{ID: "751A"}}, nil).Once()
	if list, err := r.ListBulkIngestBatches(ctx, "750A"); err != nil || len(list) != 1 {
		t.Fatal(err)
	}
	q.EXPECT().ListBulkIngestBatchesByJob(ctx, "750A").Return(nil, errors.New("err")).Once()
	if _, err := r.ListBulkIngestBatches(ctx, "750A"); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().CreateBulkIngestResult(ctx, mock.Anything).Return(gen.SalesforceBulkIngestResult{ID: 1, JobID: "750A"}, nil).Once()
	if res, err := r.CreateBulkIngestResult(ctx, &domain.BulkIngestResult{JobID: "750A"}); err != nil || res.ID != 1 {
		t.Fatal(err)
	}
	q.EXPECT().CreateBulkIngestResult(ctx, mock.Anything).Return(gen.SalesforceBulkIngestResult{}, errors.New("err")).Once()
	if _, err := r.CreateBulkIngestResult(ctx, &domain.BulkIngestResult{}); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().ListBulkIngestResults(ctx, mock.Anything).Return([]gen.SalesforceBulkIngestResult{{ID: 1}}, nil).Once()
	if list, err := r.ListBulkIngestResults(ctx, "750A", true); err != nil || len(list) != 1 {
		t.Fatal(err)
	}
	q.EXPECT().ListBulkIngestResults(ctx, mock.Anything).Return(nil, errors.New("err")).Once()
	if _, err := r.ListBulkIngestResults(ctx, "750A", true); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().CreateBulkQueryJob(ctx, mock.Anything).Return(gen.SalesforceBulkQueryJob{ID: "750Q"}, nil).Once()
	if j, err := r.CreateBulkQueryJob(ctx, &domain.BulkQueryJob{Operation: "query"}); err != nil || j.ID != "750Q" {
		t.Fatal(err)
	}
	q.EXPECT().CreateBulkQueryJob(ctx, mock.Anything).Return(gen.SalesforceBulkQueryJob{}, errors.New("err")).Once()
	if _, err := r.CreateBulkQueryJob(ctx, &domain.BulkQueryJob{}); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().GetBulkQueryJob(ctx, "750Q").Return(gen.SalesforceBulkQueryJob{ID: "750Q"}, nil).Once()
	if j, err := r.GetBulkQueryJob(ctx, "750Q"); err != nil || j.ID != "750Q" {
		t.Fatal(err)
	}
	q.EXPECT().GetBulkQueryJob(ctx, "bad").Return(gen.SalesforceBulkQueryJob{}, errors.New("err")).Once()
	if _, err := r.GetBulkQueryJob(ctx, "bad"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	// Tooling
	q.EXPECT().ListApexClasses(ctx).Return([]gen.SalesforceApexClass{{ID: "01pA"}}, nil).Once()
	if _, err := r.ListToolingRecords(ctx, "v61.0", "ApexClass"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().ListApexClasses(ctx).Return(nil, errors.New("err")).Once()
	if _, err := r.ListToolingRecords(ctx, "v61.0", "ApexClass"); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().ListApexTriggers(ctx).Return([]gen.SalesforceApexTrigger{{ID: "01qA"}}, nil).Once()
	if _, err := r.ListToolingRecords(ctx, "v61.0", "ApexTrigger"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().ListApexTriggers(ctx).Return(nil, errors.New("err")).Once()
	if _, err := r.ListToolingRecords(ctx, "v61.0", "ApexTrigger"); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().ListCustomFields(ctx).Return([]gen.SalesforceCustomField{{ID: "00NA"}}, nil).Once()
	if _, err := r.ListToolingRecords(ctx, "v61.0", "CustomField"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().ListCustomFields(ctx).Return(nil, errors.New("err")).Once()
	if _, err := r.ListToolingRecords(ctx, "v61.0", "CustomField"); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().ListValidationRules(ctx).Return([]gen.SalesforceValidationRule{{ID: "03dA"}}, nil).Once()
	if _, err := r.ListToolingRecords(ctx, "v61.0", "ValidationRule"); err != nil {
		t.Fatal(err)
	}
	q.EXPECT().ListValidationRules(ctx).Return(nil, errors.New("err")).Once()
	if _, err := r.ListToolingRecords(ctx, "v61.0", "ValidationRule"); err == nil {
		t.Fatal("expected err")
	}

	if _, err := r.ListToolingRecords(ctx, "v61.0", "Unknown"); !errors.Is(err, domain.ErrInvalidSObject) {
		t.Fatal("expected ErrInvalidSObject")
	}

	q.EXPECT().CreateApexTestRun(ctx, mock.Anything).Return(gen.SalesforceApexTestRun{ID: "707A"}, nil).Once()
	if tr, err := r.CreateApexTestRun(ctx, "MyTest", "RunSpecifiedTests"); err != nil || tr.ID != "707A" {
		t.Fatal(err)
	}
	q.EXPECT().CreateApexTestRun(ctx, mock.Anything).Return(gen.SalesforceApexTestRun{}, errors.New("err")).Once()
	if _, err := r.CreateApexTestRun(ctx, "MyTest", "RunSpecifiedTests"); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().GetApexRestRoute(ctx, mock.Anything).Return(gen.SalesforceApexRestRoute{ID: "r1", ResponseStatus: 200}, nil).Once()
	if rt, err := r.GetApexRestRoute(ctx, "/v1/status", "GET"); err != nil || rt.ID != "r1" {
		t.Fatal(err)
	}
	q.EXPECT().GetApexRestRoute(ctx, mock.Anything).Return(gen.SalesforceApexRestRoute{}, errors.New("err")).Once()
	if _, err := r.GetApexRestRoute(ctx, "/missing", "GET"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().CreatePlatformEvent(ctx, mock.Anything).Return(gen.SalesforcePlatformEvent{ID: "e00A", ReplayID: 1}, nil).Once()
	if ev, err := r.CreatePlatformEvent(ctx, "/event/Test__e", "Test__e", `{}`); err != nil || ev.ID != "e00A" {
		t.Fatal(err)
	}
	q.EXPECT().CreatePlatformEvent(ctx, mock.Anything).Return(gen.SalesforcePlatformEvent{}, errors.New("err")).Once()
	if _, err := r.CreatePlatformEvent(ctx, "/event/Test__e", "Test__e", `{}`); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().ListPlatformEventsSince(ctx, mock.Anything).Return([]gen.SalesforcePlatformEvent{{ID: "e00A", ReplayID: 1}}, nil).Once()
	if evs, err := r.ListPlatformEventsSince(ctx, "/event/Test__e", 0); err != nil || len(evs) != 1 {
		t.Fatal(err)
	}
	q.EXPECT().ListPlatformEventsSince(ctx, mock.Anything).Return(nil, errors.New("err")).Once()
	if _, err := r.ListPlatformEventsSince(ctx, "/event/Test__e", 0); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().UpsertCometDClient(ctx, "c1").Return(gen.SalesforceCometdClient{ClientID: "c1", IsConnected: true}, nil).Once()
	if c, err := r.UpsertCometDClient(ctx, "c1"); err != nil || c.ClientID != "c1" {
		t.Fatal(err)
	}
	q.EXPECT().UpsertCometDClient(ctx, "c1").Return(gen.SalesforceCometdClient{}, errors.New("err")).Once()
	if _, err := r.UpsertCometDClient(ctx, "c1"); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().GetCometDClient(ctx, "c1").Return(gen.SalesforceCometdClient{ClientID: "c1"}, nil).Once()
	if c, err := r.GetCometDClient(ctx, "c1"); err != nil || c.ClientID != "c1" {
		t.Fatal(err)
	}
	q.EXPECT().GetCometDClient(ctx, "bad").Return(gen.SalesforceCometdClient{}, errors.New("err")).Once()
	if _, err := r.GetCometDClient(ctx, "bad"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}

	q.EXPECT().DisconnectCometDClient(ctx, "c1").Return(nil).Once()
	if err := r.DisconnectCometDClient(ctx, "c1"); err != nil {
		t.Fatal(err)
	}

	q.EXPECT().UpsertCometDSubscription(ctx, mock.Anything).Return(gen.SalesforceCometdSubscription{ID: 1, ClientID: "c1"}, nil).Once()
	if s, err := r.UpsertCometDSubscription(ctx, "c1", "/event/Test__e", -1); err != nil || s.ID != 1 {
		t.Fatal(err)
	}
	q.EXPECT().UpsertCometDSubscription(ctx, mock.Anything).Return(gen.SalesforceCometdSubscription{}, errors.New("err")).Once()
	if _, err := r.UpsertCometDSubscription(ctx, "c1", "/event/Test__e", -1); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().DeleteCometDSubscription(ctx, mock.Anything).Return(nil).Once()
	if err := r.DeleteCometDSubscription(ctx, "c1", "/event/Test__e"); err != nil {
		t.Fatal(err)
	}

	q.EXPECT().ListCometDSubscriptionsByClient(ctx, "c1").Return([]gen.SalesforceCometdSubscription{{ID: 1}}, nil).Once()
	if subs, err := r.ListCometDSubscriptions(ctx, "c1"); err != nil || len(subs) != 1 {
		t.Fatal(err)
	}
	q.EXPECT().ListCometDSubscriptionsByClient(ctx, "c1").Return(nil, errors.New("err")).Once()
	if _, err := r.ListCometDSubscriptions(ctx, "c1"); err == nil {
		t.Fatal("expected err")
	}

	q.EXPECT().UpdateCometDSubscriptionReplayID(ctx, mock.Anything).Return(nil).Once()
	if err := r.UpdateCometDSubscriptionReplayID(ctx, 1, 5); err != nil {
		t.Fatal(err)
	}
}
