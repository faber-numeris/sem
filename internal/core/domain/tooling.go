package domain

import "time"

type ApexClass struct {
	ID                    string
	Name                  string
	Body                  string
	APIVersion            float64
	Status                string
	IsValid               bool
	LengthWithoutComments int
	NamespacePrefix       string
	CreatedByID           string
	CreatedDate           time.Time
	LastModifiedByID      string
	LastModifiedDate      time.Time
	SystemModstamp        time.Time
}

type ApexTrigger struct {
	ID                string
	Name              string
	TableEnumOrID     string
	Body              string
	APIVersion        float64
	Status            string
	IsValid           bool
	UsageBeforeInsert bool
	UsageAfterInsert  bool
	UsageBeforeUpdate bool
	UsageAfterUpdate  bool
	CreatedByID       string
	CreatedDate       time.Time
	LastModifiedDate  time.Time
}

type CustomField struct {
	ID               string
	DeveloperName    string
	TableEnumOrID    string
	FullName         string
	Label            string
	DataType         string
	Length           int
	CreatedDate      time.Time
	LastModifiedDate time.Time
}

type ValidationRule struct {
	ID                    string
	ValidationName        string
	EntityDefinitionID    string
	Active                bool
	Description           string
	ErrorConditionFormula string
	ErrorMessage          string
	CreatedDate           time.Time
	LastModifiedDate      time.Time
}

type AnonymousApexResult struct {
	Line                int     `json:"line"`
	Column              int     `json:"column"`
	Compiled            bool    `json:"compiled"`
	Success             bool    `json:"success"`
	CompileProblem      *string `json:"compileProblem"`
	ExceptionStackTrace *string `json:"exceptionStackTrace"`
	ExceptionMessage    *string `json:"exceptionMessage"`
}

type ApexTestRun struct {
	ID               string
	ClassNames       string
	TestLevel        string
	Status           string
	MethodsEnqueued  int
	MethodsCompleted int
	MethodsFailed    int
	CreatedDate      time.Time
}

type ApexRestRoute struct {
	ID             string
	URLMapping     string
	HTTPMethod     string
	ClassName      string
	ResponseStatus int
	ResponseBody   string
	IsActive       bool
}

type PlatformEvent struct {
	ID          string
	ReplayID    int64
	ChannelName string
	EventType   string
	Payload     string
	CreatedByID string
	CreatedDate time.Time
}

type CometDClient struct {
	ClientID    string
	IsConnected bool
}

type CometDSubscription struct {
	ID           int64
	ClientID     string
	ChannelName  string
	LastReplayID int64
}

type BayeuxMessage struct {
	Channel                  string         `json:"channel"`
	ID                       string         `json:"id,omitempty"`
	ClientID                 string         `json:"clientId,omitempty"`
	Version                  string         `json:"version,omitempty"`
	MinimumVersion           string         `json:"minimumVersion,omitempty"`
	SupportedConnectionTypes []string       `json:"supportedConnectionTypes,omitempty"`
	ConnectionType           string         `json:"connectionType,omitempty"`
	Subscription             string         `json:"subscription,omitempty"`
	Data                     map[string]any `json:"data,omitempty"`
	Successful               *bool          `json:"successful,omitempty"`
	Error                    string         `json:"error,omitempty"`
	Advice                   map[string]any `json:"advice,omitempty"`
	Ext                      map[string]any `json:"ext,omitempty"`
}
