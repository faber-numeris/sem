package domain

import "time"

type APIVersion struct {
	Label   string `json:"label"`
	URL     string `json:"url"`
	Version string `json:"version"`
}

type OrgLimit struct {
	Name      string
	Max       int `json:"Max"`
	Remaining int `json:"Remaining"`
}

type SObjectDefinition struct {
	Name          string
	Label         string
	LabelPlural   string
	KeyPrefix     string
	Custom        bool
	Createable    bool
	Updateable    bool
	Deletable     bool
	Queryable     bool
	Searchable    bool
	Replicateable bool
	Retrieveable  bool
	Undeletable   bool
	Triggerable   bool
	IsTooling     bool
	TableName     string
}

type SObjectFieldDefinition struct {
	SObjectName string
	Name        string
	ColumnName  string
	Label       string
	FieldType   string
	SoapType    string
	Length      int
	Nillable    bool
	Createable  bool
	Updateable  bool
	Filterable  bool
	Sortable    bool
	Custom      bool
	ExternalID  bool
	UniqueField bool
	NameField   bool
	ReferenceTo string
}

type SObjectDescribe struct {
	Definition SObjectDefinition
	Fields     []SObjectFieldDefinition
}

type RecordAttributes struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type SaveResult struct {
	ID      string   `json:"id"`
	Success bool     `json:"success"`
	Errors  []string `json:"errors"`
	Created bool     `json:"created,omitempty"`
}

type Account struct {
	ID                string
	ExternalIDC       string
	Name              string
	Type              string
	ParentID          string
	BillingStreet     string
	BillingCity       string
	BillingState      string
	BillingPostalCode string
	BillingCountry    string
	Phone             string
	Website           string
	Industry          string
	AnnualRevenue     *float64
	NumberOfEmployees *int
	Description       string
	Rating            string
	Site              string
	OwnerID           string
	CreatedByID       string
	CreatedDate       time.Time
	LastModifiedByID  string
	LastModifiedDate  time.Time
	SystemModstamp    time.Time
	IsDeleted         bool
}

type Contact struct {
	ID                string
	ExternalIDC       string
	AccountID         string
	LastName          string
	FirstName         string
	Salutation        string
	Name              string
	MailingStreet     string
	MailingCity       string
	MailingState      string
	MailingPostalCode string
	MailingCountry    string
	Phone             string
	MobilePhone       string
	Email             string
	Title             string
	Department        string
	LeadSource        string
	Description       string
	OwnerID           string
	CreatedByID       string
	CreatedDate       time.Time
	LastModifiedByID  string
	LastModifiedDate  time.Time
	SystemModstamp    time.Time
	IsDeleted         bool
}

type Lead struct {
	ID                string
	ExternalIDC       string
	LastName          string
	FirstName         string
	Salutation        string
	Name              string
	Title             string
	Company           string
	Street            string
	City              string
	State             string
	PostalCode        string
	Country           string
	Phone             string
	MobilePhone       string
	Email             string
	Website           string
	Description       string
	LeadSource        string
	Status            string
	Industry          string
	Rating            string
	AnnualRevenue     *float64
	NumberOfEmployees *int
	IsConverted       bool
	OwnerID           string
	CreatedByID       string
	CreatedDate       time.Time
	LastModifiedByID  string
	LastModifiedDate  time.Time
	SystemModstamp    time.Time
	IsDeleted         bool
}

type Case struct {
	ID                  string
	ExternalIDC         string
	CaseNumber          string
	AccountID           string
	ContactID           string
	ParentID            string
	RecordTypeID        string
	TypeName            string
	Status              string
	Origin              string
	Subject             string
	Priority            string
	Description         string
	AdditionalInfoC     string
	SubTypeC            string
	TypeC               string
	PopulationImpactedC string
	WeblinksC           string
	RootCauseC          string
	ImpactC             string
	VIPC                bool
	OwnerID             string
	CreatedByID         string
	CreatedDate         time.Time
	ClosedDate          *time.Time
	LastModifiedDate    time.Time
	SystemModstamp      time.Time
	IsDeleted           bool
}

type CustomObject struct {
	ID               string
	ExternalIDC      string
	Name             string
	FieldToUpdateC   string
	DescriptionC     string
	StatusC          string
	OwnerID          string
	CreatedByID      string
	CreatedDate      time.Time
	LastModifiedByID string
	LastModifiedDate time.Time
	SystemModstamp   time.Time
	IsDeleted        bool
}

type ContentDocument struct {
	ID          string
	Type        string
	Title       string
	CreatedDate time.Time
	IsDeleted   bool
}

type ContentVersion struct {
	ID                string
	ContentDocumentID string
	IsLatest          bool
	VersionNumber     string
	Title             string
	Description       string
	PathOnClient      string
	FileType          string
	VersionData       string
	VersionDataURL    string
	OwnerID           string
	CreatedByID       string
	CreatedDate       time.Time
	LastModifiedDate  time.Time
	IsDeleted         bool
}

type ContentDocumentLink struct {
	ID                string
	ContentDocumentID string
	LinkedEntityID    string
	ShareType         string
	Visibility        string
	IsDeleted         bool
}
