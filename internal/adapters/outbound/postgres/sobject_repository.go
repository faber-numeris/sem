package postgres

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/faber-numeris/sem/internal/adapters/outbound/postgres/gen"
	"github.com/faber-numeris/sem/internal/core/domain"
)

func (r *Repository) ListOrgLimits(ctx context.Context) ([]domain.OrgLimit, error) {
	rows, err := r.queries.ListOrgLimits(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]domain.OrgLimit, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.OrgLimit{
			Name:      row.LimitName,
			Max:       row.MaxLimit,
			Remaining: row.RemainingLimit,
		})
	}
	return res, nil
}

func mapSObjectDef(d gen.SalesforceSobjectDefinition) domain.SObjectDefinition {
	return domain.SObjectDefinition{
		Name:          d.Name,
		Label:         d.Label,
		LabelPlural:   d.LabelPlural,
		KeyPrefix:     d.KeyPrefix,
		Custom:        d.Custom,
		Createable:    d.Createable,
		Updateable:    d.Updateable,
		Deletable:     d.Deletable,
		Queryable:     d.Queryable,
		Searchable:    d.Searchable,
		Replicateable: d.Replicateable,
		Retrieveable:  d.Retrieveable,
		Undeletable:   d.Undeletable,
		Triggerable:   d.Triggerable,
		IsTooling:     d.IsTooling,
		TableName:     d.TableName,
	}
}

func (r *Repository) ListSObjectDefinitions(ctx context.Context, isTooling bool) ([]domain.SObjectDefinition, error) {
	rows, err := r.queries.ListSObjectDefinitions(ctx, isTooling)
	if err != nil {
		return nil, err
	}
	res := make([]domain.SObjectDefinition, 0, len(rows))
	for _, row := range rows {
		res = append(res, mapSObjectDef(row))
	}
	return res, nil
}

func (r *Repository) GetSObjectDescribe(ctx context.Context, name string) (*domain.SObjectDescribe, error) {
	def, err := r.queries.GetSObjectDefinition(ctx, name)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	fields, err := r.queries.ListSObjectFieldDefinitions(ctx, def.Name)
	if err != nil {
		return nil, err
	}
	mappedFields := make([]domain.SObjectFieldDefinition, 0, len(fields))
	for _, f := range fields {
		refTo := ""
		if f.ReferenceTo != nil {
			refTo = *f.ReferenceTo
		}
		mappedFields = append(mappedFields, domain.SObjectFieldDefinition{
			SObjectName: f.SobjectName,
			Name:        f.Name,
			ColumnName:  f.ColumnName,
			Label:       f.Label,
			FieldType:   f.FieldType,
			SoapType:    f.SoapType,
			Length:      f.Length,
			Nillable:    f.Nillable,
			Createable:  f.Createable,
			Updateable:  f.Updateable,
			Filterable:  f.Filterable,
			Sortable:    f.Sortable,
			Custom:      f.Custom,
			ExternalID:  f.ExternalID,
			UniqueField: f.UniqueField,
			NameField:   f.NameField,
			ReferenceTo: refTo,
		})
	}
	return &domain.SObjectDescribe{
		Definition: mapSObjectDef(def),
		Fields:     mappedFields,
	}, nil
}

func accountToMap(version string, a gen.SalesforceAccount) map[string]any {
	return map[string]any{
		"attributes":         makeAttributes(version, "Account", a.ID),
		"Id":                 a.ID,
		"External_ID__c":     a.ExternalIDC,
		"Name":               a.Name,
		"Type":               a.Type,
		"ParentId":           a.ParentID,
		"BillingStreet":      a.BillingStreet,
		"BillingCity":        a.BillingCity,
		"BillingState":       a.BillingState,
		"BillingPostalCode":  a.BillingPostalCode,
		"BillingCountry":     a.BillingCountry,
		"ShippingStreet":     a.ShippingStreet,
		"ShippingCity":       a.ShippingCity,
		"ShippingState":      a.ShippingState,
		"ShippingPostalCode": a.ShippingPostalCode,
		"ShippingCountry":    a.ShippingCountry,
		"Phone":              a.Phone,
		"Fax":                a.Fax,
		"AccountNumber":      a.AccountNumber,
		"Website":            a.Website,
		"Industry":           a.Industry,
		"AnnualRevenue":      a.AnnualRevenue,
		"NumberOfEmployees":  a.NumberOfEmployees,
		"Ownership":          a.Ownership,
		"TickerSymbol":       a.TickerSymbol,
		"Description":        a.Description,
		"Rating":             a.Rating,
		"Site":               a.Site,
		"OwnerId":            a.OwnerID,
		"CreatedById":        a.CreatedByID,
		"CreatedDate":        formatSFTime(a.CreatedDate),
		"LastModifiedById":   a.LastModifiedByID,
		"LastModifiedDate":   formatSFTime(a.LastModifiedDate),
		"SystemModstamp":     formatSFTime(a.SystemModstamp),
		"IsDeleted":          a.IsDeleted,
	}
}

func contactToMap(version string, c gen.SalesforceContact) map[string]any {
	return map[string]any{
		"attributes":        makeAttributes(version, "Contact", c.ID),
		"Id":                c.ID,
		"External_ID__c":    c.ExternalIDC,
		"AccountId":         c.AccountID,
		"LastName":          c.LastName,
		"FirstName":         c.FirstName,
		"Salutation":        c.Salutation,
		"Name":              c.Name,
		"MailingStreet":     c.MailingStreet,
		"MailingCity":       c.MailingCity,
		"MailingState":      c.MailingState,
		"MailingPostalCode": c.MailingPostalCode,
		"MailingCountry":    c.MailingCountry,
		"Phone":             c.Phone,
		"Fax":               c.Fax,
		"MobilePhone":       c.MobilePhone,
		"Email":             c.Email,
		"Title":             c.Title,
		"Department":        c.Department,
		"LeadSource":        c.LeadSource,
		"Description":       c.Description,
		"OwnerId":           c.OwnerID,
		"CreatedById":       c.CreatedByID,
		"CreatedDate":       formatSFTime(c.CreatedDate),
		"LastModifiedById":  c.LastModifiedByID,
		"LastModifiedDate":  formatSFTime(c.LastModifiedDate),
		"SystemModstamp":    formatSFTime(c.SystemModstamp),
		"IsDeleted":         c.IsDeleted,
	}
}

func leadToMap(version string, l gen.SalesforceLead) map[string]any {
	return map[string]any{
		"attributes":         makeAttributes(version, "Lead", l.ID),
		"Id":                 l.ID,
		"External_ID__c":     l.ExternalIDC,
		"LastName":           l.LastName,
		"FirstName":          l.FirstName,
		"Salutation":         l.Salutation,
		"Name":               l.Name,
		"Title":              l.Title,
		"Company":            l.Company,
		"Street":             l.Street,
		"City":               l.City,
		"State":              l.State,
		"PostalCode":         l.PostalCode,
		"Country":            l.Country,
		"Phone":              l.Phone,
		"MobilePhone":        l.MobilePhone,
		"Email":              l.Email,
		"Website":            l.Website,
		"Description":        l.Description,
		"LeadSource":         l.LeadSource,
		"Status":             l.Status,
		"Industry":           l.Industry,
		"Rating":             l.Rating,
		"AnnualRevenue":      l.AnnualRevenue,
		"NumberOfEmployees":  l.NumberOfEmployees,
		"IsConverted":        l.IsConverted,
		"ConvertedAccountId": l.ConvertedAccountID,
		"ConvertedContactId": l.ConvertedContactID,
		"OwnerId":            l.OwnerID,
		"CreatedById":        l.CreatedByID,
		"CreatedDate":        formatSFTime(l.CreatedDate),
		"LastModifiedById":   l.LastModifiedByID,
		"LastModifiedDate":   formatSFTime(l.LastModifiedDate),
		"SystemModstamp":     formatSFTime(l.SystemModstamp),
		"IsDeleted":          l.IsDeleted,
	}
}

func caseToMap(version string, c gen.SalesforceCase) map[string]any {
	return map[string]any{
		"attributes":                   makeAttributes(version, "Case", c.ID),
		"Id":                           c.ID,
		"External_ID__c":               c.ExternalIDC,
		"CaseNumber":                   c.CaseNumber,
		"AdditionalInfo__c":            c.AdditionalinfoC,
		"Sub_Type__c":                  c.SubTypeC,
		"Type__c":                      c.TypeC,
		"Type":                         c.TypeName,
		"Origin":                       c.Origin,
		"RecordTypeId":                 c.RecordTypeID,
		"ParentId":                     c.ParentID,
		"Population_Impacted__c":       c.PopulationImpactedC,
		"ClosedDate":                   formatSFTimePtr(c.ClosedDate),
		"CreatedDate":                  formatSFTime(c.CreatedDate),
		"Weblinks__c":                  c.WeblinksC,
		"Root_Cause__c":                c.RootCauseC,
		"Impact__c":                    c.ImpactC,
		"VIP__c":                       c.VipC,
		"Priority":                     c.Priority,
		"AccountId":                    c.AccountID,
		"ContactId":                    c.ContactID,
		"Recurrence__c":                c.RecurrenceC,
		"Human_Factor__c":              c.HumanFactorC,
		"Technical_System__c":          c.TechnicalSystemC,
		"Contact_Conclusion_Level_2__c": c.ContactConclusionLevel2C,
		"Contact_Conclusion_Level_1__c": c.ContactConclusionLevel1C,
		"Contact_Action_Level_3__c":     c.ContactActionLevel3C,
		"Contact_Action_Level_2__c":     c.ContactActionLevel2C,
		"Contact_Action_Level_1__c":     c.ContactActionLevel1C,
		"Contact_Topic_Level_3__c":      c.ContactTopicLevel3C,
		"Contact_Topic_Level_2__c":      c.ContactTopicLevel2C,
		"Contact_Topic_Level_1__c":      c.ContactTopicLevel1C,
		"Components__c":                 c.ComponentsC,
		"Steps_To_Reproduce__c":         c.StepsToReproduceC,
		"Expected_Result__c":            c.ExpectedResultC,
		"Actual_Result__c":              c.ActualResultC,
		"Description":                   c.Description,
		"OwnerId":                       c.OwnerID,
		"Sub_Status_Level_3__c":         c.SubStatusLevel3C,
		"Sub_Status_Level_2__c":         c.SubStatusLevel2C,
		"Sub_Status__c":                 c.SubStatusC,
		"Status":                        c.Status,
		"CreatedById":                   c.CreatedByID,
		"Subject":                       c.Subject,
		"Send_Communication__c":         c.SendCommunicationC,
		"LastModifiedDate":              formatSFTime(c.LastModifiedDate),
		"SystemModstamp":                formatSFTime(c.SystemModstamp),
		"IsDeleted":                     c.IsDeleted,
	}
}

func customObjectToMap(version string, o gen.SalesforceCustomObjectC) map[string]any {
	return map[string]any{
		"attributes":         makeAttributes(version, "Custom_Object__c", o.ID),
		"Id":                 o.ID,
		"External_ID__c":     o.ExternalIDC,
		"Name":               o.Name,
		"Field_To_Update__c": o.FieldToUpdateC,
		"Description__c":     o.DescriptionC,
		"Status__c":          o.StatusC,
		"OwnerId":            o.OwnerID,
		"CreatedById":        o.CreatedByID,
		"CreatedDate":        formatSFTime(o.CreatedDate),
		"LastModifiedById":   o.LastModifiedByID,
		"LastModifiedDate":   formatSFTime(o.LastModifiedDate),
		"SystemModstamp":     formatSFTime(o.SystemModstamp),
		"IsDeleted":          o.IsDeleted,
	}
}

func contentDocumentToMap(version string, d gen.SalesforceContentDocument) map[string]any {
	return map[string]any{
		"attributes":  makeAttributes(version, "ContentDocument", d.ID),
		"Id":          d.ID,
		"Type":        d.Type,
		"Title":       d.Title,
		"CreatedDate": formatSFTime(d.CreatedDate),
		"IsDeleted":   d.IsDeleted,
	}
}

func contentVersionToMap(version string, v gen.SalesforceContentVersion) map[string]any {
	return map[string]any{
		"attributes":        makeAttributes(version, "ContentVersion", v.ID),
		"Id":                v.ID,
		"ContentDocumentId": v.ContentDocumentID,
		"Title":             v.Title,
		"PathOnClient":      v.PathOnClient,
		"VersionNumber":     v.VersionNumber,
		"FileType":          v.FileType,
		"VersionData":       fmt.Sprintf("/services/data/v%s/sobjects/ContentVersion/%s/VersionData", strings.TrimPrefix(version, "v"), v.ID),
		"CreatedDate":       formatSFTimePtr(v.CreatedDate),
		"IsDeleted":         v.IsDeleted,
	}
}

func contentDocumentLinkToMap(version string, l gen.SalesforceContentDocumentLink) map[string]any {
	return map[string]any{
		"attributes":        makeAttributes(version, "ContentDocumentLink", l.ID),
		"Id":                l.ID,
		"ContentDocumentId": l.ContentDocumentID,
		"LinkedEntityId":    l.LinkedEntityID,
		"ShareType":         l.ShareType,
		"Visibility":        l.Visibility,
		"IsDeleted":         l.IsDeleted,
	}
}

func (r *Repository) CreateRecord(ctx context.Context, sobject string, fields map[string]any) (*domain.SaveResult, map[string]any, error) {
	switch strings.ToLower(sobject) {
	case "account":
		name := getString(fields, "Name", "")
		if name == "" {
			return nil, nil, domain.ErrRequiredFieldMissing
		}
		acc, err := r.queries.CreateAccount(ctx, gen.CreateAccountParams{
			ExternalIDC:        getStringPtr(fields, "External_ID__c"),
			Name:               name,
			Type:               getStringPtr(fields, "Type"),
			ParentID:           getStringPtr(fields, "ParentId"),
			BillingStreet:      getStringPtr(fields, "BillingStreet"),
			BillingCity:        getStringPtr(fields, "BillingCity"),
			BillingState:       getStringPtr(fields, "BillingState"),
			BillingPostalCode:  getStringPtr(fields, "BillingPostalCode"),
			BillingCountry:     getStringPtr(fields, "BillingCountry"),
			ShippingStreet:     getStringPtr(fields, "ShippingStreet"),
			ShippingCity:       getStringPtr(fields, "ShippingCity"),
			ShippingState:      getStringPtr(fields, "ShippingState"),
			ShippingPostalCode: getStringPtr(fields, "ShippingPostalCode"),
			ShippingCountry:    getStringPtr(fields, "ShippingCountry"),
			Phone:              getStringPtr(fields, "Phone"),
			Fax:                getStringPtr(fields, "Fax"),
			AccountNumber:      getStringPtr(fields, "AccountNumber"),
			Website:            getStringPtr(fields, "Website"),
			Industry:           getStringPtr(fields, "Industry"),
			AnnualRevenue:      getFloatPtr(fields, "AnnualRevenue"),
			NumberOfEmployees:  getIntPtr(fields, "NumberOfEmployees"),
			Ownership:          getStringPtr(fields, "Ownership"),
			TickerSymbol:       getStringPtr(fields, "TickerSymbol"),
			Description:        getStringPtr(fields, "Description"),
			Rating:             getStringPtr(fields, "Rating"),
			Site:               getStringPtr(fields, "Site"),
		})
		if err != nil {
			return nil, nil, err
		}
		return &domain.SaveResult{ID: acc.ID, Success: true, Created: true, Errors: []string{}}, accountToMap("61.0", acc), nil

	case "contact":
		lastName := getString(fields, "LastName", "")
		if lastName == "" {
			return nil, nil, domain.ErrRequiredFieldMissing
		}
		firstName := getStringPtr(fields, "FirstName")
		fullName := getString(fields, "Name", "")
		if fullName == "" {
			if firstName != nil && *firstName != "" {
				fullName = *firstName + " " + lastName
			} else {
				fullName = lastName
			}
		}
		con, err := r.queries.CreateContact(ctx, gen.CreateContactParams{
			ExternalIDC:       getStringPtr(fields, "External_ID__c"),
			AccountID:         getStringPtr(fields, "AccountId"),
			LastName:          lastName,
			FirstName:         firstName,
			Salutation:        getStringPtr(fields, "Salutation"),
			Name:              fullName,
			MailingStreet:     getStringPtr(fields, "MailingStreet"),
			MailingCity:       getStringPtr(fields, "MailingCity"),
			MailingState:      getStringPtr(fields, "MailingState"),
			MailingPostalCode: getStringPtr(fields, "MailingPostalCode"),
			MailingCountry:    getStringPtr(fields, "MailingCountry"),
			Phone:             getStringPtr(fields, "Phone"),
			Fax:               getStringPtr(fields, "Fax"),
			MobilePhone:       getStringPtr(fields, "MobilePhone"),
			Email:             getStringPtr(fields, "Email"),
			Title:             getStringPtr(fields, "Title"),
			Department:        getStringPtr(fields, "Department"),
			LeadSource:        getStringPtr(fields, "LeadSource"),
			Description:       getStringPtr(fields, "Description"),
		})
		if err != nil {
			return nil, nil, err
		}
		return &domain.SaveResult{ID: con.ID, Success: true, Created: true, Errors: []string{}}, contactToMap("61.0", con), nil

	case "lead":
		lastName := getString(fields, "LastName", "")
		if lastName == "" {
			return nil, nil, domain.ErrRequiredFieldMissing
		}
		company := getString(fields, "Company", "Default Company")
		status := getString(fields, "Status", "Open - Not Contacted")
		firstName := getStringPtr(fields, "FirstName")
		fullName := getString(fields, "Name", "")
		if fullName == "" {
			if firstName != nil && *firstName != "" {
				fullName = *firstName + " " + lastName
			} else {
				fullName = lastName
			}
		}
		lead, err := r.queries.CreateLead(ctx, gen.CreateLeadParams{
			ExternalIDC:       getStringPtr(fields, "External_ID__c"),
			LastName:          lastName,
			FirstName:         firstName,
			Salutation:        getStringPtr(fields, "Salutation"),
			Name:              fullName,
			Title:             getStringPtr(fields, "Title"),
			Company:           company,
			Street:            getStringPtr(fields, "Street"),
			City:              getStringPtr(fields, "City"),
			State:             getStringPtr(fields, "State"),
			PostalCode:        getStringPtr(fields, "PostalCode"),
			Country:           getStringPtr(fields, "Country"),
			Phone:             getStringPtr(fields, "Phone"),
			MobilePhone:       getStringPtr(fields, "MobilePhone"),
			Email:             getStringPtr(fields, "Email"),
			Website:           getStringPtr(fields, "Website"),
			Description:       getStringPtr(fields, "Description"),
			LeadSource:        getStringPtr(fields, "LeadSource"),
			Status:            status,
			Industry:          getStringPtr(fields, "Industry"),
			Rating:            getStringPtr(fields, "Rating"),
			AnnualRevenue:     getFloatPtr(fields, "AnnualRevenue"),
			NumberOfEmployees: getIntPtr(fields, "NumberOfEmployees"),
		})
		if err != nil {
			return nil, nil, err
		}
		return &domain.SaveResult{ID: lead.ID, Success: true, Created: true, Errors: []string{}}, leadToMap("61.0", lead), nil

	case "case":
		priority := getString(fields, "Priority", "Medium")
		cs, err := r.queries.CreateCase(ctx, gen.CreateCaseParams{
			ExternalIDC:              getStringPtr(fields, "External_ID__c"),
			AdditionalInfoC:          getStringPtr(fields, "AdditionalInfo__c"),
			SubTypeC:                 getStringPtr(fields, "Sub_Type__c"),
			TypeC:                    getStringPtr(fields, "Type__c"),
			TypeName:                 getStringPtr(fields, "Type"),
			Origin:                   getStringPtr(fields, "Origin"),
			RecordTypeID:             getStringPtr(fields, "RecordTypeId"),
			ParentID:                 getStringPtr(fields, "ParentId"),
			PopulationImpactedC:      getStringPtr(fields, "Population_Impacted__c"),
			ClosedDate:               getTimePtr(fields, "ClosedDate"),
			WeblinksC:                getStringPtr(fields, "Weblinks__c"),
			RootCauseC:               getStringPtr(fields, "Root_Cause__c"),
			ImpactC:                  getStringPtr(fields, "Impact__c"),
			VipC:                     getBoolPtr(fields, "VIP__c"),
			Priority:                 priority,
			AccountID:                getStringPtr(fields, "AccountId"),
			ContactID:                getStringPtr(fields, "ContactId"),
			RecurrenceC:              getStringPtr(fields, "Recurrence__c"),
			HumanFactorC:             getStringPtr(fields, "Human_Factor__c"),
			TechnicalSystemC:         getStringPtr(fields, "Technical_System__c"),
			ContactConclusionLevel2C: getStringPtr(fields, "Contact_Conclusion_Level_2__c"),
			ContactConclusionLevel1C: getStringPtr(fields, "Contact_Conclusion_Level_1__c"),
			ContactActionLevel3C:     getStringPtr(fields, "Contact_Action_Level_3__c"),
			ContactActionLevel2C:     getStringPtr(fields, "Contact_Action_Level_2__c"),
			ContactActionLevel1C:     getStringPtr(fields, "Contact_Action_Level_1__c"),
			ContactTopicLevel3C:      getStringPtr(fields, "Contact_Topic_Level_3__c"),
			ContactTopicLevel2C:      getStringPtr(fields, "Contact_Topic_Level_2__c"),
			ContactTopicLevel1C:      getStringPtr(fields, "Contact_Topic_Level_1__c"),
			ComponentsC:              getStringPtr(fields, "Components__c"),
			StepsToReproduceC:        getStringPtr(fields, "Steps_To_Reproduce__c"),
			ExpectedResultC:          getStringPtr(fields, "Expected_Result__c"),
			ActualResultC:            getStringPtr(fields, "Actual_Result__c"),
			Description:              getStringPtr(fields, "Description"),
			OwnerID:                  getStringPtr(fields, "OwnerId"),
			SubStatusLevel3C:         getStringPtr(fields, "Sub_Status_Level_3__c"),
			SubStatusLevel2C:         getStringPtr(fields, "Sub_Status_Level_2__c"),
			SubStatusC:               getStringPtr(fields, "Sub_Status__c"),
			Status:                   getStringPtr(fields, "Status"),
			CreatedByID:              getStringPtr(fields, "CreatedById"),
			Subject:                  getStringPtr(fields, "Subject"),
		})
		if err != nil {
			return nil, nil, err
		}
		return &domain.SaveResult{ID: cs.ID, Success: true, Created: true, Errors: []string{}}, caseToMap("61.0", cs), nil

	case "custom_object__c":
		name := getString(fields, "Name", "Custom Record")
		co, err := r.queries.CreateCustomObject(ctx, gen.CreateCustomObjectParams{
			ExternalIDC:    getStringPtr(fields, "External_ID__c"),
			Name:           name,
			FieldToUpdateC: getStringPtr(fields, "Field_To_Update__c"),
			DescriptionC:   getStringPtr(fields, "Description__c"),
			StatusC:        getStringPtr(fields, "Status__c"),
		})
		if err != nil {
			return nil, nil, err
		}
		return &domain.SaveResult{ID: co.ID, Success: true, Created: true, Errors: []string{}}, customObjectToMap("61.0", co), nil

	case "contentdocument":
		doc, err := r.queries.CreateContentDocument(ctx, gen.CreateContentDocumentParams{
			Type:  getStringPtr(fields, "Type"),
			Title: getStringPtr(fields, "Title"),
		})
		if err != nil {
			return nil, nil, err
		}
		return &domain.SaveResult{ID: doc.ID, Success: true, Created: true, Errors: []string{}}, contentDocumentToMap("61.0", doc), nil

	case "contentversion":
		title := getString(fields, "Title", "Untitled")
		pathOnClient := getString(fields, "PathOnClient", title)
		versionData := getString(fields, "VersionData", "")
		docID := getString(fields, "ContentDocumentId", "")
		if docID == "" {
			doc, err := r.queries.CreateContentDocument(ctx, gen.CreateContentDocumentParams{
				Title: &title,
			})
			if err != nil {
				return nil, nil, err
			}
			docID = doc.ID
		}
		cv, err := r.queries.CreateContentVersion(ctx, gen.CreateContentVersionParams{
			PathOnClient:      pathOnClient,
			Title:             title,
			VersionData:       versionData,
			ContentDocumentID: docID,
			VersionDataUrl:    getStringPtr(fields, "VersionDataUrl"),
			FileType:          getStringPtr(fields, "FileType"),
		})
		if err != nil {
			return nil, nil, err
		}
		return &domain.SaveResult{ID: cv.ID, Success: true, Created: true, Errors: []string{}}, contentVersionToMap("61.0", cv), nil

	case "contentdocumentlink":
		cdl, err := r.queries.CreateContentDocumentLink(ctx, gen.CreateContentDocumentLinkParams{
			ContentDocumentID: getString(fields, "ContentDocumentId", ""),
			LinkedEntityID:    getString(fields, "LinkedEntityId", ""),
			ShareType:         getString(fields, "ShareType", "V"),
			Visibility:        getString(fields, "Visibility", "AllUsers"),
		})
		if err != nil {
			return nil, nil, err
		}
		return &domain.SaveResult{ID: cdl.ID, Success: true, Created: true, Errors: []string{}}, contentDocumentLinkToMap("61.0", cdl), nil

	default:
		if strings.HasSuffix(strings.ToLower(sobject), "__e") {
			payloadBytes, _ := json.Marshal(fields)
			channel := "/event/" + sobject
			ev, err := r.queries.CreatePlatformEvent(ctx, gen.CreatePlatformEventParams{
				ChannelName: channel,
				EventType:   sobject,
				Payload:     string(payloadBytes),
			})
			if err != nil {
				return nil, nil, err
			}
			m := map[string]any{
				"Id":          ev.ID,
				"ReplayId":    ev.ReplayID,
				"ChannelName": ev.ChannelName,
				"EventType":   ev.EventType,
			}
			return &domain.SaveResult{ID: ev.ID, Success: true, Created: true, Errors: []string{}}, m, nil
		}
		return nil, nil, domain.ErrInvalidSObject
	}
}

func (r *Repository) GetRecord(ctx context.Context, version, sobject, id string) (map[string]any, error) {
	switch strings.ToLower(sobject) {
	case "account":
		a, err := r.queries.GetAccountByID(ctx, id)
		if err != nil {
			return nil, domain.ErrNotFound
		}
		return accountToMap(version, a), nil
	case "contact":
		c, err := r.queries.GetContactByID(ctx, id)
		if err != nil {
			return nil, domain.ErrNotFound
		}
		return contactToMap(version, c), nil
	case "lead":
		l, err := r.queries.GetLeadByID(ctx, id)
		if err != nil {
			return nil, domain.ErrNotFound
		}
		return leadToMap(version, l), nil
	case "case":
		c, err := r.queries.SelectCase(ctx, id)
		if err != nil {
			return nil, domain.ErrNotFound
		}
		return caseToMap(version, c), nil
	case "custom_object__c":
		o, err := r.queries.GetCustomObjectByID(ctx, id)
		if err != nil {
			return nil, domain.ErrNotFound
		}
		return customObjectToMap(version, o), nil
	case "contentdocument":
		d, err := r.queries.GetContentDocument(ctx, id)
		if err != nil {
			return nil, domain.ErrNotFound
		}
		return contentDocumentToMap(version, d), nil
	case "contentversion":
		v, err := r.queries.GetContentVersion(ctx, id)
		if err != nil {
			return nil, domain.ErrNotFound
		}
		return contentVersionToMap(version, v), nil
	case "contentdocumentlink":
		l, err := r.queries.GetContentDocumentLink(ctx, id)
		if err != nil {
			return nil, domain.ErrNotFound
		}
		return contentDocumentLinkToMap(version, l), nil
	default:
		return nil, domain.ErrInvalidSObject
	}
}

func (r *Repository) UpdateRecord(ctx context.Context, sobject, id string, fields map[string]any) (map[string]any, error) {
	switch strings.ToLower(sobject) {
	case "account":
		a, err := r.queries.UpdateAccount(ctx, gen.UpdateAccountParams{
			ExternalIDC:       getStringPtr(fields, "External_ID__c"),
			Name:              getStringPtr(fields, "Name"),
			Type:              getStringPtr(fields, "Type"),
			ParentID:          getStringPtr(fields, "ParentId"),
			BillingStreet:     getStringPtr(fields, "BillingStreet"),
			BillingCity:       getStringPtr(fields, "BillingCity"),
			BillingState:      getStringPtr(fields, "BillingState"),
			BillingPostalCode: getStringPtr(fields, "BillingPostalCode"),
			BillingCountry:    getStringPtr(fields, "BillingCountry"),
			Phone:             getStringPtr(fields, "Phone"),
			Website:           getStringPtr(fields, "Website"),
			Industry:          getStringPtr(fields, "Industry"),
			AnnualRevenue:     getFloatPtr(fields, "AnnualRevenue"),
			NumberOfEmployees: getIntPtr(fields, "NumberOfEmployees"),
			Description:       getStringPtr(fields, "Description"),
			Rating:            getStringPtr(fields, "Rating"),
			Site:              getStringPtr(fields, "Site"),
			ID:                id,
		})
		if err != nil {
			return nil, domain.ErrNotFound
		}
		return accountToMap("61.0", a), nil

	case "contact":
		c, err := r.queries.UpdateContact(ctx, gen.UpdateContactParams{
			ExternalIDC:       getStringPtr(fields, "External_ID__c"),
			AccountID:         getStringPtr(fields, "AccountId"),
			LastName:          getStringPtr(fields, "LastName"),
			FirstName:         getStringPtr(fields, "FirstName"),
			Name:              getStringPtr(fields, "Name"),
			MailingStreet:     getStringPtr(fields, "MailingStreet"),
			MailingCity:       getStringPtr(fields, "MailingCity"),
			MailingState:      getStringPtr(fields, "MailingState"),
			MailingPostalCode: getStringPtr(fields, "MailingPostalCode"),
			MailingCountry:    getStringPtr(fields, "MailingCountry"),
			Phone:             getStringPtr(fields, "Phone"),
			MobilePhone:       getStringPtr(fields, "MobilePhone"),
			Email:             getStringPtr(fields, "Email"),
			Title:             getStringPtr(fields, "Title"),
			Department:        getStringPtr(fields, "Department"),
			Description:       getStringPtr(fields, "Description"),
			ID:                id,
		})
		if err != nil {
			return nil, domain.ErrNotFound
		}
		return contactToMap("61.0", c), nil

	case "lead":
		l, err := r.queries.UpdateLead(ctx, gen.UpdateLeadParams{
			ExternalIDC: getStringPtr(fields, "External_ID__c"),
			LastName:    getStringPtr(fields, "LastName"),
			FirstName:   getStringPtr(fields, "FirstName"),
			Name:        getStringPtr(fields, "Name"),
			Title:       getStringPtr(fields, "Title"),
			Company:     getStringPtr(fields, "Company"),
			Phone:       getStringPtr(fields, "Phone"),
			Email:       getStringPtr(fields, "Email"),
			Website:     getStringPtr(fields, "Website"),
			Description: getStringPtr(fields, "Description"),
			Status:      getStringPtr(fields, "Status"),
			Industry:    getStringPtr(fields, "Industry"),
			Rating:      getStringPtr(fields, "Rating"),
			ID:          id,
		})
		if err != nil {
			return nil, domain.ErrNotFound
		}
		return leadToMap("61.0", l), nil

	case "case":
		c, err := r.queries.UpdateCase(ctx, gen.UpdateCaseParams{
			ExternalIDC:         getStringPtr(fields, "External_ID__c"),
			AdditionalInfoC:     getStringPtr(fields, "AdditionalInfo__c"),
			SubTypeC:            getStringPtr(fields, "Sub_Type__c"),
			TypeC:               getStringPtr(fields, "Type__c"),
			TypeName:            getStringPtr(fields, "Type"),
			Origin:              getStringPtr(fields, "Origin"),
			RecordTypeID:        getStringPtr(fields, "RecordTypeId"),
			ParentID:            getStringPtr(fields, "ParentId"),
			PopulationImpactedC: getStringPtr(fields, "Population_Impacted__c"),
			ClosedDate:          getTimePtr(fields, "ClosedDate"),
			WeblinksC:           getStringPtr(fields, "Weblinks__c"),
			RootCauseC:          getStringPtr(fields, "Root_Cause__c"),
			ImpactC:             getStringPtr(fields, "Impact__c"),
			VipC:                getBoolPtr(fields, "VIP__c"),
			Priority:            getStringPtr(fields, "Priority"),
			AccountID:           getStringPtr(fields, "AccountId"),
			ContactID:           getStringPtr(fields, "ContactId"),
			Description:         getStringPtr(fields, "Description"),
			Status:              getStringPtr(fields, "Status"),
			Subject:             getStringPtr(fields, "Subject"),
			ID:                  id,
		})
		if err != nil {
			return nil, domain.ErrNotFound
		}
		return caseToMap("61.0", c), nil

	case "custom_object__c":
		o, err := r.queries.UpdateCustomObject(ctx, gen.UpdateCustomObjectParams{
			ExternalIDC:    getStringPtr(fields, "External_ID__c"),
			Name:           getStringPtr(fields, "Name"),
			FieldToUpdateC: getStringPtr(fields, "Field_To_Update__c"),
			DescriptionC:   getStringPtr(fields, "Description__c"),
			StatusC:        getStringPtr(fields, "Status__c"),
			ID:             id,
		})
		if err != nil {
			return nil, domain.ErrNotFound
		}
		return customObjectToMap("61.0", o), nil

	case "contentdocument":
		d, err := r.queries.UpdateContentDocument(ctx, gen.UpdateContentDocumentParams{
			Type:  getStringPtr(fields, "Type"),
			Title: getStringPtr(fields, "Title"),
			ID:    id,
		})
		if err != nil {
			return nil, domain.ErrNotFound
		}
		return contentDocumentToMap("61.0", d), nil

	default:
		return nil, domain.ErrInvalidSObject
	}
}

func (r *Repository) UpsertRecordByExternalID(ctx context.Context, sobject, externalIDField, externalIDValue string, fields map[string]any) (*domain.SaveResult, error) {
	merged := make(map[string]any, len(fields)+1)
	for k, v := range fields {
		merged[k] = v
	}
	if strings.EqualFold(externalIDField, "Id") {
		if _, err := r.UpdateRecord(ctx, sobject, externalIDValue, merged); err == nil {
			return &domain.SaveResult{ID: externalIDValue, Success: true, Created: false, Errors: []string{}}, nil
		}
		res, _, err := r.CreateRecord(ctx, sobject, merged)
		return res, err
	}

	merged["External_ID__c"] = externalIDValue
	var existingID string
	switch strings.ToLower(sobject) {
	case "account":
		if acc, err := r.queries.GetAccountByExternalID(ctx, externalIDValue); err == nil {
			existingID = acc.ID
		}
	case "contact":
		if con, err := r.queries.GetContactByExternalID(ctx, externalIDValue); err == nil {
			existingID = con.ID
		}
	case "lead":
		if l, err := r.queries.GetLeadByExternalID(ctx, externalIDValue); err == nil {
			existingID = l.ID
		}
	case "case":
		if c, err := r.queries.GetCaseByExternalID(ctx, externalIDValue); err == nil {
			existingID = c.ID
		}
	case "custom_object__c":
		if co, err := r.queries.GetCustomObjectByExternalID(ctx, externalIDValue); err == nil {
			existingID = co.ID
		}
	default:
		return nil, domain.ErrInvalidSObject
	}

	if existingID != "" {
		if _, err := r.UpdateRecord(ctx, sobject, existingID, merged); err != nil {
			return nil, err
		}
		return &domain.SaveResult{ID: existingID, Success: true, Created: false, Errors: []string{}}, nil
	}

	res, _, err := r.CreateRecord(ctx, sobject, merged)
	return res, err
}

func (r *Repository) DeleteRecord(ctx context.Context, sobject, id string) error {
	var rows int64
	var err error
	switch strings.ToLower(sobject) {
	case "account":
		rows, err = r.queries.SoftDeleteAccount(ctx, id)
	case "contact":
		rows, err = r.queries.SoftDeleteContact(ctx, id)
	case "lead":
		rows, err = r.queries.SoftDeleteLead(ctx, id)
	case "case":
		rows, err = r.queries.SoftDeleteCase(ctx, id)
	case "custom_object__c":
		rows, err = r.queries.SoftDeleteCustomObject(ctx, id)
	case "contentdocument":
		rows, err = r.queries.SoftDeleteContentDocument(ctx, id)
	case "contentversion":
		rows, err = r.queries.SoftDeleteContentVersion(ctx, id)
	case "contentdocumentlink":
		rows, err = r.queries.SoftDeleteContentDocumentLink(ctx, id)
	default:
		return domain.ErrInvalidSObject
	}
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) GetContentVersionData(ctx context.Context, id string) ([]byte, error) {
	cv, err := r.queries.GetContentVersion(ctx, id)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	decoded, err := base64.StdEncoding.DecodeString(cv.VersionData)
	if err != nil {
		return []byte(cv.VersionData), nil
	}
	return decoded, nil
}

func (r *Repository) ListRecordsForSObject(ctx context.Context, version, sobject string, includeDeleted bool) ([]map[string]any, error) {
	switch strings.ToLower(sobject) {
	case "account":
		var rows []gen.SalesforceAccount
		var err error
		if includeDeleted {
			rows, err = r.queries.ListAccountsAll(ctx)
		} else {
			rows, err = r.queries.ListAccounts(ctx)
		}
		if err != nil {
			return nil, err
		}
		res := make([]map[string]any, 0, len(rows))
		for _, item := range rows {
			res = append(res, accountToMap(version, item))
		}
		return res, nil

	case "contact":
		var rows []gen.SalesforceContact
		var err error
		if includeDeleted {
			rows, err = r.queries.ListContactsAll(ctx)
		} else {
			rows, err = r.queries.ListContacts(ctx)
		}
		if err != nil {
			return nil, err
		}
		res := make([]map[string]any, 0, len(rows))
		for _, item := range rows {
			res = append(res, contactToMap(version, item))
		}
		return res, nil

	case "lead":
		var rows []gen.SalesforceLead
		var err error
		if includeDeleted {
			rows, err = r.queries.ListLeadsAll(ctx)
		} else {
			rows, err = r.queries.ListLeads(ctx)
		}
		if err != nil {
			return nil, err
		}
		res := make([]map[string]any, 0, len(rows))
		for _, item := range rows {
			res = append(res, leadToMap(version, item))
		}
		return res, nil

	case "case":
		var rows []gen.SalesforceCase
		var err error
		if includeDeleted {
			rows, err = r.queries.ListCasesAll(ctx)
		} else {
			rows, err = r.queries.ListCases(ctx)
		}
		if err != nil {
			return nil, err
		}
		res := make([]map[string]any, 0, len(rows))
		for _, item := range rows {
			res = append(res, caseToMap(version, item))
		}
		return res, nil

	case "custom_object__c":
		var rows []gen.SalesforceCustomObjectC
		var err error
		if includeDeleted {
			rows, err = r.queries.ListCustomObjectsAll(ctx)
		} else {
			rows, err = r.queries.ListCustomObjects(ctx)
		}
		if err != nil {
			return nil, err
		}
		res := make([]map[string]any, 0, len(rows))
		for _, item := range rows {
			res = append(res, customObjectToMap(version, item))
		}
		return res, nil

	case "contentdocument":
		rows, err := r.queries.ListContentDocuments(ctx)
		if err != nil {
			return nil, err
		}
		res := make([]map[string]any, 0, len(rows))
		for _, item := range rows {
			res = append(res, contentDocumentToMap(version, item))
		}
		return res, nil

	case "contentversion":
		rows, err := r.queries.ListContentVersions(ctx)
		if err != nil {
			return nil, err
		}
		res := make([]map[string]any, 0, len(rows))
		for _, item := range rows {
			res = append(res, contentVersionToMap(version, item))
		}
		return res, nil

	case "contentdocumentlink":
		rows, err := r.queries.ListContentDocumentLinks(ctx)
		if err != nil {
			return nil, err
		}
		res := make([]map[string]any, 0, len(rows))
		for _, item := range rows {
			res = append(res, contentDocumentLinkToMap(version, item))
		}
		return res, nil

	default:
		return nil, domain.ErrInvalidSObject
	}
}

func (r *Repository) SearchRecordsForSObject(ctx context.Context, version, sobject, term string) ([]map[string]any, error) {
	switch strings.ToLower(sobject) {
	case "account":
		rows, err := r.queries.SearchAccounts(ctx, term)
		if err != nil {
			return nil, err
		}
		res := make([]map[string]any, 0, len(rows))
		for _, item := range rows {
			res = append(res, accountToMap(version, item))
		}
		return res, nil
	case "contact":
		rows, err := r.queries.SearchContacts(ctx, term)
		if err != nil {
			return nil, err
		}
		res := make([]map[string]any, 0, len(rows))
		for _, item := range rows {
			res = append(res, contactToMap(version, item))
		}
		return res, nil
	case "lead":
		rows, err := r.queries.SearchLeads(ctx, term)
		if err != nil {
			return nil, err
		}
		res := make([]map[string]any, 0, len(rows))
		for _, item := range rows {
			res = append(res, leadToMap(version, item))
		}
		return res, nil
	case "case":
		rows, err := r.queries.SearchCases(ctx, term)
		if err != nil {
			return nil, err
		}
		res := make([]map[string]any, 0, len(rows))
		for _, item := range rows {
			res = append(res, caseToMap(version, item))
		}
		return res, nil
	case "custom_object__c":
		rows, err := r.queries.SearchCustomObjects(ctx, term)
		if err != nil {
			return nil, err
		}
		res := make([]map[string]any, 0, len(rows))
		for _, item := range rows {
			res = append(res, customObjectToMap(version, item))
		}
		return res, nil
	default:
		return nil, domain.ErrInvalidSObject
	}
}
