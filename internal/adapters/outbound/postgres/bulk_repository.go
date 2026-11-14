package postgres

import (
	"context"

	"github.com/faber-numeris/sem/internal/adapters/outbound/postgres/gen"
	"github.com/faber-numeris/sem/internal/core/domain"
)

func mapBulkIngestJob(j gen.SalesforceBulkIngestJob) *domain.BulkIngestJob {
	ext := ""
	if j.ExternalIDFieldName != nil {
		ext = *j.ExternalIDFieldName
	}
	return &domain.BulkIngestJob{
		ID:                     j.ID,
		Operation:              j.Operation,
		ObjectName:             j.ObjectName,
		ExternalIDFieldName:    ext,
		ContentType:            j.ContentType,
		ColumnDelimiter:        j.ColumnDelimiter,
		LineEnding:             j.LineEnding,
		State:                  j.State,
		ConcurrencyMode:        j.ConcurrencyMode,
		APIVersion:             j.ApiVersion,
		NumberRecordsProcessed: j.NumberRecordsProcessed,
		NumberRecordsFailed:    j.NumberRecordsFailed,
		Retries:                j.Retries,
		TotalProcessingTime:    j.TotalProcessingTime,
		CreatedByID:            j.CreatedByID,
		CreatedDate:            j.CreatedDate,
		SystemModstamp:         j.SystemModstamp,
	}
}

func (r *Repository) CreateBulkIngestJob(ctx context.Context, job *domain.BulkIngestJob) (*domain.BulkIngestJob, error) {
	var extPtr *string
	if job.ExternalIDFieldName != "" {
		ext := job.ExternalIDFieldName
		extPtr = &ext
	}
	created, err := r.queries.CreateBulkIngestJob(ctx, gen.CreateBulkIngestJobParams{
		Operation:           job.Operation,
		ObjectName:          job.ObjectName,
		ExternalIDFieldName: extPtr,
		ContentType:         job.ContentType,
		ColumnDelimiter:     job.ColumnDelimiter,
		LineEnding:          job.LineEnding,
		ApiVersion:          job.APIVersion,
	})
	if err != nil {
		return nil, err
	}
	return mapBulkIngestJob(created), nil
}

func (r *Repository) GetBulkIngestJob(ctx context.Context, id string) (*domain.BulkIngestJob, error) {
	j, err := r.queries.GetBulkIngestJob(ctx, id)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	return mapBulkIngestJob(j), nil
}

func (r *Repository) UpdateBulkIngestJobState(ctx context.Context, id, state string) (*domain.BulkIngestJob, error) {
	j, err := r.queries.UpdateBulkIngestJobState(ctx, gen.UpdateBulkIngestJobStateParams{
		State: state,
		ID:    id,
	})
	if err != nil {
		return nil, domain.ErrNotFound
	}
	return mapBulkIngestJob(j), nil
}

func (r *Repository) UpdateBulkIngestJobStats(ctx context.Context, id, state string, processed, failed int) (*domain.BulkIngestJob, error) {
	j, err := r.queries.UpdateBulkIngestJobStats(ctx, gen.UpdateBulkIngestJobStatsParams{
		State:                  state,
		NumberRecordsProcessed: processed,
		NumberRecordsFailed:    failed,
		ID:                     id,
	})
	if err != nil {
		return nil, domain.ErrNotFound
	}
	return mapBulkIngestJob(j), nil
}

func (r *Repository) ListBulkIngestJobs(ctx context.Context) ([]domain.BulkIngestJob, error) {
	rows, err := r.queries.ListBulkIngestJobs(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]domain.BulkIngestJob, 0, len(rows))
	for _, item := range rows {
		res = append(res, *mapBulkIngestJob(item))
	}
	return res, nil
}

func (r *Repository) CreateBulkIngestBatch(ctx context.Context, jobID, csvContent string) (*domain.BulkIngestBatch, error) {
	b, err := r.queries.CreateBulkIngestBatch(ctx, gen.CreateBulkIngestBatchParams{
		JobID:      jobID,
		CsvContent: csvContent,
	})
	if err != nil {
		return nil, err
	}
	return &domain.BulkIngestBatch{
		ID:          b.ID,
		JobID:       b.JobID,
		CSVContent:  b.CsvContent,
		CreatedDate: b.CreatedDate,
	}, nil
}

func (r *Repository) ListBulkIngestBatches(ctx context.Context, jobID string) ([]domain.BulkIngestBatch, error) {
	rows, err := r.queries.ListBulkIngestBatchesByJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	res := make([]domain.BulkIngestBatch, 0, len(rows))
	for _, b := range rows {
		res = append(res, domain.BulkIngestBatch{
			ID:          b.ID,
			JobID:       b.JobID,
			CSVContent:  b.CsvContent,
			CreatedDate: b.CreatedDate,
		})
	}
	return res, nil
}

func (r *Repository) CreateBulkIngestResult(ctx context.Context, res *domain.BulkIngestResult) (*domain.BulkIngestResult, error) {
	created, err := r.queries.CreateBulkIngestResult(ctx, gen.CreateBulkIngestResultParams{
		JobID:     res.JobID,
		SfID:      res.SFID,
		SfCreated: res.SFCreated,
		SfError:   res.SFError,
		IsSuccess: res.IsSuccess,
		RowCsv:    res.RowCSV,
	})
	if err != nil {
		return nil, err
	}
	return &domain.BulkIngestResult{
		ID:        created.ID,
		JobID:     created.JobID,
		SFID:      created.SfID,
		SFCreated: created.SfCreated,
		SFError:   created.SfError,
		IsSuccess: created.IsSuccess,
		RowCSV:    created.RowCsv,
	}, nil
}

func (r *Repository) ListBulkIngestResults(ctx context.Context, jobID string, isSuccess bool) ([]domain.BulkIngestResult, error) {
	rows, err := r.queries.ListBulkIngestResults(ctx, gen.ListBulkIngestResultsParams{
		JobID:     jobID,
		IsSuccess: isSuccess,
	})
	if err != nil {
		return nil, err
	}
	res := make([]domain.BulkIngestResult, 0, len(rows))
	for _, item := range rows {
		res = append(res, domain.BulkIngestResult{
			ID:        item.ID,
			JobID:     item.JobID,
			SFID:      item.SfID,
			SFCreated: item.SfCreated,
			SFError:   item.SfError,
			IsSuccess: item.IsSuccess,
			RowCSV:    item.RowCsv,
		})
	}
	return res, nil
}

func mapBulkQueryJob(j gen.SalesforceBulkQueryJob) *domain.BulkQueryJob {
	return &domain.BulkQueryJob{
		ID:                     j.ID,
		Operation:              j.Operation,
		QueryText:              j.QueryText,
		ObjectName:             j.ObjectName,
		ContentType:            j.ContentType,
		ColumnDelimiter:        j.ColumnDelimiter,
		LineEnding:             j.LineEnding,
		State:                  j.State,
		ConcurrencyMode:        j.ConcurrencyMode,
		APIVersion:             j.ApiVersion,
		NumberRecordsProcessed: j.NumberRecordsProcessed,
		Retries:                j.Retries,
		TotalProcessingTime:    j.TotalProcessingTime,
		CreatedByID:            j.CreatedByID,
		CreatedDate:            j.CreatedDate,
		SystemModstamp:         j.SystemModstamp,
	}
}

func (r *Repository) CreateBulkQueryJob(ctx context.Context, job *domain.BulkQueryJob) (*domain.BulkQueryJob, error) {
	created, err := r.queries.CreateBulkQueryJob(ctx, gen.CreateBulkQueryJobParams{
		Operation:       job.Operation,
		QueryText:       job.QueryText,
		ObjectName:      job.ObjectName,
		ContentType:     job.ContentType,
		ColumnDelimiter: job.ColumnDelimiter,
		LineEnding:      job.LineEnding,
		State:           job.State,
		ApiVersion:      job.APIVersion,
	})
	if err != nil {
		return nil, err
	}
	return mapBulkQueryJob(created), nil
}

func (r *Repository) GetBulkQueryJob(ctx context.Context, id string) (*domain.BulkQueryJob, error) {
	j, err := r.queries.GetBulkQueryJob(ctx, id)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	return mapBulkQueryJob(j), nil
}
