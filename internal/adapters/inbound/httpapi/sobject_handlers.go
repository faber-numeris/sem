package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) HandleListVersions(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.svc.ListVersions())
}

func (h *Handler) HandleListResources(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	res, err := h.svc.ListResources(version)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) HandleGetLimits(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	lims, err := h.svc.GetLimits(r.Context(), version)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, lims)
}

func (h *Handler) HandleDescribeGlobal(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	res, err := h.svc.DescribeGlobal(r.Context(), version, false)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) HandleGetSObjectBasicInfo(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	sobject := chi.URLParam(r, "sobject")
	res, err := h.svc.GetSObjectBasicInfo(r.Context(), version, sobject)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) HandleDescribeSObject(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	sobject := chi.URLParam(r, "sobject")
	res, err := h.svc.DescribeSObject(r.Context(), version, sobject, false)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) HandleCreateRecord(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	sobject := chi.URLParam(r, "sobject")
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeSFError(w, err)
		return
	}
	res, _, err := h.svc.CreateRecord(r.Context(), version, sobject, body)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, res)
}

func (h *Handler) HandleGetRecord(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	sobject := chi.URLParam(r, "sobject")
	id := chi.URLParam(r, "id")
	var fields []string
	if fParam := r.URL.Query().Get("fields"); fParam != "" {
		fields = strings.Split(fParam, ",")
	}
	rec, err := h.svc.GetRecord(r.Context(), version, sobject, id, fields)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

func (h *Handler) HandleGetContentVersionData(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	id := chi.URLParam(r, "id")
	data, err := h.svc.GetContentVersionData(r.Context(), version, id)
	if err != nil {
		writeSFError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *Handler) HandleUpdateRecord(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	sobject := chi.URLParam(r, "sobject")
	id := chi.URLParam(r, "id")
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeSFError(w, err)
		return
	}
	if _, err := h.svc.UpdateRecord(r.Context(), version, sobject, id, body); err != nil {
		writeSFError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleUpsertRecord(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	sobject := chi.URLParam(r, "sobject")
	extField := chi.URLParam(r, "externalIdField")
	extVal := chi.URLParam(r, "externalIdValue")
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeSFError(w, err)
		return
	}
	res, err := h.svc.UpsertRecord(r.Context(), version, sobject, extField, extVal, body)
	if err != nil {
		writeSFError(w, err)
		return
	}
	status := http.StatusOK
	if res.Created {
		status = http.StatusCreated
	}
	writeJSON(w, status, res)
}

func (h *Handler) HandleDeleteRecord(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	sobject := chi.URLParam(r, "sobject")
	id := chi.URLParam(r, "id")
	if err := h.svc.DeleteRecord(r.Context(), version, sobject, id); err != nil {
		writeSFError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseBatchSizeHeader(r *http.Request) int {
	hdr := r.Header.Get("Sforce-Query-Options")
	if hdr == "" {
		return 2000
	}
	for _, part := range strings.Split(hdr, ";") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) == 2 && strings.EqualFold(strings.TrimSpace(kv[0]), "batchSize") {
			if n, err := strconv.Atoi(strings.TrimSpace(kv[1])); err == nil && n > 0 {
				return n
			}
		}
	}
	return 2000
}

func (h *Handler) HandleQuery(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	soql := r.URL.Query().Get("q")
	batchSize := parseBatchSizeHeader(r)
	res, err := h.svc.ExecuteSOQL(r.Context(), version, soql, false, batchSize)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) HandleQueryNextBatch(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	cursorID := chi.URLParam(r, "cursorId")
	res, err := h.svc.QueryNextBatch(r.Context(), version, cursorID)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) HandleQueryAll(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	soql := r.URL.Query().Get("q")
	batchSize := parseBatchSizeHeader(r)
	res, err := h.svc.ExecuteSOQL(r.Context(), version, soql, true, batchSize)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	sosl := r.URL.Query().Get("q")
	res, err := h.svc.ExecuteSOSL(r.Context(), version, sosl)
	if err != nil {
		writeSFError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}
