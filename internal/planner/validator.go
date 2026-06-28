package planner

import (
	"fmt"
	"strings"
)

// ValidateIntent checks a QueryIntent against all allow-lists.
// It returns an error describing the first violation found.
func ValidateIntent(intent QueryIntent) error {
	if intent.Entity == "" {
		return fmt.Errorf("entity is required")
	}
	if !ValidEntities[intent.Entity] {
		return fmt.Errorf("invalid entity %q: allowed values are %v", intent.Entity, keys(ValidEntities))
	}

	if len(intent.Metrics) == 0 {
		return fmt.Errorf("at least one metric is required")
	}
	allowedMetrics, ok := ValidMetrics[intent.Entity]
	if !ok {
		return fmt.Errorf("no metrics defined for entity %q", intent.Entity)
	}
	for _, m := range intent.Metrics {
		if !allowedMetrics[m] {
			return fmt.Errorf("invalid metric %q for entity %q: allowed values are %v", m, intent.Entity, keys(allowedMetrics))
		}
	}

	if intent.Aggregation == "" {
		return fmt.Errorf("aggregation is required")
	}
	if !ValidAggregations[intent.Aggregation] {
		return fmt.Errorf("invalid aggregation %q: allowed values are %v", intent.Aggregation, keys(ValidAggregations))
	}

	// Validate sort
	if intent.Sort != nil {
		if intent.Sort.Field == "" {
			return fmt.Errorf("sort.field is required when sort is provided")
		}
		if !ValidSortFields[intent.Sort.Field] {
			return fmt.Errorf("invalid sort field %q: allowed values are %v", intent.Sort.Field, keys(ValidSortFields))
		}
		order := strings.ToLower(intent.Sort.Order)
		if order == "" {
			order = "asc"
		}
		if !ValidSortOrders[order] {
			return fmt.Errorf("invalid sort order %q: allowed values are %v", intent.Sort.Order, keys(ValidSortOrders))
		}
	}

	// Validate limit
	if intent.Limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}
	if intent.Limit > MaxLimit {
		return fmt.Errorf("limit %d exceeds maximum %d", intent.Limit, MaxLimit)
	}

	// Validate chart hint
	if intent.ChartHint != "" && !ValidChartHints[intent.ChartHint] {
		return fmt.Errorf("invalid chart_hint %q: allowed values are %v", intent.ChartHint, keys(ValidChartHints))
	}

	// Validate filters for injection attempts
	if err := validateFilters(intent.Filters); err != nil {
		return err
	}

	return nil
}

func validateFilters(f Filters) error {
	// Reject any filter value that looks like SQL
	for _, s := range f.Slugs {
		if err := rejectSQLInjection(s); err != nil {
			return err
		}
	}
	if err := rejectSQLInjection(f.Basin); err != nil {
		return err
	}
	if err := rejectSQLInjection(f.Province); err != nil {
		return err
	}
	if err := rejectSQLInjection(f.Community); err != nil {
		return err
	}
	if err := rejectSQLInjection(f.Since); err != nil {
		return err
	}
	if err := rejectSQLInjection(f.Until); err != nil {
		return err
	}
	return nil
}

// rejectSQLInjection blocks strings containing suspicious SQL characters/patterns.
func rejectSQLInjection(s string) error {
	if s == "" {
		return nil
	}
	lower := strings.ToLower(s)
	suspicious := []string{
		";", "--", "/*", "*/", "drop ", "delete ", "insert ",
		"update ", "select ", "union ", "exec ", "execute ",
		"xp_", "sp_", "'", "\"", "\x00", "\x1a",
	}
	for _, p := range suspicious {
		if strings.Contains(lower, p) {
			return fmt.Errorf("filter value contains disallowed pattern: %q", p)
		}
	}
	return nil
}

func keys(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}
