package postgres

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/faber-numeris/sem/internal/adapters/outbound/postgres/gen"
	"github.com/faber-numeris/sem/internal/core/domain"
	"github.com/faber-numeris/sem/internal/ports/outbound"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type Repository struct {
	queries gen.Querier
	pingFn  func() bool
}

var _ outbound.Repository = (*Repository)(nil)

func NewRepository(db gen.DBTX) *Repository {
	pingFn := func() bool {
		if p, ok := db.(Pinger); ok {
			return p.Ping(context.Background()) == nil
		}
		return true
	}
	return &Repository{
		queries: gen.New(db),
		pingFn:  pingFn,
	}
}

func NewRepositoryWithQuerier(q gen.Querier, pingFn func() bool) *Repository {
	return &Repository{
		queries: q,
		pingFn:  pingFn,
	}
}

func (r *Repository) Ping() bool {
	if r.pingFn != nil {
		return r.pingFn()
	}
	return true
}

func formatSFTime(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000+0000")
}

func formatSFTimePtr(t *time.Time) any {
	if t == nil {
		return nil
	}
	return formatSFTime(*t)
}

func makeAttributes(version, sobject, id string) domain.RecordAttributes {
	v := strings.TrimPrefix(version, "v")
	if v == "" {
		v = "61.0"
	}
	return domain.RecordAttributes{
		Type: sobject,
		URL:  fmt.Sprintf("/services/data/v%s/sobjects/%s/%s", v, sobject, id),
	}
}

func lookupCaseInsensitive(m map[string]any, key string) (any, bool) {
	if v, ok := m[key]; ok {
		return v, true
	}
	lower := strings.ToLower(key)
	for k, v := range m {
		if strings.ToLower(k) == lower {
			return v, true
		}
	}
	return nil, false
}

func getString(m map[string]any, key, def string) string {
	v, ok := lookupCaseInsensitive(m, key)
	if !ok || v == nil {
		return def
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func getStringPtr(m map[string]any, key string) *string {
	v, ok := lookupCaseInsensitive(m, key)
	if !ok || v == nil {
		return nil
	}
	s := fmt.Sprintf("%v", v)
	return &s
}

func getFloatPtr(m map[string]any, key string) *float64 {
	v, ok := lookupCaseInsensitive(m, key)
	if !ok || v == nil {
		return nil
	}
	switch val := v.(type) {
	case float64:
		return &val
	case int:
		f := float64(val)
		return &f
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return &f
		}
	}
	return nil
}

func getIntPtr(m map[string]any, key string) *int {
	v, ok := lookupCaseInsensitive(m, key)
	if !ok || v == nil {
		return nil
	}
	switch val := v.(type) {
	case int:
		return &val
	case float64:
		i := int(val)
		return &i
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return &i
		}
	}
	return nil
}

func getBoolPtr(m map[string]any, key string) *bool {
	v, ok := lookupCaseInsensitive(m, key)
	if !ok || v == nil {
		return nil
	}
	switch val := v.(type) {
	case bool:
		return &val
	case string:
		if b, err := strconv.ParseBool(val); err == nil {
			return &b
		}
	}
	return nil
}

func getTimePtr(m map[string]any, key string) *time.Time {
	v, ok := lookupCaseInsensitive(m, key)
	if !ok || v == nil {
		return nil
	}
	switch val := v.(type) {
	case time.Time:
		return &val
	case string:
		for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05.000+0000", "2006-01-02"} {
			if t, err := time.Parse(layout, val); err == nil {
				return &t
			}
		}
	}
	return nil
}
