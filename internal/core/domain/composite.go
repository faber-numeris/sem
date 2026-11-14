package domain

type CompositeSubRequest struct {
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	ReferenceID string            `json:"referenceId"`
	Body        any               `json:"body,omitempty"`
	HTTPHeaders map[string]string `json:"httpHeaders,omitempty"`
}

type CompositeRequest struct {
	AllOrNone        bool                  `json:"allOrNone"`
	CollateSubreqs   bool                  `json:"collateSubrequests"`
	CompositeRequest []CompositeSubRequest `json:"compositeRequest"`
}

type CompositeSubResponse struct {
	Body           any               `json:"body"`
	HTTPHeaders    map[string]string `json:"httpHeaders"`
	HTTPStatusCode int               `json:"httpStatusCode"`
	ReferenceID    string            `json:"referenceId"`
}

type CompositeResponse struct {
	CompositeResponse []CompositeSubResponse `json:"compositeResponse"`
}

type BatchSubRequest struct {
	Method            string `json:"method"`
	URL               string `json:"url"`
	RichInput         any    `json:"richInput,omitempty"`
	BinaryPartName    string `json:"binaryPartName,omitempty"`
	BinaryPartNameAl  string `json:"binaryPartNameAlias,omitempty"`
}

type BatchRequest struct {
	HaltOnError   bool              `json:"haltOnError"`
	BatchRequests []BatchSubRequest `json:"batchRequests"`
}

type BatchSubResponse struct {
	StatusCode int `json:"statusCode"`
	Result     any `json:"result"`
}

type BatchResponse struct {
	HasErrors bool               `json:"hasErrors"`
	Results   []BatchSubResponse `json:"results"`
}
