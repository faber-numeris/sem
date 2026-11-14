package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"strings"

	"github.com/faber-numeris/sem/internal/core/domain"
)

func (s *Service) CreateIngestJob(ctx context.Context, version, operation, objectName, contentType, externalIDField string) (*domain.BulkIngestJob, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, err
	}
	op := strings.ToLower(strings.TrimSpace(operation))
	switch op {
	case "insert", "update", "upsert", "delete", "harddelete":
	default:
		return nil, domain.ErrInvalidRequest
	}
	if strings.TrimSpace(objectName) == "" {
		return nil, domain.ErrInvalidRequest
	}
	if contentType == "" {
		contentType = "CSV"
	}
	job := &domain.BulkIngestJob{
		Operation:           op,
		ObjectName:          objectName,
		ExternalIDFieldName: externalIDField,
		ContentType:         contentType,
		ColumnDelimiter:     "COMMA",
		LineEnding:          "LF",
		APIVersion:          normalizeVersion(version),
	}
	return s.repo.CreateBulkIngestJob(ctx, job)
}

func (s *Service) UploadIngestBatch(ctx context.Context, version, jobID, csvData string) error {
	if err := s.ValidateVersion(version); err != nil {
		return err
	}
	job, err := s.repo.GetBulkIngestJob(ctx, jobID)
	if err != nil {
		return err
	}
	if job.State != "Open" {
		return domain.ErrInvalidJobState
	}
	_, err = s.repo.CreateBulkIngestBatch(ctx, jobID, csvData)
	return err
}

func encodeCSVRow(cols []string) string {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write(cols)
	w.Flush()
	return strings.TrimSuffix(buf.String(), "\n")
}

func (s *Service) CloseOrAbortIngestJob(ctx context.Context, version, jobID, state string) (*domain.BulkIngestJob, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, err
	}
	job, err := s.repo.GetBulkIngestJob(ctx, jobID)
	if err != nil {
		return nil, err
	}

	switch state {
	case "Aborted":
		return s.repo.UpdateBulkIngestJobState(ctx, jobID, "Aborted")
	case "UploadComplete":
		// Process all uploaded batches synchronously so the job is immediately ready
		batches, err := s.repo.ListBulkIngestBatches(ctx, jobID)
		if err != nil {
			return nil, err
		}
		processed := 0
		failed := 0
		for _, batch := range batches {
			r := csv.NewReader(strings.NewReader(strings.TrimSpace(batch.CSVContent)))
			rows, err := r.ReadAll()
			if err != nil || len(rows) < 2 {
				continue
			}
			headers := rows[0]
			for i := range headers {
				headers[i] = strings.TrimSpace(headers[i])
			}
			for _, row := range rows[1:] {
				processed++
				fieldMap := make(map[string]any, len(headers))
				for idx, h := range headers {
					if idx < len(row) {
						fieldMap[h] = row[idx]
					}
				}
				rowCSV := encodeCSVRow(row)
				var sfID string
				var sfCreated bool
				var opErr error

				switch job.Operation {
				case "insert":
					res, _, err := s.repo.CreateRecord(ctx, job.ObjectName, fieldMap)
					if err != nil {
						opErr = err
					} else {
						sfID = res.ID
						sfCreated = true
					}
				case "update":
					idVal := stringifyFieldVal(getFieldIgnoreCase(fieldMap, "Id"))
					if idVal == "" {
						opErr = domain.ErrRequiredFieldMissing
					} else {
						_, err := s.repo.UpdateRecord(ctx, job.ObjectName, idVal, fieldMap)
						if err != nil {
							opErr = err
						} else {
							sfID = idVal
						}
					}
				case "upsert":
					extField := job.ExternalIDFieldName
					if extField == "" {
						extField = "External_ID__c"
					}
					extVal := stringifyFieldVal(getFieldIgnoreCase(fieldMap, extField))
					res, err := s.repo.UpsertRecordByExternalID(ctx, job.ObjectName, extField, extVal, fieldMap)
					if err != nil {
						opErr = err
					} else {
						sfID = res.ID
						sfCreated = res.Created
					}
				case "delete", "harddelete":
					idVal := stringifyFieldVal(getFieldIgnoreCase(fieldMap, "Id"))
					if idVal == "" {
						opErr = domain.ErrRequiredFieldMissing
					} else {
						err := s.repo.DeleteRecord(ctx, job.ObjectName, idVal)
						if err != nil {
							opErr = err
						} else {
							sfID = idVal
						}
					}
				}

				if opErr != nil {
					failed++
					_, _ = s.repo.CreateBulkIngestResult(ctx, &domain.BulkIngestResult{
						JobID:     jobID,
						SFID:      sfID,
						SFCreated: false,
						SFError:   opErr.Error(),
						IsSuccess: false,
						RowCSV:    rowCSV,
					})
				} else {
					_, _ = s.repo.CreateBulkIngestResult(ctx, &domain.BulkIngestResult{
						JobID:     jobID,
						SFID:      sfID,
						SFCreated: sfCreated,
						SFError:   "",
						IsSuccess: true,
						RowCSV:    rowCSV,
					})
				}
			}
		}
		return s.repo.UpdateBulkIngestJobStats(ctx, jobID, "JobComplete", processed, failed)
	default:
		return nil, domain.ErrInvalidJobState
	}
}

func (s *Service) GetIngestJob(ctx context.Context, version, jobID string) (*domain.BulkIngestJob, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, err
	}
	return s.repo.GetBulkIngestJob(ctx, jobID)
}

func (s *Service) ListIngestJobs(ctx context.Context, version string) ([]domain.BulkIngestJob, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, err
	}
	return s.repo.ListBulkIngestJobs(ctx)
}

func (s *Service) GetIngestResultsCSV(ctx context.Context, version, jobID string, isSuccess bool) (string, error) {
	if err := s.ValidateVersion(version); err != nil {
		return "", err
	}
	if _, err := s.repo.GetBulkIngestJob(ctx, jobID); err != nil {
		return "", err
	}
	batches, err := s.repo.ListBulkIngestBatches(ctx, jobID)
	if err != nil {
		return "", err
	}
	var origHeaders []string
	for _, b := range batches {
		r := csv.NewReader(strings.NewReader(strings.TrimSpace(b.CSVContent)))
		if hdr, err := r.Read(); err == nil && len(hdr) > 0 {
			origHeaders = hdr
			break
		}
	}
	results, err := s.repo.ListBulkIngestResults(ctx, jobID, isSuccess)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if isSuccess {
		hdr := append([]string{"sf__Id", "sf__Created"}, origHeaders...)
		_ = w.Write(hdr)
		for _, res := range results {
			createdStr := "false"
			if res.SFCreated {
				createdStr = "true"
			}
			var rowCols []string
			if res.RowCSV != "" {
				r := csv.NewReader(strings.NewReader(res.RowCSV))
				rowCols, _ = r.Read()
			}
			_ = w.Write(append([]string{res.SFID, createdStr}, rowCols...))
		}
	} else {
		hdr := append([]string{"sf__Id", "sf__Error"}, origHeaders...)
		_ = w.Write(hdr)
		for _, res := range results {
			var rowCols []string
			if res.RowCSV != "" {
				r := csv.NewReader(strings.NewReader(res.RowCSV))
				rowCols, _ = r.Read()
			}
			_ = w.Write(append([]string{res.SFID, res.SFError}, rowCols...))
		}
	}
	w.Flush()
	return buf.String(), nil
}

func (s *Service) CreateQueryJob(ctx context.Context, version, operation, query string) (*domain.BulkQueryJob, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, err
	}
	op := strings.ToLower(strings.TrimSpace(operation))
	if op == "" {
		op = "query"
	}
	if op != "query" && op != "queryall" {
		return nil, domain.ErrInvalidRequest
	}
	p, err := parseSOQL(query)
	if err != nil {
		return nil, err
	}
	job := &domain.BulkQueryJob{
		Operation:       op,
		QueryText:       query,
		ObjectName:      p.SObject,
		ContentType:     "CSV",
		ColumnDelimiter: "COMMA",
		LineEnding:      "LF",
		State:           "JobComplete",
		APIVersion:      normalizeVersion(version),
	}
	return s.repo.CreateBulkQueryJob(ctx, job)
}

func (s *Service) GetQueryJob(ctx context.Context, version, jobID string) (*domain.BulkQueryJob, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, err
	}
	return s.repo.GetBulkQueryJob(ctx, jobID)
}

func (s *Service) GetQueryJobResultsCSV(ctx context.Context, version, jobID string, maxRecords int) (string, int, error) {
	if err := s.ValidateVersion(version); err != nil {
		return "", 0, err
	}
	job, err := s.repo.GetBulkQueryJob(ctx, jobID)
	if err != nil {
		return "", 0, err
	}
	p, err := parseSOQL(job.QueryText)
	if err != nil {
		return "", 0, err
	}
	includeDeleted := strings.EqualFold(job.Operation, "queryall")
	rawRecords, err := s.repo.ListRecordsForSObject(ctx, normalizeVersion(version), p.SObject, includeDeleted)
	if err != nil {
		return "", 0, err
	}
	matching := applyParsedSOQL(rawRecords, p)
	if maxRecords > 0 && maxRecords < len(matching) {
		matching = matching[:maxRecords]
	}

	cols := p.Fields
	if len(cols) == 0 || (len(cols) == 1 && cols[0] == "*") {
		cols = []string{"Id", "Name"}
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write(cols)
	for _, rec := range matching {
		row := make([]string, 0, len(cols))
		for _, c := range cols {
			row = append(row, stringifyFieldVal(getFieldIgnoreCase(rec, c)))
		}
		_ = w.Write(row)
	}
	w.Flush()
	return buf.String(), len(matching), nil
}
