package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/faber-numeris/sem/internal/core/domain"
)

func (s *Service) GetLimits(ctx context.Context, version string) (map[string]domain.OrgLimit, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, err
	}
	limits, err := s.repo.ListOrgLimits(ctx)
	if err != nil {
		return nil, err
	}
	res := make(map[string]domain.OrgLimit, len(limits))
	for _, l := range limits {
		res[l.Name] = l
	}
	return res, nil
}

func formatDefSummary(version string, d domain.SObjectDefinition, isTooling bool) map[string]any {
	v := normalizeVersion(version)
	prefix := fmt.Sprintf("/services/data/v%s/sobjects/%s", v, d.Name)
	if isTooling {
		prefix = fmt.Sprintf("/services/data/v%s/tooling/sobjects/%s", v, d.Name)
	}
	return map[string]any{
		"activateable":        false,
		"createable":          d.Createable,
		"custom":              d.Custom,
		"customSetting":       false,
		"deepCloneable":       false,
		"deletable":           d.Deletable,
		"deprecatedAndHidden": false,
		"feedEnabled":         false,
		"hasSubtypes":         false,
		"isInterface":         false,
		"isSubtype":           false,
		"keyPrefix":           d.KeyPrefix,
		"label":               d.Label,
		"labelPlural":         d.LabelPlural,
		"layoutable":          true,
		"mergeable":           false,
		"mruEnabled":          true,
		"name":                d.Name,
		"queryable":           d.Queryable,
		"replicateable":       d.Replicateable,
		"retrieveable":        d.Retrieveable,
		"searchable":          d.Searchable,
		"triggerable":         d.Triggerable,
		"undeletable":         d.Undeletable,
		"updateable":          d.Updateable,
		"urls": map[string]string{
			"sobject":  prefix,
			"describe": prefix + "/describe",
			"rowTemplate": prefix + "/{ID}",
		},
	}
}

func (s *Service) DescribeGlobal(ctx context.Context, version string, isTooling bool) (map[string]any, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, err
	}
	defs, err := s.repo.ListSObjectDefinitions(ctx, isTooling)
	if err != nil {
		return nil, err
	}
	sobjects := make([]map[string]any, 0, len(defs))
	for _, d := range defs {
		sobjects = append(sobjects, formatDefSummary(version, d, isTooling))
	}
	return map[string]any{
		"encoding":     "UTF-8",
		"maxBatchSize": 200,
		"sobjects":     sobjects,
	}, nil
}

func (s *Service) GetSObjectBasicInfo(ctx context.Context, version, sobject string) (map[string]any, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, err
	}
	desc, err := s.repo.GetSObjectDescribe(ctx, sobject)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"objectDescribe": formatDefSummary(version, desc.Definition, desc.Definition.IsTooling),
		"recentItems":    []any{},
	}, nil
}

func (s *Service) DescribeSObject(ctx context.Context, version, sobject string, isTooling bool) (map[string]any, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, err
	}
	desc, err := s.repo.GetSObjectDescribe(ctx, sobject)
	if err != nil {
		return nil, err
	}
	out := formatDefSummary(version, desc.Definition, isTooling)
	fields := make([]map[string]any, 0, len(desc.Fields))
	for _, f := range desc.Fields {
		refTo := []string{}
		if f.ReferenceTo != "" {
			refTo = []string{f.ReferenceTo}
		}
		fields = append(fields, map[string]any{
			"name":        f.Name,
			"label":       f.Label,
			"type":        f.FieldType,
			"soapType":    f.SoapType,
			"length":      f.Length,
			"nillable":    f.Nillable,
			"createable":  f.Createable,
			"updateable":  f.Updateable,
			"filterable":  f.Filterable,
			"sortable":    f.Sortable,
			"custom":      f.Custom,
			"externalId":  f.ExternalID,
			"unique":      f.UniqueField,
			"nameField":   f.NameField,
			"referenceTo": refTo,
		})
	}
	out["fields"] = fields
	out["childRelationships"] = []any{}
	out["recordTypeInfos"] = []any{}
	return out, nil
}

func (s *Service) CreateRecord(ctx context.Context, version, sobject string, body map[string]any) (*domain.SaveResult, map[string]any, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, nil, err
	}
	return s.repo.CreateRecord(ctx, sobject, body)
}

func (s *Service) GetRecord(ctx context.Context, version, sobject, id string, fields []string) (map[string]any, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, err
	}
	rec, err := s.repo.GetRecord(ctx, normalizeVersion(version), sobject, id)
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return rec, nil
	}
	filtered := map[string]any{
		"attributes": rec["attributes"],
		"Id":         rec["Id"],
	}
	for _, reqField := range fields {
		f := strings.TrimSpace(reqField)
		if f == "" {
			continue
		}
		for k, v := range rec {
			if strings.EqualFold(k, f) {
				filtered[k] = v
				break
			}
		}
	}
	return filtered, nil
}

func (s *Service) UpdateRecord(ctx context.Context, version, sobject, id string, body map[string]any) (map[string]any, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, err
	}
	return s.repo.UpdateRecord(ctx, sobject, id, body)
}

func (s *Service) UpsertRecord(ctx context.Context, version, sobject, externalIDField, externalIDValue string, body map[string]any) (*domain.SaveResult, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, err
	}
	return s.repo.UpsertRecordByExternalID(ctx, sobject, externalIDField, externalIDValue, body)
}

func (s *Service) DeleteRecord(ctx context.Context, version, sobject, id string) error {
	if err := s.ValidateVersion(version); err != nil {
		return err
	}
	return s.repo.DeleteRecord(ctx, sobject, id)
}

func (s *Service) GetContentVersionData(ctx context.Context, version, id string) ([]byte, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, err
	}
	return s.repo.GetContentVersionData(ctx, id)
}
