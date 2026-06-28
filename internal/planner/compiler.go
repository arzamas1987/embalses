package planner

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CompilePlan turns a validated QueryIntent into an ExecutedPlan.
// No user input is ever interpolated into SQL — only allow-listed
// column/entity names are used, and values are passed as parameters.
func CompilePlan(intent QueryIntent) (ExecutedPlan, error) {
	plan := ExecutedPlan{
		Intent:  intent,
		Limit:   intent.Limit,
		OrderBy: "",
	}

	if plan.Limit <= 0 || plan.Limit > MaxLimit {
		plan.Limit = DefaultLimit
	}

	// Determine entity table and select columns
	table, columns, err := buildSelect(intent.Entity, intent.Metrics)
	if err != nil {
		return plan, err
	}
	plan.EntityTable = table
	plan.SelectedColumns = columns

	// Build WHERE clause with numbered parameters
	var whereParts []string
	var params []interface{}
	paramNum := 1

	if len(intent.Filters.Slugs) > 0 {
		placeholders := make([]string, len(intent.Filters.Slugs))
		for i := range intent.Filters.Slugs {
			placeholders[i] = fmt.Sprintf("$%d", paramNum)
			params = append(params, intent.Filters.Slugs[i])
			paramNum++
		}
		whereParts = append(whereParts, fmt.Sprintf("name = ANY(ARRAY[%s])", strings.Join(placeholders, ",")))
	}
	if intent.Filters.Basin != "" {
		whereParts = append(whereParts, fmt.Sprintf("basin_name = $%d", paramNum))
		params = append(params, intent.Filters.Basin)
		paramNum++
	}
	if intent.Filters.Province != "" {
		whereParts = append(whereParts, fmt.Sprintf("province_name = $%d", paramNum))
		params = append(params, intent.Filters.Province)
		paramNum++
	}
	if intent.Filters.Community != "" {
		whereParts = append(whereParts, fmt.Sprintf("community = $%d", paramNum))
		params = append(params, intent.Filters.Community)
		paramNum++
	}
	if intent.Filters.Since != "" {
		t, err := dateOrEmpty(intent.Filters.Since)
		if err != nil {
			return plan, err
		}
		if !t.IsZero() {
			whereParts = append(whereParts, fmt.Sprintf("observed_at >= $%d", paramNum))
			params = append(params, t)
			paramNum++
		}
	}
	if intent.Filters.Until != "" {
		t, err := dateOrEmpty(intent.Filters.Until)
		if err != nil {
			return plan, err
		}
		if !t.IsZero() {
			whereParts = append(whereParts, fmt.Sprintf("observed_at <= $%d", paramNum))
			params = append(params, t)
			paramNum++
		}
	}

	if len(whereParts) > 0 {
		plan.WhereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	// Build ORDER BY
	if intent.Sort != nil {
		order := "ASC"
		if strings.ToLower(intent.Sort.Order) == "desc" {
			order = "DESC"
		}
		plan.OrderBy = fmt.Sprintf("ORDER BY %s %s", allowListColumn(intent.Sort.Field), order)
	}

	// Build audit SQL (never executed directly)
	var sqlParts []string
	sqlParts = append(sqlParts, "SELECT")
	sqlParts = append(sqlParts, strings.Join(columns, ", "))
	sqlParts = append(sqlParts, "FROM", table)
	if plan.WhereClause != "" {
		sqlParts = append(sqlParts, plan.WhereClause)
	}
	if plan.OrderBy != "" {
		sqlParts = append(sqlParts, plan.OrderBy)
	}
	sqlParts = append(sqlParts, fmt.Sprintf("LIMIT %d", plan.Limit))
	plan.QuerySQL = strings.Join(sqlParts, " ")

	// Store parameter strings for audit
	for _, p := range params {
		plan.Parameters = append(plan.Parameters, fmt.Sprintf("%v", p))
	}

	return plan, nil
}

// ExecutePlan runs a compiled plan against the database and returns results.
func ExecutePlan(ctx context.Context, pool *pgxpool.Pool, plan ExecutedPlan) (QueryResult, error) {
	var result QueryResult

	// Build the actual parameterized SQL using the plan
	sql, params, err := buildParameterizedSQL(plan)
	if err != nil {
		return result, fmt.Errorf("build SQL: %w", err)
	}

	rows, err := pool.Query(ctx, sql, params...)
	if err != nil {
		return result, fmt.Errorf("execute query: %w", err)
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	colNames := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		colNames[i] = string(fd.Name)
	}

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return result, fmt.Errorf("scan row: %w", err)
		}
		row := make(map[string]interface{})
		for i, col := range colNames {
			row[col] = values[i]
		}
		result.Results = append(result.Results, row)
		result.Count++
	}
	if err := rows.Err(); err != nil {
		return result, fmt.Errorf("row iteration: %w", err)
	}

	result.Plan = plan
	return result, nil
}

// buildSelect returns the table name and column list for an entity + metrics.
func buildSelect(entity string, metrics []string) (string, []string, error) {
	var columns []string
	switch entity {
	case "reservoir":
		columns = append(columns, "r.id", "r.name", "b.name AS basin_name", "p.name AS province_name")
		for _, m := range metrics {
			columns = append(columns, metricColumn(m))
		}
		return "reservoirs r LEFT JOIN basins b ON r.basin_id = b.id LEFT JOIN provinces p ON r.province_id = p.id", columns, nil
	case "basin":
		columns = append(columns, "b.id", "b.name")
		for _, m := range metrics {
			columns = append(columns, metricColumn(m))
		}
		return "basins b", columns, nil
	case "province":
		columns = append(columns, "p.id", "p.name")
		for _, m := range metrics {
			columns = append(columns, metricColumn(m))
		}
		return "provinces p", columns, nil
	case "national":
		columns = append(columns, "'Spain' AS entity")
		for _, m := range metrics {
			columns = append(columns, metricColumn(m))
		}
		return "reservoirs r", columns, nil
	default:
		return "", nil, fmt.Errorf("unsupported entity: %s", entity)
	}
}

// metricColumn maps a metric name to a SQL expression.
func metricColumn(metric string) string {
	switch metric {
	case "fill_percent":
		return "COALESCE(rd.fill_pct, 0) AS fill_percent"
	case "stored_hm3":
		return "COALESCE(rd.volume_hm3, 0) AS stored_hm3"
	case "capacity_hm3":
		return "COALESCE(r.capacity_hm3, d.capacity_hm3, 0) AS capacity_hm3"
	case "change_hm3":
		return "COALESCE(rd.weekly_variation_hm3, 0) AS change_hm3"
	default:
		return "0"
	}
}

// allowListColumn maps a sort field to a safe SQL identifier.
func allowListColumn(field string) string {
	switch field {
	case "name":
		return "name"
	case "fill_percent":
		return "fill_percent"
	case "stored_hm3":
		return "stored_hm3"
	case "capacity_hm3":
		return "capacity_hm3"
	case "change_hm3":
		return "change_hm3"
	case "observed_at":
		return "observed_at"
	case "basin_name":
		return "basin_name"
	case "province_name":
		return "province_name"
	default:
		return "name" // safe fallback
	}
}

// buildParameterizedSQL constructs the actual SQL with parameters.
// This is the ONLY function that builds SQL strings, and it uses
// only allow-listed identifiers.
func buildParameterizedSQL(plan ExecutedPlan) (string, []interface{}, error) {
	var params []interface{}
	paramNum := 1

	// Build WHERE clause with fresh parameter numbering
	var whereParts []string
	intent := plan.Intent

	if len(intent.Filters.Slugs) > 0 {
		placeholders := make([]string, len(intent.Filters.Slugs))
		for i := range intent.Filters.Slugs {
			placeholders[i] = fmt.Sprintf("$%d", paramNum)
			params = append(params, intent.Filters.Slugs[i])
			paramNum++
		}
		whereParts = append(whereParts, fmt.Sprintf("name = ANY(ARRAY[%s])", strings.Join(placeholders, ",")))
	}
	if intent.Filters.Basin != "" {
		whereParts = append(whereParts, fmt.Sprintf("basin_name = $%d", paramNum))
		params = append(params, intent.Filters.Basin)
		paramNum++
	}
	if intent.Filters.Province != "" {
		whereParts = append(whereParts, fmt.Sprintf("province_name = $%d", paramNum))
		params = append(params, intent.Filters.Province)
		paramNum++
	}
	if intent.Filters.Since != "" {
		t, err := dateOrEmpty(intent.Filters.Since)
		if err != nil {
			return "", nil, err
		}
		if !t.IsZero() {
			whereParts = append(whereParts, fmt.Sprintf("observed_at >= $%d", paramNum))
			params = append(params, t)
			paramNum++
		}
	}
	if intent.Filters.Until != "" {
		t, err := dateOrEmpty(intent.Filters.Until)
		if err != nil {
			return "", nil, err
		}
		if !t.IsZero() {
			whereParts = append(whereParts, fmt.Sprintf("observed_at <= $%d", paramNum))
			params = append(params, t)
			paramNum++
		}
	}

	var sqlParts []string
	sqlParts = append(sqlParts, "SELECT")
	sqlParts = append(sqlParts, strings.Join(plan.SelectedColumns, ", "))
	sqlParts = append(sqlParts, "FROM", plan.EntityTable)
	if len(whereParts) > 0 {
		sqlParts = append(sqlParts, "WHERE", strings.Join(whereParts, " AND "))
	}
	if plan.OrderBy != "" {
		sqlParts = append(sqlParts, plan.OrderBy)
	}
	sqlParts = append(sqlParts, fmt.Sprintf("LIMIT %d", plan.Limit))

	return strings.Join(sqlParts, " "), params, nil
}
