package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/faber-numeris/sem/internal/adapters/outbound/postgres/gen"
	"github.com/faber-numeris/sem/internal/core/domain"
)

func makeToolingAttributes(version, sobject, id string) domain.RecordAttributes {
	v := strings.TrimPrefix(version, "v")
	if v == "" {
		v = "61.0"
	}
	return domain.RecordAttributes{
		Type: sobject,
		URL:  fmt.Sprintf("/services/data/v%s/tooling/sobjects/%s/%s", v, sobject, id),
	}
}

func (r *Repository) ListToolingRecords(ctx context.Context, version, sobject string) ([]map[string]any, error) {
	switch strings.ToLower(sobject) {
	case "apexclass":
		rows, err := r.queries.ListApexClasses(ctx)
		if err != nil {
			return nil, err
		}
		res := make([]map[string]any, 0, len(rows))
		for _, item := range rows {
			res = append(res, map[string]any{
				"attributes":            makeToolingAttributes(version, "ApexClass", item.ID),
				"Id":                    item.ID,
				"Name":                  item.Name,
				"Body":                  item.Body,
				"ApiVersion":            item.ApiVersion,
				"Status":                item.Status,
				"IsValid":               item.IsValid,
				"LengthWithoutComments": item.LengthWithoutComments,
				"NamespacePrefix":       item.NamespacePrefix,
				"CreatedById":           item.CreatedByID,
				"CreatedDate":           formatSFTime(item.CreatedDate),
				"LastModifiedById":      item.LastModifiedByID,
				"LastModifiedDate":      formatSFTime(item.LastModifiedDate),
				"SystemModstamp":        formatSFTime(item.SystemModstamp),
			})
		}
		return res, nil

	case "apextrigger":
		rows, err := r.queries.ListApexTriggers(ctx)
		if err != nil {
			return nil, err
		}
		res := make([]map[string]any, 0, len(rows))
		for _, item := range rows {
			res = append(res, map[string]any{
				"attributes":        makeToolingAttributes(version, "ApexTrigger", item.ID),
				"Id":                item.ID,
				"Name":              item.Name,
				"TableEnumOrId":     item.TableEnumOrID,
				"Body":              item.Body,
				"ApiVersion":        item.ApiVersion,
				"Status":            item.Status,
				"IsValid":           item.IsValid,
				"UsageBeforeInsert": item.UsageBeforeInsert,
				"UsageAfterInsert":  item.UsageAfterInsert,
				"UsageBeforeUpdate": item.UsageBeforeUpdate,
				"UsageAfterUpdate":  item.UsageAfterUpdate,
				"CreatedById":       item.CreatedByID,
				"CreatedDate":       formatSFTime(item.CreatedDate),
				"LastModifiedDate":  formatSFTime(item.LastModifiedDate),
			})
		}
		return res, nil

	case "customfield":
		rows, err := r.queries.ListCustomFields(ctx)
		if err != nil {
			return nil, err
		}
		res := make([]map[string]any, 0, len(rows))
		for _, item := range rows {
			res = append(res, map[string]any{
				"attributes":       makeToolingAttributes(version, "CustomField", item.ID),
				"Id":               item.ID,
				"DeveloperName":    item.DeveloperName,
				"TableEnumOrId":    item.TableEnumOrID,
				"FullName":         item.FullName,
				"Label":            item.Label,
				"DataType":         item.DataType,
				"Length":           item.Length,
				"CreatedDate":      formatSFTime(item.CreatedDate),
				"LastModifiedDate": formatSFTime(item.LastModifiedDate),
			})
		}
		return res, nil

	case "validationrule":
		rows, err := r.queries.ListValidationRules(ctx)
		if err != nil {
			return nil, err
		}
		res := make([]map[string]any, 0, len(rows))
		for _, item := range rows {
			res = append(res, map[string]any{
				"attributes":            makeToolingAttributes(version, "ValidationRule", item.ID),
				"Id":                    item.ID,
				"ValidationName":        item.ValidationName,
				"EntityDefinitionId":    item.EntityDefinitionID,
				"Active":                item.Active,
				"Description":           item.Description,
				"ErrorConditionFormula": item.ErrorConditionFormula,
				"ErrorMessage":          item.ErrorMessage,
				"CreatedDate":           formatSFTime(item.CreatedDate),
				"LastModifiedDate":      formatSFTime(item.LastModifiedDate),
			})
		}
		return res, nil

	default:
		return nil, domain.ErrInvalidSObject
	}
}

func (r *Repository) CreateApexTestRun(ctx context.Context, classNames, testLevel string) (*domain.ApexTestRun, error) {
	created, err := r.queries.CreateApexTestRun(ctx, gen.CreateApexTestRunParams{
		ClassNames: classNames,
		TestLevel:  testLevel,
		Status:     "Completed",
	})
	if err != nil {
		return nil, err
	}
	return &domain.ApexTestRun{
		ID:               created.ID,
		ClassNames:       created.ClassNames,
		TestLevel:        created.TestLevel,
		Status:           created.Status,
		MethodsEnqueued:  created.MethodsEnqueued,
		MethodsCompleted: created.MethodsCompleted,
		MethodsFailed:    created.MethodsFailed,
		CreatedDate:      created.CreatedDate,
	}, nil
}

func (r *Repository) GetApexRestRoute(ctx context.Context, urlMapping, method string) (*domain.ApexRestRoute, error) {
	route, err := r.queries.GetApexRestRoute(ctx, gen.GetApexRestRouteParams{
		UrlMapping: urlMapping,
		HttpMethod: method,
	})
	if err != nil {
		return nil, domain.ErrNotFound
	}
	return &domain.ApexRestRoute{
		ID:             route.ID,
		URLMapping:     route.UrlMapping,
		HTTPMethod:     route.HttpMethod,
		ClassName:      route.ClassName,
		ResponseStatus: route.ResponseStatus,
		ResponseBody:   route.ResponseBody,
		IsActive:       route.IsActive,
	}, nil
}

func (r *Repository) CreatePlatformEvent(ctx context.Context, channel, eventType, payload string) (*domain.PlatformEvent, error) {
	ev, err := r.queries.CreatePlatformEvent(ctx, gen.CreatePlatformEventParams{
		ChannelName: channel,
		EventType:   eventType,
		Payload:     payload,
	})
	if err != nil {
		return nil, err
	}
	return &domain.PlatformEvent{
		ID:          ev.ID,
		ReplayID:    ev.ReplayID,
		ChannelName: ev.ChannelName,
		EventType:   ev.EventType,
		Payload:     ev.Payload,
		CreatedByID: ev.CreatedByID,
		CreatedDate: ev.CreatedDate,
	}, nil
}

func (r *Repository) ListPlatformEventsSince(ctx context.Context, channel string, afterReplayID int64) ([]domain.PlatformEvent, error) {
	rows, err := r.queries.ListPlatformEventsSince(ctx, gen.ListPlatformEventsSinceParams{
		ChannelName:   channel,
		AfterReplayID: afterReplayID,
	})
	if err != nil {
		return nil, err
	}
	res := make([]domain.PlatformEvent, 0, len(rows))
	for _, ev := range rows {
		res = append(res, domain.PlatformEvent{
			ID:          ev.ID,
			ReplayID:    ev.ReplayID,
			ChannelName: ev.ChannelName,
			EventType:   ev.EventType,
			Payload:     ev.Payload,
			CreatedByID: ev.CreatedByID,
			CreatedDate: ev.CreatedDate,
		})
	}
	return res, nil
}

func (r *Repository) UpsertCometDClient(ctx context.Context, clientID string) (*domain.CometDClient, error) {
	c, err := r.queries.UpsertCometDClient(ctx, clientID)
	if err != nil {
		return nil, err
	}
	return &domain.CometDClient{
		ClientID:    c.ClientID,
		IsConnected: c.IsConnected,
	}, nil
}

func (r *Repository) GetCometDClient(ctx context.Context, clientID string) (*domain.CometDClient, error) {
	c, err := r.queries.GetCometDClient(ctx, clientID)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	return &domain.CometDClient{
		ClientID:    c.ClientID,
		IsConnected: c.IsConnected,
	}, nil
}

func (r *Repository) DisconnectCometDClient(ctx context.Context, clientID string) error {
	return r.queries.DisconnectCometDClient(ctx, clientID)
}

func (r *Repository) UpsertCometDSubscription(ctx context.Context, clientID, channel string, replayID int64) (*domain.CometDSubscription, error) {
	s, err := r.queries.UpsertCometDSubscription(ctx, gen.UpsertCometDSubscriptionParams{
		ClientID:     clientID,
		ChannelName:  channel,
		LastReplayID: replayID,
	})
	if err != nil {
		return nil, err
	}
	return &domain.CometDSubscription{
		ID:           s.ID,
		ClientID:     s.ClientID,
		ChannelName:  s.ChannelName,
		LastReplayID: s.LastReplayID,
	}, nil
}

func (r *Repository) DeleteCometDSubscription(ctx context.Context, clientID, channel string) error {
	return r.queries.DeleteCometDSubscription(ctx, gen.DeleteCometDSubscriptionParams{
		ClientID:    clientID,
		ChannelName: channel,
	})
}

func (r *Repository) ListCometDSubscriptions(ctx context.Context, clientID string) ([]domain.CometDSubscription, error) {
	rows, err := r.queries.ListCometDSubscriptionsByClient(ctx, clientID)
	if err != nil {
		return nil, err
	}
	res := make([]domain.CometDSubscription, 0, len(rows))
	for _, s := range rows {
		res = append(res, domain.CometDSubscription{
			ID:           s.ID,
			ClientID:     s.ClientID,
			ChannelName:  s.ChannelName,
			LastReplayID: s.LastReplayID,
		})
	}
	return res, nil
}

func (r *Repository) UpdateCometDSubscriptionReplayID(ctx context.Context, id, replayID int64) error {
	return r.queries.UpdateCometDSubscriptionReplayID(ctx, gen.UpdateCometDSubscriptionReplayIDParams{
		LastReplayID: replayID,
		ID:           id,
	})
}
