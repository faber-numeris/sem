package inbound

import (
	"context"

	"github.com/faber-numeris/sem/internal/core/domain"
)

type AuthService interface {
	IssueToken(ctx context.Context, req domain.TokenRequest) (*domain.OAuthToken, error)
	Authorize(ctx context.Context, req domain.AuthorizeRequest) (*domain.AuthorizeResult, error)
	RevokeToken(ctx context.Context, token string) error
	IntrospectToken(ctx context.Context, token string) (*domain.IntrospectResult, error)
	GetUserInfo(ctx context.Context, accessToken, baseURL string) (*domain.UserInfoResult, error)
}

type SObjectService interface {
	ListVersions() []domain.APIVersion
	ValidateVersion(version string) error
	ListResources(version string) (map[string]string, error)
	GetLimits(ctx context.Context, version string) (map[string]domain.OrgLimit, error)
	DescribeGlobal(ctx context.Context, version string, isTooling bool) (map[string]any, error)
	GetSObjectBasicInfo(ctx context.Context, version, sobject string) (map[string]any, error)
	DescribeSObject(ctx context.Context, version, sobject string, isTooling bool) (map[string]any, error)
	CreateRecord(ctx context.Context, version, sobject string, body map[string]any) (*domain.SaveResult, map[string]any, error)
	GetRecord(ctx context.Context, version, sobject, id string, fields []string) (map[string]any, error)
	UpdateRecord(ctx context.Context, version, sobject, id string, body map[string]any) (map[string]any, error)
	UpsertRecord(ctx context.Context, version, sobject, externalIDField, externalIDValue string, body map[string]any) (*domain.SaveResult, error)
	DeleteRecord(ctx context.Context, version, sobject, id string) error
	GetContentVersionData(ctx context.Context, version, id string) ([]byte, error)
}

type QueryService interface {
	ExecuteSOQL(ctx context.Context, version, soql string, includeDeleted bool, batchSize int) (*domain.QueryResult, error)
	QueryNextBatch(ctx context.Context, version, cursorID string) (*domain.QueryResult, error)
	ExecuteSOSL(ctx context.Context, version, sosl string) (*domain.SearchResult, error)
}

type BulkService interface {
	CreateIngestJob(ctx context.Context, version, operation, objectName, contentType, externalIDField string) (*domain.BulkIngestJob, error)
	UploadIngestBatch(ctx context.Context, version, jobID, csvData string) error
	CloseOrAbortIngestJob(ctx context.Context, version, jobID, state string) (*domain.BulkIngestJob, error)
	GetIngestJob(ctx context.Context, version, jobID string) (*domain.BulkIngestJob, error)
	ListIngestJobs(ctx context.Context, version string) ([]domain.BulkIngestJob, error)
	GetIngestResultsCSV(ctx context.Context, version, jobID string, isSuccess bool) (string, error)
	CreateQueryJob(ctx context.Context, version, operation, query string) (*domain.BulkQueryJob, error)
	GetQueryJob(ctx context.Context, version, jobID string) (*domain.BulkQueryJob, error)
	GetQueryJobResultsCSV(ctx context.Context, version, jobID string, maxRecords int) (string, int, error)
}

type ToolingService interface {
	ExecuteToolingQuery(ctx context.Context, version, soql string) (*domain.QueryResult, error)
	ExecuteAnonymousApex(ctx context.Context, version, body string) (*domain.AnonymousApexResult, error)
	DescribeToolingSObject(ctx context.Context, version, sobject string) (map[string]any, error)
	RunTestsAsynchronous(ctx context.Context, version, classNames, testLevel string) (string, error)
	ExecuteApexRest(ctx context.Context, urlMapping, method, reqBody string) (int, string, error)
	ProcessCometD(ctx context.Context, version string, messages []domain.BayeuxMessage) ([]domain.BayeuxMessage, error)
}

type EmulatorService interface {
	AuthService
	SObjectService
	QueryService
	BulkService
	ToolingService
}
