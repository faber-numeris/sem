package domain

type QueryResult struct {
	TotalSize      int              `json:"totalSize"`
	Done           bool             `json:"done"`
	NextRecordsURL string           `json:"nextRecordsUrl,omitempty"`
	Records        []map[string]any `json:"records"`
}

type QueryCursor struct {
	ID             string
	SObjectName    string
	SOQLQuery      string
	IncludeDeleted bool
	OffsetValue    int
	BatchSize      int
	TotalSize      int
}

type SearchResult struct {
	SearchRecords []map[string]any `json:"searchRecords"`
}
