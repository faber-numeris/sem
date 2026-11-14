package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"

	"github.com/faber-numeris/sem/internal/core/domain"
	"github.com/go-chi/chi/v5"
)

func formatIngestJobMap(j *domain.BulkIngestJob) map[string]any {
	m := map[string]any{
		"id":                      j.ID,
		"operation":               j.Operation,
		"object":                  j.ObjectName,
		"createdById":             j.CreatedByID,
		"createdDate":             j.CreatedDate.UTC().Format("2006-01-02T15:04:05.000+0000"),
		"systemModstamp":          j.SystemModstamp.UTC().Format("2006-01-02T15:04:05.000+0000"),
		"state":                   j.State,
		"concurrencyMode":         j.ConcurrencyMode,
		"contentType":             j.ContentType,
		"apiVersion":              j.APIVersion,
		"columnDelimiter":         j.ColumnDelimiter,
		"lineEnding":              j.LineEnding,
		"numberRecordsProcessed":  j.NumberRecordsProcessed,
		"numberRecordsFailed":     j.NumberRecordsFailed,
		"retries":                 j.Retries,
		"totalProcessingTime":     j.TotalProcessingTime,
		"contentUrl":              fmt.Sprintf("services/data/v%s/jobs/ingest/%s/batches", j.APIVersion, j.ID),
	}
	if j.ExternalIDFieldName != "" {
		m["externalIdFieldName"] = j.ExternalIDFieldName
	}
	return m
}

func formatQueryJobMap(j *domain.BulkQueryJob) map[string]any {
	return map[string]any{
		"id":                     j.ID,
		"operation":              j.Operation,
		"object":                 j.ObjectName,
		"createdById":            j.CreatedByID,
		"createdDate":            j.CreatedDate.UTC().Format("2006-01-02T15:04:05.000+0000"),
		"systemModstamp":         j.SystemModstamp.UTC().Format("2006-01-02T15:04:05.000+0000"),
		"state":                  j.State,
		"concurrencyMode":        j.ConcurrencyMode,
		"contentType":            j.ContentType,
		"apiVersion":             j.APIVersion,
		"columnDelimiter":        j.ColumnDelimiter,
		"lineEnding":             j.LineEnding,
		"numberRecordsProcessed": j.NumberRecordsProcessed,
	}
}

func (h *Handler) HandleCreateIngestJob(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	var body struct {
		Operation           string `json:"operation"`
		Object              string `json:"object"`
		ContentType         string `json:"contentType"`
		ExternalIDFieldName string `json:"externalIdFieldName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeSFError(w, err)
		return
	}
	job, err := h.svc.CreateIngestJob(r.Context(), version, body.Operation, body.Object, body.ContentType, body.ExternalIDFieldName)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, formatIngestJobMap(job))
}

func (h *Handler) HandleListIngestJobs(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	jobs, err := h.svc.ListIngestJobs(r.Context(), version)
	if err != nil {
		writeSFError(w, err)
		return
	}
	recs := make([]map[string]any, 0, len(jobs))
	for i := range jobs {
		recs = append(recs, formatIngestJobMap(&jobs[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"done":           true,
		"records":        recs,
		"nextRecordsUrl": nil,
	})
}

func (h *Handler) HandleUploadIngestBatch(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	jobID := chi.URLParam(r, "jobId")
	csvBytes, _ := io.ReadAll(r.Body)
	if err := h.svc.UploadIngestBatch(r.Context(), version, jobID, string(csvBytes)); err != nil {
		writeSFError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) HandleCloseOrAbortIngestJob(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	jobID := chi.URLParam(r, "jobId")
	var body struct {
		State string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeSFError(w, err)
		return
	}
	job, err := h.svc.CloseOrAbortIngestJob(r.Context(), version, jobID, body.State)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, formatIngestJobMap(job))
}

func (h *Handler) HandleGetIngestJob(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	jobID := chi.URLParam(r, "jobId")
	job, err := h.svc.GetIngestJob(r.Context(), version, jobID)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, formatIngestJobMap(job))
}

func (h *Handler) HandleGetIngestSuccessfulResults(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	jobID := chi.URLParam(r, "jobId")
	csvData, err := h.svc.GetIngestResultsCSV(r.Context(), version, jobID, true)
	if err != nil {
		writeSFError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(csvData))
}

func (h *Handler) HandleGetIngestFailedResults(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	jobID := chi.URLParam(r, "jobId")
	csvData, err := h.svc.GetIngestResultsCSV(r.Context(), version, jobID, false)
	if err != nil {
		writeSFError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(csvData))
}

func (h *Handler) HandleCreateQueryJob(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	var body struct {
		Operation string `json:"operation"`
		Query     string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeSFError(w, err)
		return
	}
	job, err := h.svc.CreateQueryJob(r.Context(), version, body.Operation, body.Query)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, formatQueryJobMap(job))
}

func (h *Handler) HandleGetQueryJob(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	jobID := chi.URLParam(r, "jobId")
	job, err := h.svc.GetQueryJob(r.Context(), version, jobID)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, formatQueryJobMap(job))
}

func (h *Handler) HandleGetQueryJobResults(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	jobID := chi.URLParam(r, "jobId")
	maxRecords, _ := strconv.Atoi(r.URL.Query().Get("maxRecords"))
	csvData, count, err := h.svc.GetQueryJobResultsCSV(r.Context(), version, jobID, maxRecords)
	if err != nil {
		writeSFError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=UTF-8")
	w.Header().Set("Sforce-NumberOfRecords", strconv.Itoa(count))
	w.Header().Set("Sforce-Locator", "null")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(csvData))
}

func (h *Handler) HandleToolingQuery(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	soql := r.URL.Query().Get("q")
	res, err := h.svc.ExecuteToolingQuery(r.Context(), version, soql)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) HandleExecuteAnonymousApex(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	body := r.URL.Query().Get("anonymousBody")
	res, err := h.svc.ExecuteAnonymousApex(r.Context(), version, body)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) HandleDescribeToolingGlobal(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	res, err := h.svc.DescribeGlobal(r.Context(), version, true)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) HandleDescribeToolingSObject(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	sobject := chi.URLParam(r, "sobject")
	res, err := h.svc.DescribeToolingSObject(r.Context(), version, sobject)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) HandleRunTestsAsynchronous(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	var body struct {
		ClassNames string `json:"classNames"`
		TestLevel  string `json:"testLevel"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	jobID, err := h.svc.RunTestsAsynchronous(r.Context(), version, body.ClassNames, body.TestLevel)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, jobID)
}

func (h *Handler) HandleApexRest(w http.ResponseWriter, r *http.Request) {
	mapping := strings.TrimPrefix(r.URL.Path, "/services/apexrest")
	bodyBytes, _ := io.ReadAll(r.Body)
	status, respBody, err := h.svc.ExecuteApexRest(r.Context(), mapping, r.Method, string(bodyBytes))
	if err != nil {
		writeSFError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(respBody))
}

func (h *Handler) HandleCometD(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	rawBytes, _ := io.ReadAll(r.Body)
	trimmed := bytes.TrimSpace(rawBytes)
	var msgs []domain.BayeuxMessage
	if len(trimmed) > 0 && trimmed[0] == '[' {
		if err := json.Unmarshal(trimmed, &msgs); err != nil {
			writeSFError(w, err)
			return
		}
	} else {
		var single domain.BayeuxMessage
		if err := json.Unmarshal(trimmed, &single); err != nil {
			writeSFError(w, err)
			return
		}
		msgs = []domain.BayeuxMessage{single}
	}
	resps, err := h.svc.ProcessCometD(r.Context(), version, msgs)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resps)
}

var reRef = regexp.MustCompile(`@\{([a-zA-Z0-9_]+)\.([a-zA-Z0-9_.]+)\}`)

func interpolateRefs(input string, refs map[string]map[string]any) string {
	return reRef.ReplaceAllStringFunc(input, func(match string) string {
		m := reRef.FindStringSubmatch(match)
		refID, field := m[1], m[2]
		if refMap, ok := refs[refID]; ok {
			for k, v := range refMap {
				if strings.EqualFold(k, field) {
					return fmt.Sprintf("%v", v)
				}
			}
		}
		return match
	})
}

func (h *Handler) dispatchSubRequest(r *http.Request, method, rawURL string, body any) (int, map[string]string, any) {
	urlStr := strings.TrimSpace(rawURL)
	if !strings.HasPrefix(urlStr, "/") {
		if strings.HasPrefix(urlStr, "v") {
			urlStr = "/services/data/" + urlStr
		} else {
			urlStr = "/" + urlStr
		}
	}
	var reqBody io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewReader(b)
	}
	subReq, _ := http.NewRequest(strings.ToUpper(method), urlStr, reqBody)
	subReq.Header.Set("Content-Type", "application/json")
	if auth := r.Header.Get("Authorization"); auth != "" {
		subReq.Header.Set("Authorization", auth)
	}
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, subReq)

	hdrs := map[string]string{}
	for k, vals := range rec.Header() {
		if len(vals) > 0 {
			hdrs[k] = vals[0]
		}
	}
	var parsedBody any
	if rec.Body.Len() > 0 {
		_ = json.Unmarshal(rec.Body.Bytes(), &parsedBody)
	}
	return rec.Code, hdrs, parsedBody
}

func (h *Handler) HandleComposite(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	if err := h.svc.ValidateVersion(version); err != nil {
		writeSFError(w, err)
		return
	}
	var req domain.CompositeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeSFError(w, err)
		return
	}

	refs := make(map[string]map[string]any)
	responses := make([]domain.CompositeSubResponse, 0, len(req.CompositeRequest))

	for _, sub := range req.CompositeRequest {
		interpURL := interpolateRefs(sub.URL, refs)
		var interpBody map[string]any
		if sub.Body != nil {
			b, _ := json.Marshal(sub.Body)
			interpolatedJSON := interpolateRefs(string(b), refs)
			_ = json.Unmarshal([]byte(interpolatedJSON), &interpBody)
		}
		code, hdrs, parsed := h.dispatchSubRequest(r, sub.Method, interpURL, interpBody)
		if m, ok := parsed.(map[string]any); ok && sub.ReferenceID != "" {
			refs[sub.ReferenceID] = m
		}
		responses = append(responses, domain.CompositeSubResponse{
			Body:           parsed,
			HTTPHeaders:    hdrs,
			HTTPStatusCode: code,
			ReferenceID:    sub.ReferenceID,
		})
	}

	writeJSON(w, http.StatusOK, domain.CompositeResponse{CompositeResponse: responses})
}

func (h *Handler) HandleCompositeBatch(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	if err := h.svc.ValidateVersion(version); err != nil {
		writeSFError(w, err)
		return
	}
	var req domain.BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeSFError(w, err)
		return
	}

	hasErrors := false
	results := make([]domain.BatchSubResponse, 0, len(req.BatchRequests))
	for _, sub := range req.BatchRequests {
		code, _, parsed := h.dispatchSubRequest(r, sub.Method, sub.URL, sub.RichInput)
		if code >= 400 {
			hasErrors = true
		}
		results = append(results, domain.BatchSubResponse{
			StatusCode: code,
			Result:     parsed,
		})
	}
	writeJSON(w, http.StatusOK, domain.BatchResponse{
		HasErrors: hasErrors,
		Results:   results,
	})
}
