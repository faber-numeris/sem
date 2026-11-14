package services

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/faber-numeris/sem/internal/core/domain"
)

func (s *Service) ExecuteToolingQuery(ctx context.Context, version, soql string) (*domain.QueryResult, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, err
	}
	p, err := parseSOQL(soql)
	if err != nil {
		return nil, err
	}
	v := normalizeVersion(version)
	rawRecords, err := s.repo.ListToolingRecords(ctx, v, p.SObject)
	if err != nil {
		return nil, err
	}
	matching := applyParsedSOQL(rawRecords, p)
	if p.IsCount {
		return &domain.QueryResult{
			TotalSize: len(matching),
			Done:      true,
			Records:   []map[string]any{},
		}, nil
	}
	return &domain.QueryResult{
		TotalSize: len(matching),
		Done:      true,
		Records:   matching,
	}, nil
}

func (s *Service) ExecuteAnonymousApex(ctx context.Context, version, body string) (*domain.AnonymousApexResult, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(body)
	if trimmed == "" || strings.Contains(strings.ToUpper(trimmed), "COMPILE_ERROR") {
		problem := "Unexpected token or empty anonymous Apex body"
		return &domain.AnonymousApexResult{
			Line:           1,
			Column:         1,
			Compiled:       false,
			Success:        false,
			CompileProblem: &problem,
		}, nil
	}
	if strings.Contains(strings.ToUpper(trimmed), "THROW NEW") {
		msg := "System.CalloutException: Simulated anonymous Apex exception"
		trace := "AnonymousBlock: line 1, column 1"
		return &domain.AnonymousApexResult{
			Line:                1,
			Column:              1,
			Compiled:            true,
			Success:             false,
			ExceptionMessage:    &msg,
			ExceptionStackTrace: &trace,
		}, nil
	}
	return &domain.AnonymousApexResult{
		Line:     -1,
		Column:   -1,
		Compiled: true,
		Success:  true,
	}, nil
}

func (s *Service) DescribeToolingSObject(ctx context.Context, version, sobject string) (map[string]any, error) {
	return s.DescribeSObject(ctx, version, sobject, true)
}

func (s *Service) RunTestsAsynchronous(ctx context.Context, version, classNames, testLevel string) (string, error) {
	if err := s.ValidateVersion(version); err != nil {
		return "", err
	}
	if testLevel == "" {
		testLevel = "RunSpecifiedTests"
	}
	if classNames == "" {
		classNames = "AccountServiceTest"
	}
	tr, err := s.repo.CreateApexTestRun(ctx, classNames, testLevel)
	if err != nil {
		return "", err
	}
	return tr.ID, nil
}

func (s *Service) ExecuteApexRest(ctx context.Context, urlMapping, method, reqBody string) (int, string, error) {
	normMapping := "/" + strings.TrimPrefix(strings.TrimSpace(urlMapping), "/")
	if method == "" {
		method = "GET"
	}
	route, err := s.repo.GetApexRestRoute(ctx, normMapping, method)
	if err == nil {
		return route.ResponseStatus, route.ResponseBody, nil
	}
	// Fallback for dynamic/unregistered mappings under /v1/* so custom Apex REST calls work out-of-the-box
	if strings.HasPrefix(normMapping, "/v1/") {
		resp := map[string]any{
			"status":  "OK",
			"mapping": normMapping,
			"method":  strings.ToUpper(method),
		}
		if reqBody != "" {
			var parsed any
			if json.Unmarshal([]byte(reqBody), &parsed) == nil {
				resp["received"] = parsed
			}
		}
		b, _ := json.Marshal(resp)
		return 200, string(b), nil
	}
	return 0, "", domain.ErrNotFound
}

func (s *Service) ProcessCometD(ctx context.Context, version string, messages []domain.BayeuxMessage) ([]domain.BayeuxMessage, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, err
	}
	var responses []domain.BayeuxMessage
	tVal := true
	fVal := false

	for _, msg := range messages {
		switch msg.Channel {
		case "/meta/handshake":
			clientID := "cometd_" + randomHex(10)
			if _, err := s.repo.UpsertCometDClient(ctx, clientID); err != nil {
				return nil, err
			}
			responses = append(responses, domain.BayeuxMessage{
				Channel:                  "/meta/handshake",
				ID:                       msg.ID,
				ClientID:                 clientID,
				Version:                  "1.0",
				MinimumVersion:           "1.0",
				SupportedConnectionTypes: []string{"long-polling"},
				Successful:               &tVal,
				Ext:                      map[string]any{"replay": true, "payload.format": true},
				Advice:                   map[string]any{"reconnect": "retry", "interval": 0, "timeout": 110000},
			})

		case "/meta/connect":
			if msg.ClientID == "" {
				responses = append(responses, domain.BayeuxMessage{
					Channel:    "/meta/connect",
					ID:         msg.ID,
					Successful: &fVal,
					Error:      "402::Unknown client",
				})
				continue
			}
			if _, err := s.repo.GetCometDClient(ctx, msg.ClientID); err != nil {
				responses = append(responses, domain.BayeuxMessage{
					Channel:    "/meta/connect",
					ID:         msg.ID,
					ClientID:   msg.ClientID,
					Successful: &fVal,
					Error:      "402::Unknown client",
				})
				continue
			}
			subs, err := s.repo.ListCometDSubscriptions(ctx, msg.ClientID)
			if err != nil {
				return nil, err
			}
			for _, sub := range subs {
				events, err := s.repo.ListPlatformEventsSince(ctx, sub.ChannelName, sub.LastReplayID)
				if err != nil {
					return nil, err
				}
				for _, ev := range events {
					var payloadMap map[string]any
					_ = json.Unmarshal([]byte(ev.Payload), &payloadMap)
					responses = append(responses, domain.BayeuxMessage{
						Channel: sub.ChannelName,
						Data: map[string]any{
							"schema":  ev.ID,
							"payload": payloadMap,
							"event": map[string]any{
								"replayId":    ev.ReplayID,
								"createdDate": ev.CreatedDate,
							},
						},
					})
					_ = s.repo.UpdateCometDSubscriptionReplayID(ctx, sub.ID, ev.ReplayID)
				}
			}
			responses = append(responses, domain.BayeuxMessage{
				Channel:    "/meta/connect",
				ID:         msg.ID,
				ClientID:   msg.ClientID,
				Successful: &tVal,
				Advice:     map[string]any{"reconnect": "retry", "interval": 0, "timeout": 110000},
			})

		case "/meta/subscribe":
			replayID := int64(0)
			if msg.Ext != nil {
				if replayMap, ok := msg.Ext["replay"].(map[string]any); ok {
					if v, ok := replayMap[msg.Subscription].(float64); ok && v > 0 {
						replayID = int64(v)
					}
				}
			}
			if _, err := s.repo.UpsertCometDSubscription(ctx, msg.ClientID, msg.Subscription, replayID); err != nil {
				return nil, err
			}
			responses = append(responses, domain.BayeuxMessage{
				Channel:      "/meta/subscribe",
				ID:           msg.ID,
				ClientID:     msg.ClientID,
				Subscription: msg.Subscription,
				Successful:   &tVal,
			})

		case "/meta/unsubscribe":
			if err := s.repo.DeleteCometDSubscription(ctx, msg.ClientID, msg.Subscription); err != nil {
				return nil, err
			}
			responses = append(responses, domain.BayeuxMessage{
				Channel:      "/meta/unsubscribe",
				ID:           msg.ID,
				ClientID:     msg.ClientID,
				Subscription: msg.Subscription,
				Successful:   &tVal,
			})

		case "/meta/disconnect":
			if err := s.repo.DisconnectCometDClient(ctx, msg.ClientID); err != nil {
				return nil, err
			}
			responses = append(responses, domain.BayeuxMessage{
				Channel:    "/meta/disconnect",
				ID:         msg.ID,
				ClientID:   msg.ClientID,
				Successful: &tVal,
			})

		default:
			if strings.HasPrefix(msg.Channel, "/event/") || strings.HasPrefix(msg.Channel, "/data/") {
				payloadBytes, _ := json.Marshal(msg.Data)
				eventType := strings.TrimPrefix(msg.Channel, "/event/")
				if _, err := s.repo.CreatePlatformEvent(ctx, msg.Channel, eventType, string(payloadBytes)); err != nil {
					return nil, err
				}
				responses = append(responses, domain.BayeuxMessage{
					Channel:    msg.Channel,
					ID:         msg.ID,
					ClientID:   msg.ClientID,
					Successful: &tVal,
				})
			} else {
				responses = append(responses, domain.BayeuxMessage{
					Channel:    msg.Channel,
					ID:         msg.ID,
					ClientID:   msg.ClientID,
					Successful: &fVal,
					Error:      "400::Unsupported channel",
				})
			}
		}
	}
	return responses, nil
}
