package planner

import (
	"encoding/json"
	"fmt"
	"time"
)

// QueryIntent is the user-submitted structured query.
// It must pass validation before compilation.
type QueryIntent struct {
	Entity      string    `json:"entity"`
	Metrics     []string  `json:"metrics"`
	Filters     Filters   `json:"filters,omitempty"`
	Aggregation string    `json:"aggregation"`
	Sort        *SortSpec `json:"sort,omitempty"`
	Limit       int       `json:"limit,omitempty"`
	ChartHint   string    `json:"chart_hint,omitempty"`
}

// Filters holds all filter criteria.
type Filters struct {
	Slugs     []string `json:"slugs,omitempty"`
	Basin     string   `json:"basin,omitempty"`
	Province  string   `json:"province,omitempty"`
	Community string   `json:"community,omitempty"`
	Since     string   `json:"since,omitempty"`
	Until     string   `json:"until,omitempty"`
}

// SortSpec defines ordering.
type SortSpec struct {
	Field string `json:"field"`
	Order string `json:"order,omitempty"`
}

// ExecutedPlan is the compiled plan + SQL for transparency.
type ExecutedPlan struct {
	Intent          QueryIntent `json:"intent"`
	EntityTable     string      `json:"entity_table"`
	SelectedColumns []string    `json:"selected_columns"`
	WhereClause     string      `json:"where_clause"`
	OrderBy         string      `json:"order_by"`
	Limit           int         `json:"limit"`
	Parameters      []string    `json:"parameters"`
	QuerySQL        string      `json:"query_sql"` // For audit; never executed directly
}

// QueryResult is the response from the planner executor.
type QueryResult struct {
	Results []map[string]interface{} `json:"results"`
	Plan    ExecutedPlan             `json:"plan"`
	Count   int                      `json:"count"`
}

// ValidEntities is the allow-list of queryable entities.
var ValidEntities = map[string]bool{
	"reservoir": true,
	"basin":     true,
	"province":  true,
	"community": true,
	"national":  true,
}

// ValidMetrics is the allow-list of metrics per entity.
var ValidMetrics = map[string]map[string]bool{
	"reservoir": {
		"fill_percent": true,
		"stored_hm3":   true,
		"capacity_hm3": true,
		"change_hm3":   true,
	},
	"basin": {
		"fill_percent": true,
		"stored_hm3":   true,
		"capacity_hm3": true,
	},
	"province": {
		"fill_percent": true,
		"stored_hm3":   true,
		"capacity_hm3": true,
	},
	"community": {
		"fill_percent": true,
		"stored_hm3":   true,
		"capacity_hm3": true,
	},
	"national": {
		"fill_percent": true,
		"stored_hm3":   true,
		"capacity_hm3": true,
	},
}

// ValidAggregations is the allow-list of aggregation modes.
var ValidAggregations = map[string]bool{
	"latest":     true,
	"timeseries": true,
	"ranking":    true,
	"compare":    true,
	"summary":    true,
}

// ValidSortFields is the allow-list of sortable fields.
var ValidSortFields = map[string]bool{
	"name":          true,
	"fill_percent":  true,
	"stored_hm3":    true,
	"capacity_hm3":  true,
	"change_hm3":    true,
	"observed_at":   true,
	"basin_name":    true,
	"province_name": true,
}

// ValidSortOrders is the allow-list of sort directions.
var ValidSortOrders = map[string]bool{
	"asc":  true,
	"desc": true,
}

// ValidChartHints is the allow-list of chart hints.
var ValidChartHints = map[string]bool{
	"table": true,
	"line":  true,
	"bar":   true,
	"map":   true,
	"none":  true,
}

// MaxLimit is the hard upper bound on result rows.
const MaxLimit = 500

// DefaultLimit is the default when not specified or zero.
const DefaultLimit = 20

// ParseIntent parses and validates a JSON string into a QueryIntent.
func ParseIntent(raw []byte) (QueryIntent, error) {
	var intent QueryIntent
	if err := json.Unmarshal(raw, &intent); err != nil {
		return intent, fmt.Errorf("invalid JSON: %w", err)
	}
	return intent, nil
}

// dateOrEmpty parses a date string; empty returns zero time.
func dateOrEmpty(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format %q (expected YYYY-MM-DD): %w", s, err)
	}
	return t, nil
}
