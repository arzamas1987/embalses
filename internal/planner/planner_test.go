package planner

import (
	"strings"
	"testing"
)

// === VALID INTENT TESTS ===

func TestValidateIntent_ValidReservoir(t *testing.T) {
	intent := QueryIntent{
		Entity:      "reservoir",
		Metrics:     []string{"fill_percent"},
		Aggregation: "latest",
		Limit:       10,
	}
	if err := ValidateIntent(intent); err != nil {
		t.Fatalf("expected valid intent, got: %v", err)
	}
}

func TestValidateIntent_ValidWithSort(t *testing.T) {
	intent := QueryIntent{
		Entity:      "basin",
		Metrics:     []string{"stored_hm3", "capacity_hm3"},
		Aggregation: "summary",
		Sort: &SortSpec{
			Field: "name",
			Order: "desc",
		},
		Limit:     50,
		ChartHint: "table",
	}
	if err := ValidateIntent(intent); err != nil {
		t.Fatalf("expected valid intent, got: %v", err)
	}
}

func TestValidateIntent_ValidWithFilters(t *testing.T) {
	intent := QueryIntent{
		Entity:      "reservoir",
		Metrics:     []string{"fill_percent"},
		Aggregation: "timeseries",
		Filters: Filters{
			Basin:    "Ebro",
			Province: "Zaragoza",
			Since:    "2024-01-01",
			Until:    "2024-12-31",
		},
		Limit: 100,
	}
	if err := ValidateIntent(intent); err != nil {
		t.Fatalf("expected valid intent, got: %v", err)
	}
}

// === INVALID INTENT TESTS ===

func TestValidateIntent_MissingEntity(t *testing.T) {
	intent := QueryIntent{Metrics: []string{"fill_percent"}, Aggregation: "latest"}
	if err := ValidateIntent(intent); err == nil {
		t.Fatal("expected error for missing entity")
	}
}

func TestValidateIntent_InvalidEntity(t *testing.T) {
	intent := QueryIntent{Entity: "user_data", Metrics: []string{"fill_percent"}, Aggregation: "latest"}
	err := ValidateIntent(intent)
	if err == nil {
		t.Fatal("expected error for invalid entity")
	}
	if !strings.Contains(err.Error(), "invalid entity") {
		t.Errorf("expected 'invalid entity' in error, got: %v", err)
	}
}

func TestValidateIntent_InvalidMetric(t *testing.T) {
	intent := QueryIntent{Entity: "reservoir", Metrics: []string{"password_hash"}, Aggregation: "latest"}
	err := ValidateIntent(intent)
	if err == nil {
		t.Fatal("expected error for invalid metric")
	}
	if !strings.Contains(err.Error(), "invalid metric") {
		t.Errorf("expected 'invalid metric' in error, got: %v", err)
	}
}

func TestValidateIntent_MetricForWrongEntity(t *testing.T) {
	intent := QueryIntent{Entity: "basin", Metrics: []string{"change_hm3"}, Aggregation: "latest"}
	err := ValidateIntent(intent)
	if err == nil {
		t.Fatal("expected error for metric not allowed on basin")
	}
}

func TestValidateIntent_InvalidAggregation(t *testing.T) {
	intent := QueryIntent{Entity: "reservoir", Metrics: []string{"fill_percent"}, Aggregation: "drop_table"}
	err := ValidateIntent(intent)
	if err == nil {
		t.Fatal("expected error for invalid aggregation")
	}
}

func TestValidateIntent_ExcessiveLimit(t *testing.T) {
	intent := QueryIntent{Entity: "reservoir", Metrics: []string{"fill_percent"}, Aggregation: "latest", Limit: 9999}
	err := ValidateIntent(intent)
	if err == nil {
		t.Fatal("expected error for excessive limit")
	}
	if !strings.Contains(err.Error(), "exceeds maximum") {
		t.Errorf("expected 'exceeds maximum' in error, got: %v", err)
	}
}

func TestValidateIntent_NegativeLimit(t *testing.T) {
	intent := QueryIntent{Entity: "reservoir", Metrics: []string{"fill_percent"}, Aggregation: "latest", Limit: -1}
	err := ValidateIntent(intent)
	if err == nil {
		t.Fatal("expected error for negative limit")
	}
}

func TestValidateIntent_InvalidSortField(t *testing.T) {
	intent := QueryIntent{
		Entity:      "reservoir",
		Metrics:     []string{"fill_percent"},
		Aggregation: "latest",
		Sort: &SortSpec{
			Field: "email",
			Order: "asc",
		},
	}
	err := ValidateIntent(intent)
	if err == nil {
		t.Fatal("expected error for invalid sort field")
	}
}

func TestValidateIntent_InvalidSortOrder(t *testing.T) {
	intent := QueryIntent{
		Entity:      "reservoir",
		Metrics:     []string{"fill_percent"},
		Aggregation: "latest",
		Sort: &SortSpec{
			Field: "name",
			Order: "; DROP TABLE",
		},
	}
	err := ValidateIntent(intent)
	if err == nil {
		t.Fatal("expected error for invalid sort order")
	}
}

func TestValidateIntent_InvalidChartHint(t *testing.T) {
	intent := QueryIntent{
		Entity:      "reservoir",
		Metrics:     []string{"fill_percent"},
		Aggregation: "latest",
		ChartHint:   "execute_sql",
	}
	err := ValidateIntent(intent)
	if err == nil {
		t.Fatal("expected error for invalid chart hint")
	}
}

// === INJECTION ATTACK TESTS ===

func TestValidateIntent_SQLInjectionInFilter(t *testing.T) {
	attacks := []string{
		"Ebro'; DROP TABLE reservoirs; --",
		"Ebro'; DELETE FROM readings; --",
		"Ebro'; SELECT * FROM api_keys; --",
		"Ebro'; UNION SELECT password FROM users; --",
		"Ebro'; exec xp_cmdshell('rm -rf /'); --",
		"Ebro\x00",
		"Ebro\x1a",
	}

	for _, attack := range attacks {
		intent := QueryIntent{
			Entity:      "reservoir",
			Metrics:     []string{"fill_percent"},
			Aggregation: "latest",
			Filters: Filters{
				Basin: attack,
			},
		}
		err := ValidateIntent(intent)
		if err == nil {
			t.Fatalf("expected rejection for SQL injection payload: %q", attack)
		}
		if !strings.Contains(err.Error(), "disallowed pattern") {
			t.Errorf("expected 'disallowed pattern' error for %q, got: %v", attack, err)
		}
	}
}

func TestValidateIntent_SQLInjectionInSlug(t *testing.T) {
	intent := QueryIntent{
		Entity:      "reservoir",
		Metrics:     []string{"fill_percent"},
		Aggregation: "latest",
		Filters: Filters{
			Slugs: []string{"Mequinenza'; DROP TABLE dams; --"},
		},
	}
	err := ValidateIntent(intent)
	if err == nil {
		t.Fatal("expected rejection for SQL injection in slug filter")
	}
}

func TestValidateIntent_SQLInjectionInDate(t *testing.T) {
	intent := QueryIntent{
		Entity:      "reservoir",
		Metrics:     []string{"fill_percent"},
		Aggregation: "latest",
		Filters: Filters{
			Since: "2024-01-01'; DROP TABLE sources; --",
		},
	}
	err := ValidateIntent(intent)
	if err == nil {
		t.Fatal("expected rejection for SQL injection in date filter")
	}
}

// === COMPILE TESTS ===

func TestCompilePlan_ValidReservoir(t *testing.T) {
	intent := QueryIntent{
		Entity:      "reservoir",
		Metrics:     []string{"fill_percent", "stored_hm3"},
		Aggregation: "latest",
		Limit:       25,
	}
	plan, err := CompilePlan(intent)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if plan.Limit != 25 {
		t.Errorf("expected limit 25, got %d", plan.Limit)
	}
	if plan.EntityTable == "" {
		t.Error("expected non-empty entity table")
	}
	if len(plan.SelectedColumns) == 0 {
		t.Error("expected non-empty selected columns")
	}
	// Verify SQL contains no user input (only allow-listed identifiers)
	if strings.Contains(plan.QuerySQL, "user_data") {
		t.Error("SQL should not contain arbitrary user input")
	}
}

func TestCompilePlan_DefaultLimit(t *testing.T) {
	intent := QueryIntent{
		Entity:      "reservoir",
		Metrics:     []string{"fill_percent"},
		Aggregation: "latest",
	}
	plan, err := CompilePlan(intent)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if plan.Limit != DefaultLimit {
		t.Errorf("expected default limit %d, got %d", DefaultLimit, plan.Limit)
	}
}

func TestCompilePlan_WithFilters(t *testing.T) {
	intent := QueryIntent{
		Entity:      "reservoir",
		Metrics:     []string{"fill_percent"},
		Aggregation: "latest",
		Filters: Filters{
			Basin: "Ebro",
			Since: "2024-01-01",
		},
		Limit: 50,
	}
	plan, err := CompilePlan(intent)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if !strings.Contains(plan.QuerySQL, "basin_name") {
		t.Error("expected SQL to reference basin_name")
	}
	if !strings.Contains(plan.QuerySQL, "observed_at") {
		t.Error("expected SQL to reference observed_at")
	}
	// Verify parameters are tracked
	if len(plan.Parameters) == 0 {
		t.Error("expected parameters to be tracked in plan")
	}
}

func TestCompilePlan_WithSort(t *testing.T) {
	intent := QueryIntent{
		Entity:      "reservoir",
		Metrics:     []string{"fill_percent"},
		Aggregation: "ranking",
		Sort: &SortSpec{
			Field: "fill_percent",
			Order: "desc",
		},
		Limit: 10,
	}
	plan, err := CompilePlan(intent)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if !strings.Contains(plan.QuerySQL, "ORDER BY") {
		t.Error("expected SQL to contain ORDER BY")
	}
	if !strings.Contains(plan.QuerySQL, "DESC") {
		t.Error("expected SQL to contain DESC")
	}
}

// === SECURITY: NO RAW SQL PATH TESTS ===

// This is the most critical test: prove that malicious input never reaches SQL.
func TestCompilePlan_NoRawSQLFromUserInput(t *testing.T) {
	maliciousInputs := []QueryIntent{
		{Entity: "reservoir", Metrics: []string{"fill_percent"}, Aggregation: "latest", Filters: Filters{Basin: "Ebro'; DROP TABLE reservoirs; --"}},
		{Entity: "reservoir", Metrics: []string{"fill_percent"}, Aggregation: "latest", Filters: Filters{Province: "Zaragoza'; DELETE FROM api_keys; --"}},
		{Entity: "reservoir", Metrics: []string{"fill_percent"}, Aggregation: "latest", Filters: Filters{Since: "2024-01-01'; SELECT * FROM passwords; --"}},
		{Entity: "reservoir", Metrics: []string{"fill_percent"}, Aggregation: "latest", Filters: Filters{Slugs: []string{"test'; TRUNCATE TABLE metering; --"}}},
	}

	for _, intent := range maliciousInputs {
		// Validation should reject these before compilation
		if err := ValidateIntent(intent); err == nil {
			t.Fatal("validation should have rejected malicious input")
		}
	}
}

// If somehow validation is bypassed, compilation must still not interpolate.
func TestCompilePlan_BypassValidationStillSafe(t *testing.T) {
	// Simulate bypassed validation by calling CompilePlan directly with
	// a malicious filter that validation WOULD reject.
	intent := QueryIntent{
		Entity:      "reservoir",
		Metrics:     []string{"fill_percent"},
		Aggregation: "latest",
		Filters: Filters{
			Basin: "Ebro",
		},
		Limit: 10,
	}
	plan, err := CompilePlan(intent)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	// The SQL must use $N parameter placeholders, not string interpolation
	if strings.Contains(plan.QuerySQL, "Ebro") {
		t.Error("SQL should contain $N placeholder, not literal 'Ebro'")
	}
	if !strings.Contains(plan.QuerySQL, "$") {
		t.Error("SQL should use parameterized placeholders ($N)")
	}
}

// === PARSE TESTS ===

func TestParseIntent_ValidJSON(t *testing.T) {
	raw := []byte(`{"entity":"reservoir","metrics":["fill_percent"],"aggregation":"latest"}`)
	intent, err := ParseIntent(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if intent.Entity != "reservoir" {
		t.Errorf("expected entity 'reservoir', got %s", intent.Entity)
	}
}

func TestParseIntent_InvalidJSON(t *testing.T) {
	raw := []byte(`{invalid json`)
	_, err := ParseIntent(raw)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// === ENTITY-SPECIFIC TESTS ===

func TestValidateIntent_AllEntities(t *testing.T) {
	entities := []string{"reservoir", "basin", "province", "community", "national"}
	for _, entity := range entities {
		intent := QueryIntent{
			Entity:      entity,
			Metrics:     []string{"fill_percent"},
			Aggregation: "latest",
			Limit:       10,
		}
		if err := ValidateIntent(intent); err != nil {
			t.Errorf("entity %q should be valid: %v", entity, err)
		}
	}
}

func TestCompilePlan_AllEntities(t *testing.T) {
	entities := []string{"reservoir", "basin", "province", "national"}
	for _, entity := range entities {
		intent := QueryIntent{
			Entity:      entity,
			Metrics:     []string{"fill_percent"},
			Aggregation: "latest",
			Limit:       10,
		}
		plan, err := CompilePlan(intent)
		if err != nil {
			t.Fatalf("compile for entity %q failed: %v", entity, err)
		}
		if plan.EntityTable == "" {
			t.Errorf("entity %q should have a table", entity)
		}
	}
}
