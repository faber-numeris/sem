package services

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/faber-numeris/sem/internal/core/domain"
)

type parsedSOQL struct {
	Fields      []string
	IsCount     bool
	SObject     string
	WhereClause string
	OrderBy     string
	OrderDesc   bool
	Limit       int
	Offset      int
}

var (
	reSelect = regexp.MustCompile(`(?is)^\s*SELECT\s+(.+?)\s+FROM\s+([a-zA-Z0-9_]+)(.*)$`)
	reLimit  = regexp.MustCompile(`(?is)\bLIMIT\s+(\d+)\b`)
	reOffset = regexp.MustCompile(`(?is)\bOFFSET\s+(\d+)\b`)
	reOrder  = regexp.MustCompile(`(?is)\bORDER\s+BY\s+([a-zA-Z0-9_.]+)(?:\s+(ASC|DESC))?`)
	reWhere  = regexp.MustCompile(`(?is)\bWHERE\s+(.+?)(?:\bORDER\s+BY\b|\bLIMIT\b|\bOFFSET\b|$)`)
)

func parseSOQL(raw string) (*parsedSOQL, error) {
	q := strings.TrimSpace(raw)
	m := reSelect.FindStringSubmatch(q)
	if m == nil {
		return nil, domain.ErrMalformedQuery
	}
	fieldPart := strings.TrimSpace(m[1])
	sobject := strings.TrimSpace(m[2])
	tail := m[3]

	p := &parsedSOQL{
		SObject: sobject,
		Limit:   -1,
	}

	if strings.EqualFold(fieldPart, "COUNT()") {
		p.IsCount = true
	} else {
		for _, f := range strings.Split(fieldPart, ",") {
			trimmed := strings.TrimSpace(f)
			if trimmed != "" {
				p.Fields = append(p.Fields, trimmed)
			}
		}
		if len(p.Fields) == 0 {
			return nil, domain.ErrMalformedQuery
		}
	}

	if wm := reWhere.FindStringSubmatch(tail); wm != nil {
		p.WhereClause = strings.TrimSpace(wm[1])
	}
	if om := reOrder.FindStringSubmatch(tail); om != nil {
		p.OrderBy = strings.TrimSpace(om[1])
		p.OrderDesc = strings.EqualFold(om[2], "DESC")
	}
	if lm := reLimit.FindStringSubmatch(tail); lm != nil {
		p.Limit, _ = strconv.Atoi(lm[1])
	}
	if ofm := reOffset.FindStringSubmatch(tail); ofm != nil {
		p.Offset, _ = strconv.Atoi(ofm[1])
	}
	return p, nil
}

func getFieldIgnoreCase(rec map[string]any, field string) any {
	for k, v := range rec {
		if strings.EqualFold(k, field) {
			return v
		}
	}
	return nil
}

func stringifyFieldVal(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case *string:
		if val == nil {
			return ""
		}
		return *val
	case *float64:
		if val == nil {
			return ""
		}
		return strconv.FormatFloat(*val, 'f', -1, 64)
	case *int:
		if val == nil {
			return ""
		}
		return strconv.Itoa(*val)
	case *bool:
		if val == nil {
			return ""
		}
		return strconv.FormatBool(*val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func stripQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && ((s[0] == '\'' && s[len(s)-1] == '\'') || (s[0] == '"' && s[len(s)-1] == '"')) {
		return s[1 : len(s)-1]
	}
	return s
}

func evalSimpleCondition(rec map[string]any, cond string) bool {
	cond = strings.TrimSpace(cond)
	for strings.HasPrefix(cond, "(") && strings.HasSuffix(cond, ")") {
		cond = strings.TrimSpace(cond[1 : len(cond)-1])
	}

	// IN (...)
	reIn := regexp.MustCompile(`(?is)^([a-zA-Z0-9_]+)\s+IN\s*\((.+)\)$`)
	if m := reIn.FindStringSubmatch(cond); m != nil {
		actual := stringifyFieldVal(getFieldIgnoreCase(rec, m[1]))
		for _, item := range strings.Split(m[2], ",") {
			if strings.EqualFold(actual, stripQuotes(item)) {
				return true
			}
		}
		return false
	}

	// LIKE
	reLike := regexp.MustCompile(`(?is)^([a-zA-Z0-9_]+)\s+LIKE\s+(.+)$`)
	if m := reLike.FindStringSubmatch(cond); m != nil {
		actual := strings.ToLower(stringifyFieldVal(getFieldIgnoreCase(rec, m[1])))
		pattern := strings.ToLower(stripQuotes(m[2]))
		regexPat := "^" + strings.ReplaceAll(regexp.QuoteMeta(pattern), "%", ".*") + "$"
		matched, _ := regexp.MatchString(regexPat, actual)
		return matched
	}

	// Comparison operators: !=, <=, >=, =, <, >
	for _, op := range []string{"!=", "<>", "<=", ">=", "=", "<", ">"} {
		idx := strings.Index(cond, op)
		if idx > 0 {
			field := strings.TrimSpace(cond[:idx])
			rhsRaw := strings.TrimSpace(cond[idx+len(op):])
			rhs := stripQuotes(rhsRaw)
			if strings.EqualFold(rhs, "null") {
				rhs = ""
			}
			actual := stringifyFieldVal(getFieldIgnoreCase(rec, field))
			fActual, err1 := strconv.ParseFloat(actual, 64)
			fRhs, err2 := strconv.ParseFloat(rhs, 64)
			switch op {
			case "=":
				return strings.EqualFold(actual, rhs)
			case "!=", "<>":
				return !strings.EqualFold(actual, rhs)
			case "<":
				if err1 == nil && err2 == nil {
					return fActual < fRhs
				}
				return actual < rhs
			case ">":
				if err1 == nil && err2 == nil {
					return fActual > fRhs
				}
				return actual > rhs
			case "<=":
				if err1 == nil && err2 == nil {
					return fActual <= fRhs
				}
				return actual <= rhs
			case ">=":
				if err1 == nil && err2 == nil {
					return fActual >= fRhs
				}
				return actual >= rhs
			}
		}
	}
	return true
}

func matchesWhere(rec map[string]any, where string) bool {
	if strings.TrimSpace(where) == "" {
		return true
	}
	reOr := regexp.MustCompile(`(?i)\s+OR\s+`)
	reAnd := regexp.MustCompile(`(?i)\s+AND\s+`)
	orParts := reOr.Split(where, -1)
	for _, orPart := range orParts {
		andParts := reAnd.Split(orPart, -1)
		allMatch := true
		for _, andPart := range andParts {
			if !evalSimpleCondition(rec, andPart) {
				allMatch = false
				break
			}
		}
		if allMatch {
			return true
		}
	}
	return false
}

func projectRecord(rec map[string]any, fields []string) map[string]any {
	if len(fields) == 1 && (fields[0] == "*" || strings.HasPrefix(strings.ToUpper(fields[0]), "FIELDS(")) {
		return rec
	}
	out := map[string]any{
		"attributes": rec["attributes"],
	}
	for _, f := range fields {
		for k, v := range rec {
			if strings.EqualFold(k, f) {
				out[k] = v
				break
			}
		}
	}
	return out
}

func applyParsedSOQL(records []map[string]any, p *parsedSOQL) []map[string]any {
	filtered := make([]map[string]any, 0, len(records))
	for _, r := range records {
		if matchesWhere(r, p.WhereClause) {
			filtered = append(filtered, r)
		}
	}

	if p.OrderBy != "" {
		sort.SliceStable(filtered, func(i, j int) bool {
			vi := stringifyFieldVal(getFieldIgnoreCase(filtered[i], p.OrderBy))
			vj := stringifyFieldVal(getFieldIgnoreCase(filtered[j], p.OrderBy))
			if p.OrderDesc {
				return vi > vj
			}
			return vi < vj
		})
	}

	if p.Offset > 0 {
		if p.Offset >= len(filtered) {
			filtered = []map[string]any{}
		} else {
			filtered = filtered[p.Offset:]
		}
	}
	if p.Limit >= 0 && p.Limit < len(filtered) {
		filtered = filtered[:p.Limit]
	}

	if p.IsCount {
		return filtered
	}

	projected := make([]map[string]any, 0, len(filtered))
	for _, r := range filtered {
		projected = append(projected, projectRecord(r, p.Fields))
	}
	return projected
}

func (s *Service) ExecuteSOQL(ctx context.Context, version, soql string, includeDeleted bool, batchSize int) (*domain.QueryResult, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, err
	}
	p, err := parseSOQL(soql)
	if err != nil {
		return nil, err
	}
	v := normalizeVersion(version)
	rawRecords, err := s.repo.ListRecordsForSObject(ctx, v, p.SObject, includeDeleted)
	if err != nil {
		return nil, err
	}

	allMatching := applyParsedSOQL(rawRecords, p)
	totalSize := len(allMatching)
	if p.IsCount {
		return &domain.QueryResult{
			TotalSize: totalSize,
			Done:      true,
			Records:   []map[string]any{},
		}, nil
	}

	if batchSize <= 0 {
		batchSize = 2000
	}

	if totalSize > batchSize {
		cursorID := "01g" + randomHex(8)
		_, err := s.repo.CreateQueryCursor(ctx, &domain.QueryCursor{
			ID:             cursorID,
			SObjectName:    p.SObject,
			SOQLQuery:      soql,
			IncludeDeleted: includeDeleted,
			OffsetValue:    batchSize,
			BatchSize:      batchSize,
			TotalSize:      totalSize,
		})
		if err != nil {
			return nil, err
		}
		return &domain.QueryResult{
			TotalSize:      totalSize,
			Done:           false,
			NextRecordsURL: fmt.Sprintf("/services/data/v%s/query/%s-%d", v, cursorID, batchSize),
			Records:        allMatching[:batchSize],
		}, nil
	}

	return &domain.QueryResult{
		TotalSize: totalSize,
		Done:      true,
		Records:   allMatching,
	}, nil
}

func (s *Service) QueryNextBatch(ctx context.Context, version, cursorIdentifier string) (*domain.QueryResult, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, err
	}
	v := normalizeVersion(version)
	cursorID := cursorIdentifier
	if idx := strings.LastIndex(cursorIdentifier, "-"); idx > 0 {
		cursorID = cursorIdentifier[:idx]
	}

	cursor, err := s.repo.GetQueryCursor(ctx, cursorID)
	if err != nil {
		return nil, err
	}
	p, err := parseSOQL(cursor.SOQLQuery)
	if err != nil {
		return nil, err
	}
	rawRecords, err := s.repo.ListRecordsForSObject(ctx, v, cursor.SObjectName, cursor.IncludeDeleted)
	if err != nil {
		return nil, err
	}
	allMatching := applyParsedSOQL(rawRecords, p)
	totalSize := len(allMatching)
	start := cursor.OffsetValue
	if start > totalSize {
		start = totalSize
	}
	end := start + cursor.BatchSize
	if end >= totalSize {
		_ = s.repo.DeleteQueryCursor(ctx, cursor.ID)
		return &domain.QueryResult{
			TotalSize: totalSize,
			Done:      true,
			Records:   allMatching[start:],
		}, nil
	}

	if _, err := s.repo.UpdateQueryCursorOffset(ctx, cursor.ID, end); err != nil {
		return nil, err
	}
	return &domain.QueryResult{
		TotalSize:      totalSize,
		Done:           false,
		NextRecordsURL: fmt.Sprintf("/services/data/v%s/query/%s-%d", v, cursor.ID, end),
		Records:        allMatching[start:end],
	}, nil
}

var (
	reSOSLFind      = regexp.MustCompile(`(?is)FIND\s*\{([^}]+)\}`)
	reSOSLReturning = regexp.MustCompile(`(?is)RETURNING\s+(.+)$`)
	reSOSLTarget    = regexp.MustCompile(`(?is)([a-zA-Z0-9_]+)(?:\s*\(([^)]+)\))?`)
)

func (s *Service) ExecuteSOSL(ctx context.Context, version, sosl string) (*domain.SearchResult, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, err
	}
	fm := reSOSLFind.FindStringSubmatch(sosl)
	if fm == nil {
		return nil, domain.ErrMalformedSearch
	}
	term := strings.TrimSpace(fm[1])
	v := normalizeVersion(version)

	type targetSpec struct {
		sobject string
		fields  []string
	}
	var targets []targetSpec
	if rm := reSOSLReturning.FindStringSubmatch(sosl); rm != nil {
		matches := reSOSLTarget.FindAllStringSubmatch(rm[1], -1)
		for _, m := range matches {
			obj := strings.TrimSpace(m[1])
			var flds []string
			if m[2] != "" {
				for _, f := range strings.Split(m[2], ",") {
					if tf := strings.TrimSpace(f); tf != "" {
						flds = append(flds, tf)
					}
				}
			}
			targets = append(targets, targetSpec{sobject: obj, fields: flds})
		}
	}
	if len(targets) == 0 {
		targets = []targetSpec{
			{sobject: "Account", fields: []string{"Id", "Name"}},
			{sobject: "Contact", fields: []string{"Id", "Name"}},
			{sobject: "Lead", fields: []string{"Id", "Name"}},
		}
	}

	var searchRecords []map[string]any
	for _, tgt := range targets {
		recs, err := s.repo.SearchRecordsForSObject(ctx, v, tgt.sobject, term)
		if err != nil {
			return nil, err
		}
		for _, r := range recs {
			if len(tgt.fields) > 0 {
				searchRecords = append(searchRecords, projectRecord(r, tgt.fields))
			} else {
				searchRecords = append(searchRecords, r)
			}
		}
	}
	if searchRecords == nil {
		searchRecords = []map[string]any{}
	}
	return &domain.SearchResult{SearchRecords: searchRecords}, nil
}
