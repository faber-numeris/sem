package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/faber-numeris/sem/internal/core/domain"
	"github.com/faber-numeris/sem/internal/mocks"
	"github.com/stretchr/testify/mock"
)

func TestHandler_HealthReadyAndMiddleware(t *testing.T) {
	svc := mocks.NewMockEmulatorService(t)
	r := NewRouter(svc, func() bool { return true })

	// Trailing slash + RawPath + root "/"
	svc.EXPECT().ListVersions().Return([]domain.APIVersion{{Version: "61.0"}}).Once()
	req := httptest.NewRequest(http.MethodGet, "/services/data/", nil)
	req.URL.RawPath = "/services/data/"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Only slashes path "///" -> "/"
	reqSlash := httptest.NewRequest(http.MethodGet, "///", nil)
	wSlash := httptest.NewRecorder()
	r.ServeHTTP(wSlash, reqSlash)

	// Health & Ready (up & down)
	reqH := httptest.NewRequest(http.MethodGet, "/health", nil)
	wH := httptest.NewRecorder()
	r.ServeHTTP(wH, reqH)
	if wH.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", wH.Code)
	}
	reqR := httptest.NewRequest(http.MethodGet, "/ready", nil)
	wR := httptest.NewRecorder()
	r.ServeHTTP(wR, reqR)
	if wR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", wR.Code)
	}

	rDown := NewRouter(svc, func() bool { return false })
	wDown := httptest.NewRecorder()
	rDown.ServeHTTP(wDown, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if wDown.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", wDown.Code)
	}
}

func TestHandler_AuthEndpoints(t *testing.T) {
	svc := mocks.NewMockEmulatorService(t)
	r := NewRouter(svc, nil)

	// Token: form-urlencoded with BasicAuth & X-Forwarded-Proto
	svc.EXPECT().IssueToken(mock.Anything, mock.Anything).Return(&domain.OAuthToken{
		AccessToken:  "at1",
		RefreshToken: "rt1",
		InstanceURL:  "https://example.com",
		IDURL:        "https://example.com/id/1/1",
		TokenType:    "Bearer",
	}, nil).Once()
	req := httptest.NewRequest(http.MethodPost, "/services/oauth2/token", strings.NewReader("grant_type=client_credentials"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Forwarded-Proto", "https")
	req.SetBasicAuth("cid", "csec")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Token: JSON body & error
	svc.EXPECT().IssueToken(mock.Anything, mock.Anything).Return(nil, domain.ErrInvalidGrant).Once()
	reqJSON := httptest.NewRequest(http.MethodPost, "/services/oauth2/token", strings.NewReader(`{"grant_type":"bad"}`))
	reqJSON.Header.Set("Content-Type", "application/json")
	wJSON := httptest.NewRecorder()
	r.ServeHTTP(wJSON, reqJSON)
	if wJSON.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", wJSON.Code)
	}

	// Authorize: code with existing query param & state, token implicit flow, and error
	svc.EXPECT().Authorize(mock.Anything, mock.Anything).Return(&domain.AuthorizeResult{
		Code:        "c1",
		RedirectURI: "http://localhost/cb?foo=1",
		State:       "xyz",
	}, nil).Once()
	wAuth := httptest.NewRecorder()
	r.ServeHTTP(wAuth, httptest.NewRequest(http.MethodGet, "/services/oauth2/authorize?response_type=code&client_id=cid&state=xyz", nil))
	if wAuth.Code != http.StatusFound || !strings.Contains(wAuth.Header().Get("Location"), "&code=c1&state=xyz") {
		t.Fatalf("unexpected authorize response: %d %s", wAuth.Code, wAuth.Header().Get("Location"))
	}

	svc.EXPECT().Authorize(mock.Anything, mock.Anything).Return(&domain.AuthorizeResult{
		AccessToken: "at_imp",
		RedirectURI: "http://localhost/cb",
	}, nil).Once()
	wAuthTok := httptest.NewRecorder()
	r.ServeHTTP(wAuthTok, httptest.NewRequest(http.MethodGet, "/services/oauth2/authorize?response_type=token&client_id=cid", nil))
	if wAuthTok.Code != http.StatusFound || !strings.Contains(wAuthTok.Header().Get("Location"), "#access_token=at_imp") {
		t.Fatalf("unexpected authorize token response: %d", wAuthTok.Code)
	}

	svc.EXPECT().Authorize(mock.Anything, mock.Anything).Return(nil, domain.ErrInvalidRequest).Once()
	wAuthErr := httptest.NewRecorder()
	r.ServeHTTP(wAuthErr, httptest.NewRequest(http.MethodGet, "/services/oauth2/authorize", nil))
	if wAuthErr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", wAuthErr.Code)
	}

	// Revoke: form body success & error
	svc.EXPECT().RevokeToken(mock.Anything, "tok1").Return(nil).Once()
	reqRev := httptest.NewRequest(http.MethodPost, "/services/oauth2/revoke", strings.NewReader("token=tok1"))
	reqRev.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	wRev := httptest.NewRecorder()
	r.ServeHTTP(wRev, reqRev)
	if wRev.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", wRev.Code)
	}

	svc.EXPECT().RevokeToken(mock.Anything, "").Return(domain.ErrInvalidToken).Once()
	wRevErr := httptest.NewRecorder()
	r.ServeHTTP(wRevErr, httptest.NewRequest(http.MethodPost, "/services/oauth2/revoke", nil))
	if wRevErr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", wRevErr.Code)
	}

	// Introspect
	svc.EXPECT().IntrospectToken(mock.Anything, "tok1").Return(&domain.IntrospectResult{Active: true}, nil).Once()
	wInt := httptest.NewRecorder()
	r.ServeHTTP(wInt, httptest.NewRequest(http.MethodPost, "/services/oauth2/introspect?token=tok1", nil))
	if wInt.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", wInt.Code)
	}

	// UserInfo: ok & error
	svc.EXPECT().GetUserInfo(mock.Anything, "Bearer tok1", mock.Anything).Return(&domain.UserInfoResult{UserID: "005A"}, nil).Once()
	reqUI := httptest.NewRequest(http.MethodGet, "/services/oauth2/userinfo", nil)
	reqUI.Header.Set("Authorization", "Bearer tok1")
	wUI := httptest.NewRecorder()
	r.ServeHTTP(wUI, reqUI)
	if wUI.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", wUI.Code)
	}

	svc.EXPECT().GetUserInfo(mock.Anything, "", mock.Anything).Return(nil, domain.ErrInvalidToken).Once()
	wUIErr := httptest.NewRecorder()
	r.ServeHTTP(wUIErr, httptest.NewRequest(http.MethodGet, "/services/oauth2/userinfo", nil))
	if wUIErr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", wUIErr.Code)
	}
}

func TestHandler_SObjectAndQueryEndpoints(t *testing.T) {
	svc := mocks.NewMockEmulatorService(t)
	r := NewRouter(svc, nil)

	// ListResources (ok & err)
	svc.EXPECT().ListResources("61.0").Return(map[string]string{"sobjects": "/services/data/v61.0/sobjects"}, nil).Once()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	svc.EXPECT().ListResources("1.0").Return(nil, domain.ErrInvalidVersion).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v1.0", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}

	// Limits (ok & err)
	svc.EXPECT().GetLimits(mock.Anything, "61.0").Return(map[string]domain.OrgLimit{}, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/limits", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	svc.EXPECT().GetLimits(mock.Anything, "1.0").Return(nil, domain.ErrInvalidVersion).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v1.0/limits", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}

	// DescribeGlobal (ok & err)
	svc.EXPECT().DescribeGlobal(mock.Anything, "61.0", false).Return(map[string]any{"encoding": "UTF-8"}, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/sobjects/", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	svc.EXPECT().DescribeGlobal(mock.Anything, "1.0", false).Return(nil, domain.ErrInvalidVersion).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v1.0/sobjects", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}

	// GetSObjectBasicInfo (ok & err)
	svc.EXPECT().GetSObjectBasicInfo(mock.Anything, "61.0", "Account").Return(map[string]any{}, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/sobjects/Account", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	svc.EXPECT().GetSObjectBasicInfo(mock.Anything, "61.0", "Bad").Return(nil, domain.ErrNotFound).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/sobjects/Bad", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}

	// DescribeSObject (ok & err)
	svc.EXPECT().DescribeSObject(mock.Anything, "61.0", "Account", false).Return(map[string]any{}, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/sobjects/Account/describe", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	svc.EXPECT().DescribeSObject(mock.Anything, "61.0", "Bad", false).Return(nil, domain.ErrNotFound).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/sobjects/Bad/describe", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}

	// CreateRecord (bad json, service err, ok)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/services/data/v61.0/sobjects/Account/", strings.NewReader("{bad")))
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}
	svc.EXPECT().CreateRecord(mock.Anything, "61.0", "Account", mock.Anything).Return(nil, nil, domain.ErrRequiredFieldMissing).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/services/data/v61.0/sobjects/Account", strings.NewReader(`{}`)))
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}
	svc.EXPECT().CreateRecord(mock.Anything, "61.0", "Account", mock.Anything).Return(&domain.SaveResult{ID: "001A", Success: true}, nil, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/services/data/v61.0/sobjects/Account", strings.NewReader(`{"Name":"Acme"}`)))
	if w.Code != http.StatusCreated {
		t.Fatal(w.Code)
	}

	// GetRecord (ok with fields & err)
	svc.EXPECT().GetRecord(mock.Anything, "61.0", "Account", "001A", []string{"Id", "Name"}).Return(map[string]any{"Id": "001A"}, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/sobjects/Account/001A?fields=Id,Name", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	svc.EXPECT().GetRecord(mock.Anything, "61.0", "Account", "bad", []string(nil)).Return(nil, domain.ErrNotFound).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/sobjects/Account/bad", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}

	// GetContentVersionData (ok & err)
	svc.EXPECT().GetContentVersionData(mock.Anything, "61.0", "068A").Return([]byte("data"), nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/sobjects/ContentVersion/068A/VersionData", nil))
	if w.Code != http.StatusOK || w.Body.String() != "data" {
		t.Fatal(w.Code)
	}
	svc.EXPECT().GetContentVersionData(mock.Anything, "61.0", "bad").Return(nil, domain.ErrNotFound).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/sobjects/ContentVersion/bad/VersionData", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}

	// UpdateRecord (bad json, err, ok)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/services/data/v61.0/sobjects/Account/001A", strings.NewReader("{bad")))
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}
	svc.EXPECT().UpdateRecord(mock.Anything, "61.0", "Account", "bad", mock.Anything).Return(nil, domain.ErrNotFound).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/services/data/v61.0/sobjects/Account/bad", strings.NewReader(`{}`)))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}
	svc.EXPECT().UpdateRecord(mock.Anything, "61.0", "Account", "001A", mock.Anything).Return(map[string]any{}, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/services/data/v61.0/sobjects/Account/001A", strings.NewReader(`{"Name":"New"}`)))
	if w.Code != http.StatusNoContent {
		t.Fatal(w.Code)
	}

	// UpsertRecord (bad json, err, created=true, created=false)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/services/data/v61.0/sobjects/Account/External_ID__c/E1", strings.NewReader("{bad")))
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}
	svc.EXPECT().UpsertRecord(mock.Anything, "61.0", "Account", "External_ID__c", "E1", mock.Anything).Return(nil, domain.ErrInvalidRequest).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/services/data/v61.0/sobjects/Account/External_ID__c/E1", strings.NewReader(`{}`)))
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}
	svc.EXPECT().UpsertRecord(mock.Anything, "61.0", "Account", "External_ID__c", "E1", mock.Anything).Return(&domain.SaveResult{ID: "001A", Created: true}, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/services/data/v61.0/sobjects/Account/External_ID__c/E1", strings.NewReader(`{"Name":"N"}`)))
	if w.Code != http.StatusCreated {
		t.Fatal(w.Code)
	}
	svc.EXPECT().UpsertRecord(mock.Anything, "61.0", "Account", "External_ID__c", "E1", mock.Anything).Return(&domain.SaveResult{ID: "001A", Created: false}, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/services/data/v61.0/sobjects/Account/External_ID__c/E1", strings.NewReader(`{"Name":"N"}`)))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}

	// DeleteRecord (ok & err)
	svc.EXPECT().DeleteRecord(mock.Anything, "61.0", "Account", "001A").Return(nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/services/data/v61.0/sobjects/Account/001A", nil))
	if w.Code != http.StatusNoContent {
		t.Fatal(w.Code)
	}
	svc.EXPECT().DeleteRecord(mock.Anything, "61.0", "Account", "bad").Return(domain.ErrNotFound).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/services/data/v61.0/sobjects/Account/bad", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}

	// Query, QueryNextBatch, QueryAll, Search
	svc.EXPECT().ExecuteSOQL(mock.Anything, "61.0", "SELECT Id FROM Account", false, 50).Return(&domain.QueryResult{TotalSize: 1, Done: true}, nil).Once()
	reqQ := httptest.NewRequest(http.MethodGet, "/services/data/v61.0/query/?q=SELECT+Id+FROM+Account", nil)
	reqQ.Header.Set("Sforce-Query-Options", "batchSize=50")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, reqQ)
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}

	svc.EXPECT().ExecuteSOQL(mock.Anything, "61.0", "BAD", false, 2000).Return(nil, domain.ErrMalformedQuery).Once()
	reqQBad := httptest.NewRequest(http.MethodGet, "/services/data/v61.0/query?q=BAD", nil)
	reqQBad.Header.Set("Sforce-Query-Options", "other=1; batchSize=invalid")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, reqQBad)
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}

	svc.EXPECT().QueryNextBatch(mock.Anything, "61.0", "01g123-50").Return(&domain.QueryResult{Done: true}, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/query/01g123-50", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	svc.EXPECT().QueryNextBatch(mock.Anything, "61.0", "bad").Return(nil, domain.ErrNotFound).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/query/bad", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}

	svc.EXPECT().ExecuteSOQL(mock.Anything, "61.0", "SELECT Id FROM Account", true, 2000).Return(&domain.QueryResult{Done: true}, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/queryAll/?q=SELECT+Id+FROM+Account", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	svc.EXPECT().ExecuteSOQL(mock.Anything, "61.0", "BAD", true, 2000).Return(nil, domain.ErrMalformedQuery).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/queryAll?q=BAD", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}

	svc.EXPECT().ExecuteSOSL(mock.Anything, "61.0", "FIND {Acme}").Return(&domain.SearchResult{}, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/search/?q=FIND+{Acme}", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	svc.EXPECT().ExecuteSOSL(mock.Anything, "61.0", "BAD").Return(nil, domain.ErrMalformedSearch).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/search?q=BAD", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}
}

func TestHandler_BulkToolingApexCometDComposite(t *testing.T) {
	svc := mocks.NewMockEmulatorService(t)
	r := NewRouter(svc, nil)

	// Bulk Ingest: Create (bad json, err, ok with externalIdFieldName)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/services/data/v61.0/jobs/ingest/", strings.NewReader("{bad")))
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}
	svc.EXPECT().CreateIngestJob(mock.Anything, "61.0", "bad", "", "", "").Return(nil, domain.ErrInvalidRequest).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/services/data/v61.0/jobs/ingest", strings.NewReader(`{"operation":"bad"}`)))
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}
	svc.EXPECT().CreateIngestJob(mock.Anything, "61.0", "upsert", "Account", "CSV", "External_ID__c").Return(&domain.BulkIngestJob{ID: "750A", ExternalIDFieldName: "External_ID__c"}, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/services/data/v61.0/jobs/ingest", strings.NewReader(`{"operation":"upsert","object":"Account","contentType":"CSV","externalIdFieldName":"External_ID__c"}`)))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}

	// ListIngestJobs (ok & err)
	svc.EXPECT().ListIngestJobs(mock.Anything, "61.0").Return([]domain.BulkIngestJob{{ID: "750A"}}, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/jobs/ingest", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	svc.EXPECT().ListIngestJobs(mock.Anything, "1.0").Return(nil, domain.ErrInvalidVersion).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v1.0/jobs/ingest", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}

	// UploadIngestBatch (ok & err)
	svc.EXPECT().UploadIngestBatch(mock.Anything, "61.0", "750A", "Name\nAcme").Return(nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/services/data/v61.0/jobs/ingest/750A/batches", strings.NewReader("Name\nAcme")))
	if w.Code != http.StatusCreated {
		t.Fatal(w.Code)
	}
	svc.EXPECT().UploadIngestBatch(mock.Anything, "61.0", "bad", "").Return(domain.ErrNotFound).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/services/data/v61.0/jobs/ingest/bad/batches", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}

	// CloseOrAbortIngestJob (bad json, err, ok)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/services/data/v61.0/jobs/ingest/750A/", strings.NewReader("{bad")))
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}
	svc.EXPECT().CloseOrAbortIngestJob(mock.Anything, "61.0", "750A", "Bad").Return(nil, domain.ErrInvalidJobState).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/services/data/v61.0/jobs/ingest/750A", strings.NewReader(`{"state":"Bad"}`)))
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}
	svc.EXPECT().CloseOrAbortIngestJob(mock.Anything, "61.0", "750A", "UploadComplete").Return(&domain.BulkIngestJob{ID: "750A", State: "JobComplete"}, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/services/data/v61.0/jobs/ingest/750A", strings.NewReader(`{"state":"UploadComplete"}`)))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}

	// GetIngestJob (ok & err)
	svc.EXPECT().GetIngestJob(mock.Anything, "61.0", "750A").Return(&domain.BulkIngestJob{ID: "750A"}, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/jobs/ingest/750A/", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	svc.EXPECT().GetIngestJob(mock.Anything, "61.0", "bad").Return(nil, domain.ErrNotFound).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/jobs/ingest/bad", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}

	// SuccessfulResults & FailedResults (ok & err)
	svc.EXPECT().GetIngestResultsCSV(mock.Anything, "61.0", "750A", true).Return("sf__Id\n001A\n", nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/jobs/ingest/750A/successfulResults/", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	svc.EXPECT().GetIngestResultsCSV(mock.Anything, "61.0", "bad", true).Return("", domain.ErrNotFound).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/jobs/ingest/bad/successfulResults", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}
	svc.EXPECT().GetIngestResultsCSV(mock.Anything, "61.0", "750A", false).Return("sf__Id,sf__Error\n", nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/jobs/ingest/750A/failedResults/", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	svc.EXPECT().GetIngestResultsCSV(mock.Anything, "61.0", "bad", false).Return("", domain.ErrNotFound).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/jobs/ingest/bad/failedResults", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}

	// Bulk Query: Create (bad json, err, ok), Get (ok, err), Results (ok, err)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/services/data/v61.0/jobs/query/", strings.NewReader("{bad")))
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}
	svc.EXPECT().CreateQueryJob(mock.Anything, "61.0", "query", "BAD").Return(nil, domain.ErrMalformedQuery).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/services/data/v61.0/jobs/query", strings.NewReader(`{"operation":"query","query":"BAD"}`)))
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}
	svc.EXPECT().CreateQueryJob(mock.Anything, "61.0", "query", "SELECT Id FROM Contact").Return(&domain.BulkQueryJob{ID: "750Q"}, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/services/data/v61.0/jobs/query", strings.NewReader(`{"operation":"query","query":"SELECT Id FROM Contact"}`)))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}

	svc.EXPECT().GetQueryJob(mock.Anything, "61.0", "750Q").Return(&domain.BulkQueryJob{ID: "750Q"}, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/jobs/query/750Q", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	svc.EXPECT().GetQueryJob(mock.Anything, "61.0", "bad").Return(nil, domain.ErrNotFound).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/jobs/query/bad", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}

	svc.EXPECT().GetQueryJobResultsCSV(mock.Anything, "61.0", "750Q", 10000).Return("Id\n003A\n", 1, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/jobs/query/750Q/results?maxRecords=10000", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	svc.EXPECT().GetQueryJobResultsCSV(mock.Anything, "61.0", "bad", 0).Return("", 0, domain.ErrNotFound).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/jobs/query/bad/results", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}

	// Tooling endpoints
	svc.EXPECT().ExecuteToolingQuery(mock.Anything, "61.0", "SELECT Id FROM ApexClass").Return(&domain.QueryResult{Done: true}, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/tooling/query/?q=SELECT+Id+FROM+ApexClass", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	svc.EXPECT().ExecuteToolingQuery(mock.Anything, "61.0", "BAD").Return(nil, domain.ErrMalformedQuery).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/tooling/query?q=BAD", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}

	svc.EXPECT().ExecuteAnonymousApex(mock.Anything, "61.0", "System.debug(1);").Return(&domain.AnonymousApexResult{Compiled: true, Success: true}, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/tooling/executeAnonymous/?anonymousBody=System.debug(1);", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	svc.EXPECT().ExecuteAnonymousApex(mock.Anything, "1.0", "").Return(nil, domain.ErrInvalidVersion).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v1.0/tooling/executeAnonymous", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}

	svc.EXPECT().DescribeGlobal(mock.Anything, "61.0", true).Return(map[string]any{}, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/tooling/sobjects", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	svc.EXPECT().DescribeGlobal(mock.Anything, "1.0", true).Return(nil, domain.ErrInvalidVersion).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v1.0/tooling/sobjects", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}

	svc.EXPECT().DescribeToolingSObject(mock.Anything, "61.0", "ApexClass").Return(map[string]any{}, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/tooling/sobjects/ApexClass/describe", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	svc.EXPECT().DescribeToolingSObject(mock.Anything, "61.0", "Bad").Return(nil, domain.ErrNotFound).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/data/v61.0/tooling/sobjects/Bad/describe", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}

	svc.EXPECT().RunTestsAsynchronous(mock.Anything, "61.0", "AccountServiceTest", "RunSpecifiedTests").Return("707A", nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/services/data/v61.0/tooling/runTestsAsynchronous/", strings.NewReader(`{"classNames":"AccountServiceTest","testLevel":"RunSpecifiedTests"}`)))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	svc.EXPECT().RunTestsAsynchronous(mock.Anything, "1.0", "", "").Return("", domain.ErrInvalidVersion).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/services/data/v1.0/tooling/runTestsAsynchronous", strings.NewReader(`{}`)))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}

	// Apex REST (ok & err)
	svc.EXPECT().ExecuteApexRest(mock.Anything, "/v1/integrations/status", "GET", "").Return(200, `{"status":"UP"}`, nil).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/apexrest/v1/integrations/status", nil))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	svc.EXPECT().ExecuteApexRest(mock.Anything, "/missing", "GET", "").Return(0, "", domain.ErrNotFound).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/services/apexrest/missing", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}

	// CometD (array ok, array bad json, single object ok, single bad json, service err)
	svc.EXPECT().ProcessCometD(mock.Anything, "61.0", mock.Anything).Return([]domain.BayeuxMessage{{Channel: "/meta/handshake"}}, nil).Twice()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/cometd/61.0/", strings.NewReader(`[{"channel":"/meta/handshake"}]`)))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/cometd/61.0", strings.NewReader(`{"channel":"/meta/handshake"}`)))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/cometd/61.0", strings.NewReader(`[bad`)))
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/cometd/61.0", strings.NewReader(`{bad`)))
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}
	svc.EXPECT().ProcessCometD(mock.Anything, "1.0", mock.Anything).Return(nil, domain.ErrInvalidVersion).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/cometd/1.0", strings.NewReader(`[{"channel":"/meta/handshake"}]`)))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}

	// Composite: invalid version, bad json, and sequential reference resolution (@{newAccount.id} + unresolved @{unknown.id})
	svc.EXPECT().ValidateVersion("1.0").Return(domain.ErrInvalidVersion).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/services/data/v1.0/composite", strings.NewReader(`{}`)))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}
	svc.EXPECT().ValidateVersion("61.0").Return(nil).Twice()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/services/data/v61.0/composite", strings.NewReader(`{bad`)))
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}

	svc.EXPECT().CreateRecord(mock.Anything, "61.0", "Account", map[string]any{"Name": "Parent Corp"}).Return(&domain.SaveResult{ID: "001P", Success: true}, nil, nil).Once()
	svc.EXPECT().CreateRecord(mock.Anything, "61.0", "Contact", map[string]any{"LastName": "Smith", "AccountId": "001P", "Note": "@{unknown.id}"}).Return(&domain.SaveResult{ID: "003S", Success: true}, nil, nil).Once()
	compBody := `{
		"compositeRequest": [
			{
				"method": "POST",
				"url": "/services/data/v61.0/sobjects/Account",
				"referenceId": "newAccount",
				"body": {"Name": "Parent Corp"}
			},
			{
				"method": "POST",
				"url": "services/data/v61.0/sobjects/Contact",
				"referenceId": "newContact",
				"body": {"LastName": "Smith", "AccountId": "@{newAccount.id}", "Note": "@{unknown.id}"}
			}
		]
	}`
	reqComp := httptest.NewRequest(http.MethodPost, "/services/data/v61.0/composite", bytes.NewBufferString(compBody))
	reqComp.Header.Set("Authorization", "Bearer tok")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, reqComp)
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}

	// Composite Batch: invalid version, bad json, ok + sub-error
	svc.EXPECT().ValidateVersion("1.0").Return(domain.ErrInvalidVersion).Once()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/services/data/v1.0/composite/batch", strings.NewReader(`{}`)))
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}
	svc.EXPECT().ValidateVersion("61.0").Return(nil).Twice()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/services/data/v61.0/composite/batch", strings.NewReader(`{bad`)))
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}

	svc.EXPECT().DescribeSObject(mock.Anything, "61.0", "Account", false).Return(map[string]any{"name": "Account"}, nil).Once()
	svc.EXPECT().ExecuteSOQL(mock.Anything, "61.0", "BAD", false, 2000).Return(nil, domain.ErrMalformedQuery).Once()
	batchBody := `{
		"batchRequests": [
			{"method": "GET", "url": "v61.0/sobjects/Account/describe"},
			{"method": "GET", "url": "v61.0/query/?q=BAD"}
		]
	}`
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/services/data/v61.0/composite/batch", strings.NewReader(batchBody)))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"hasErrors":true`) {
		t.Fatalf("unexpected batch response: %d %s", w.Code, w.Body.String())
	}
}
