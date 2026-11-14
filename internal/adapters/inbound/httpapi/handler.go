package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/faber-numeris/sem/api"
	"github.com/faber-numeris/sem/internal/core/domain"
	"github.com/faber-numeris/sem/internal/ports/inbound"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	api.Unimplemented
	svc    inbound.EmulatorService
	pingFn func() bool
	router http.Handler
}

var _ api.ServerInterface = (*Handler)(nil)

func NewHandler(svc inbound.EmulatorService, pingFn func() bool) *Handler {
	return &Handler{
		svc:    svc,
		pingFn: pingFn,
	}
}

func StripTrailingSlashMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) > 1 && strings.HasSuffix(r.URL.Path, "/") {
			r.URL.Path = strings.TrimRight(r.URL.Path, "/")
			if r.URL.Path == "" {
				r.URL.Path = "/"
			}
			if r.URL.RawPath != "" {
				r.URL.RawPath = strings.TrimRight(r.URL.RawPath, "/")
			}
		}
		if strings.Contains(r.URL.RawQuery, ";") {
			r.URL.RawQuery = strings.ReplaceAll(r.URL.RawQuery, ";", "%3B")
		}
		next.ServeHTTP(w, r)
	})
}

func NewRouter(svc inbound.EmulatorService, pingFn func() bool) chi.Router {
	r := chi.NewRouter()
	r.Use(StripTrailingSlashMiddleware)

	h := NewHandler(svc, pingFn)
	h.RegisterRoutes(r)
	h.router = r
	return r
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	// Health & Readiness
	r.Get("/health", h.HandleHealth)
	r.Get("/ready", h.HandleReady)

	// 1. OAuth 2.0 & Authentication Endpoints
	r.Post("/services/oauth2/token", h.HandleToken)
	r.Get("/services/oauth2/authorize", h.HandleAuthorize)
	r.Post("/services/oauth2/revoke", h.HandleRevoke)
	r.Post("/services/oauth2/introspect", h.HandleIntrospect)
	r.Get("/services/oauth2/userinfo", h.HandleUserInfo)
	r.Post("/services/oauth2/userinfo", h.HandleUserInfo)

	// 2. REST API Core Endpoints (CRUD & Metadata)
	r.Get("/services/data", h.HandleListVersions)
	r.Get("/services/data/v{version}", h.HandleListResources)
	r.Get("/services/data/v{version}/limits", h.HandleGetLimits)
	r.Get("/services/data/v{version}/sobjects", h.HandleDescribeGlobal)
	r.Get("/services/data/v{version}/sobjects/{sobject}", h.HandleGetSObjectBasicInfo)
	r.Post("/services/data/v{version}/sobjects/{sobject}", h.HandleCreateRecord)
	r.Get("/services/data/v{version}/sobjects/{sobject}/describe", h.HandleDescribeSObject)
	r.Get("/services/data/v{version}/sobjects/ContentVersion/{id}/VersionData", h.HandleGetContentVersionData)
	r.Get("/services/data/v{version}/sobjects/{sobject}/{id}", h.HandleGetRecord)
	r.Patch("/services/data/v{version}/sobjects/{sobject}/{id}", h.HandleUpdateRecord)
	r.Patch("/services/data/v{version}/sobjects/{sobject}/{externalIdField}/{externalIdValue}", h.HandleUpsertRecord)
	r.Delete("/services/data/v{version}/sobjects/{sobject}/{id}", h.HandleDeleteRecord)

	// 3. Query & Search (SOQL / SOSL)
	r.Get("/services/data/v{version}/query", h.HandleQuery)
	r.Get("/services/data/v{version}/query/{cursorId}", h.HandleQueryNextBatch)
	r.Get("/services/data/v{version}/queryAll", h.HandleQueryAll)
	r.Get("/services/data/v{version}/search", h.HandleSearch)

	// 4. Bulk API 2.0
	r.Post("/services/data/v{version}/jobs/ingest", h.HandleCreateIngestJob)
	r.Get("/services/data/v{version}/jobs/ingest", h.HandleListIngestJobs)
	r.Put("/services/data/v{version}/jobs/ingest/{jobId}/batches", h.HandleUploadIngestBatch)
	r.Patch("/services/data/v{version}/jobs/ingest/{jobId}", h.HandleCloseOrAbortIngestJob)
	r.Get("/services/data/v{version}/jobs/ingest/{jobId}", h.HandleGetIngestJob)
	r.Get("/services/data/v{version}/jobs/ingest/{jobId}/successfulResults", h.HandleGetIngestSuccessfulResults)
	r.Get("/services/data/v{version}/jobs/ingest/{jobId}/failedResults", h.HandleGetIngestFailedResults)
	r.Post("/services/data/v{version}/jobs/query", h.HandleCreateQueryJob)
	r.Get("/services/data/v{version}/jobs/query/{jobId}", h.HandleGetQueryJob)
	r.Get("/services/data/v{version}/jobs/query/{jobId}/results", h.HandleGetQueryJobResults)

	// 5. Tooling API
	r.Get("/services/data/v{version}/tooling/query", h.HandleToolingQuery)
	r.Get("/services/data/v{version}/tooling/executeAnonymous", h.HandleExecuteAnonymousApex)
	r.Get("/services/data/v{version}/tooling/sobjects", h.HandleDescribeToolingGlobal)
	r.Get("/services/data/v{version}/tooling/sobjects/{sobject}/describe", h.HandleDescribeToolingSObject)
	r.Post("/services/data/v{version}/tooling/runTestsAsynchronous", h.HandleRunTestsAsynchronous)

	// 6. Apex REST API & Event Bus (CometD)
	r.HandleFunc("/services/apexrest/*", h.HandleApexRest)
	r.Post("/cometd/{version}", h.HandleCometD)

	// 7. Composite & Batch API
	r.Post("/services/data/v{version}/composite", h.HandleComposite)
	r.Post("/services/data/v{version}/composite/batch", h.HandleCompositeBatch)
}

func (h *Handler) HandleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "UP"})
}

func (h *Handler) HandleReady(w http.ResponseWriter, _ *http.Request) {
	if h.pingFn != nil && !h.pingFn() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "DOWN"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "READY"})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeSFError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	errCode := "INVALID_REQUEST"

	switch {
	case errors.Is(err, domain.ErrNotFound), errors.Is(err, domain.ErrInvalidVersion), errors.Is(err, domain.ErrUnsupportedSObject):
		status = http.StatusNotFound
		errCode = "NOT_FOUND"
	case errors.Is(err, domain.ErrRequiredFieldMissing):
		status = http.StatusBadRequest
		errCode = "REQUIRED_FIELD_MISSING"
	case errors.Is(err, domain.ErrMalformedQuery):
		status = http.StatusBadRequest
		errCode = "MALFORMED_QUERY"
	case errors.Is(err, domain.ErrMalformedSearch):
		status = http.StatusBadRequest
		errCode = "MALFORMED_SEARCH"
	case errors.Is(err, domain.ErrInvalidToken), errors.Is(err, domain.ErrInvalidClient), errors.Is(err, domain.ErrInvalidGrant):
		status = http.StatusUnauthorized
		errCode = "INVALID_SESSION_ID"
	}

	writeJSON(w, status, []map[string]any{
		{
			"message":   err.Error(),
			"errorCode": errCode,
			"fields":    []string{},
		},
	})
}
