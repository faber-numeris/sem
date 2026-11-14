package domain

import "time"

type BulkIngestJob struct {
	ID                     string
	Operation              string
	ObjectName             string
	ExternalIDFieldName    string
	ContentType            string
	ColumnDelimiter        string
	LineEnding             string
	State                  string
	ConcurrencyMode        string
	APIVersion             string
	NumberRecordsProcessed int
	NumberRecordsFailed    int
	Retries                int
	TotalProcessingTime    int64
	CreatedByID            string
	CreatedDate            time.Time
	SystemModstamp         time.Time
}

type BulkIngestBatch struct {
	ID          string
	JobID       string
	CSVContent  string
	CreatedDate time.Time
}

type BulkIngestResult struct {
	ID        int64
	JobID     string
	SFID      string
	SFCreated bool
	SFError   string
	IsSuccess bool
	RowCSV    string
}

type BulkQueryJob struct {
	ID                     string
	Operation              string
	QueryText              string
	ObjectName             string
	ContentType            string
	ColumnDelimiter        string
	LineEnding             string
	State                  string
	ConcurrencyMode        string
	APIVersion             string
	NumberRecordsProcessed int
	Retries                int
	TotalProcessingTime    int64
	CreatedByID            string
	CreatedDate            time.Time
	SystemModstamp         time.Time
}
