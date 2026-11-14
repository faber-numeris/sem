package outbound

import (
	"context"

	"github.com/faber-numeris/sem/internal/core/domain"
)

type AuthRepository interface {
	GetClient(ctx context.Context, clientID string) (*domain.OAuthClient, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.SalesforceUser, error)
	GetUserByID(ctx context.Context, id string) (*domain.SalesforceUser, error)
	GetDefaultUser(ctx context.Context) (*domain.SalesforceUser, error)
	CreateAuthCode(ctx context.Context, code *domain.OAuthAuthCode) (*domain.OAuthAuthCode, error)
	GetAuthCode(ctx context.Context, code string) (*domain.OAuthAuthCode, error)
	MarkAuthCodeUsed(ctx context.Context, code string) error
	CreateToken(ctx context.Context, token *domain.OAuthToken) (*domain.OAuthToken, error)
	GetTokenByAccess(ctx context.Context, accessToken string) (*domain.OAuthToken, error)
	GetTokenByRefresh(ctx context.Context, refreshToken string) (*domain.OAuthToken, error)
	GetTokenByAny(ctx context.Context, token string) (*domain.OAuthToken, error)
	RevokeToken(ctx context.Context, token string) error
}

type SObjectRepository interface {
	ListOrgLimits(ctx context.Context) ([]domain.OrgLimit, error)
	ListSObjectDefinitions(ctx context.Context, isTooling bool) ([]domain.SObjectDefinition, error)
	GetSObjectDescribe(ctx context.Context, name string) (*domain.SObjectDescribe, error)
	CreateRecord(ctx context.Context, sobject string, fields map[string]any) (*domain.SaveResult, map[string]any, error)
	GetRecord(ctx context.Context, version, sobject, id string) (map[string]any, error)
	UpdateRecord(ctx context.Context, sobject, id string, fields map[string]any) (map[string]any, error)
	UpsertRecordByExternalID(ctx context.Context, sobject, externalIDField, externalIDValue string, fields map[string]any) (*domain.SaveResult, error)
	DeleteRecord(ctx context.Context, sobject, id string) error
	GetContentVersionData(ctx context.Context, id string) ([]byte, error)
	ListRecordsForSObject(ctx context.Context, version, sobject string, includeDeleted bool) ([]map[string]any, error)
	SearchRecordsForSObject(ctx context.Context, version, sobject, term string) ([]map[string]any, error)
}

type QueryRepository interface {
	CreateQueryCursor(ctx context.Context, cursor *domain.QueryCursor) (*domain.QueryCursor, error)
	GetQueryCursor(ctx context.Context, id string) (*domain.QueryCursor, error)
	UpdateQueryCursorOffset(ctx context.Context, id string, offset int) (*domain.QueryCursor, error)
	DeleteQueryCursor(ctx context.Context, id string) error
}

type BulkRepository interface {
	CreateBulkIngestJob(ctx context.Context, job *domain.BulkIngestJob) (*domain.BulkIngestJob, error)
	GetBulkIngestJob(ctx context.Context, id string) (*domain.BulkIngestJob, error)
	UpdateBulkIngestJobState(ctx context.Context, id, state string) (*domain.BulkIngestJob, error)
	UpdateBulkIngestJobStats(ctx context.Context, id, state string, processed, failed int) (*domain.BulkIngestJob, error)
	ListBulkIngestJobs(ctx context.Context) ([]domain.BulkIngestJob, error)
	CreateBulkIngestBatch(ctx context.Context, jobID, csvContent string) (*domain.BulkIngestBatch, error)
	ListBulkIngestBatches(ctx context.Context, jobID string) ([]domain.BulkIngestBatch, error)
	CreateBulkIngestResult(ctx context.Context, res *domain.BulkIngestResult) (*domain.BulkIngestResult, error)
	ListBulkIngestResults(ctx context.Context, jobID string, isSuccess bool) ([]domain.BulkIngestResult, error)
	CreateBulkQueryJob(ctx context.Context, job *domain.BulkQueryJob) (*domain.BulkQueryJob, error)
	GetBulkQueryJob(ctx context.Context, id string) (*domain.BulkQueryJob, error)
}

type ToolingRepository interface {
	ListToolingRecords(ctx context.Context, version, sobject string) ([]map[string]any, error)
	CreateApexTestRun(ctx context.Context, classNames, testLevel string) (*domain.ApexTestRun, error)
	GetApexRestRoute(ctx context.Context, urlMapping, method string) (*domain.ApexRestRoute, error)
	CreatePlatformEvent(ctx context.Context, channel, eventType, payload string) (*domain.PlatformEvent, error)
	ListPlatformEventsSince(ctx context.Context, channel string, afterReplayID int64) ([]domain.PlatformEvent, error)
	UpsertCometDClient(ctx context.Context, clientID string) (*domain.CometDClient, error)
	GetCometDClient(ctx context.Context, clientID string) (*domain.CometDClient, error)
	DisconnectCometDClient(ctx context.Context, clientID string) error
	UpsertCometDSubscription(ctx context.Context, clientID, channel string, replayID int64) (*domain.CometDSubscription, error)
	DeleteCometDSubscription(ctx context.Context, clientID, channel string) error
	ListCometDSubscriptions(ctx context.Context, clientID string) ([]domain.CometDSubscription, error)
	UpdateCometDSubscriptionReplayID(ctx context.Context, id, replayID int64) error
}

type Repository interface {
	AuthRepository
	SObjectRepository
	QueryRepository
	BulkRepository
	ToolingRepository
	Ping() bool
}
