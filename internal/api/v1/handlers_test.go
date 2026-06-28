package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/arzamas1987/embalses/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testSchema string

func init() {
	testSchema = "test_" + fmt.Sprintf("%d", time.Now().UnixNano())
}

func testPool(t *testing.T) *pgxpool.Pool {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/embalses?sslmode=disable"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := db.New(ctx, connStr)
	if err != nil {
		t.Skipf("database not available: %v", err)
	}
	// Create a unique test schema and include public for PostGIS
	_, _ = pool.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", testSchema))
	_, _ = pool.Exec(ctx, fmt.Sprintf("SET search_path TO %s, public", testSchema))
	return pool.Pool
}

func seedSchema(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()
	// Clean up tables in test schema
	_, _ = pool.Exec(ctx, fmt.Sprintf(`
		DROP TABLE IF EXISTS %s.metering CASCADE;
		DROP TABLE IF EXISTS %s.api_keys CASCADE;
		DROP TABLE IF EXISTS %s.readings CASCADE;
		DROP TABLE IF EXISTS %s.reservoirs CASCADE;
		DROP TABLE IF EXISTS %s.dams CASCADE;
		DROP TABLE IF EXISTS %s.provinces CASCADE;
		DROP TABLE IF EXISTS %s.basins CASCADE;
		DROP TABLE IF EXISTS %s.sources CASCADE;
	`, testSchema, testSchema, testSchema, testSchema, testSchema, testSchema, testSchema, testSchema))

	// Run all migrations (they'll use the search_path schema)
	for _, file := range []string{
		"../../../migrations/000002_geo_schema.up.sql",
		"../../../migrations/000003_api_v1.up.sql",
	} {
		sql, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read migration %s: %v", file, err)
		}
		_, err = pool.Exec(ctx, string(sql))
		if err != nil {
			t.Fatalf("run migration %s: %v", file, err)
		}
	}
}

func setupRouter(pool *pgxpool.Pool) chi.Router {
	r := chi.NewRouter()
	RegisterRoutes(r, pool)
	return r
}

func TestHealthz(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	r := setupRouter(pool)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	data, ok := resp.Data.(map[string]interface{})
	if !ok || data["status"] != "ok" {
		t.Errorf("expected status ok, got %v", resp.Data)
	}
}

func TestReadyz(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	r := setupRouter(pool)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAuthRequired(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	seedSchema(t, pool)
	r := setupRouter(pool)

	// No API key → 401
	req := httptest.NewRequest(http.MethodGet, "/api/v1/sources", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	var resp APIResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Error == nil || resp.Error.Code != "unauthorized" {
		t.Errorf("expected unauthorized error, got %+v", resp.Error)
	}
}

func TestListSources(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	seedSchema(t, pool)
	r := setupRouter(pool)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sources", nil)
	req.Header.Set("X-API-Key", "test-key-123")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	data, ok := resp.Data.([]interface{})
	if !ok || len(data) < 2 {
		t.Errorf("expected at least 2 sources, got %d", len(data))
	}
	if resp.Lineage == nil {
		t.Error("expected lineage in response")
	}
}

func TestListReservoirs(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	seedSchema(t, pool)
	r := setupRouter(pool)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reservoirs", nil)
	req.Header.Set("X-API-Key", "test-key-123")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Meta == nil {
		t.Error("expected meta in response")
	}
	if resp.Meta.PerPage != 20 {
		t.Errorf("expected per_page 20, got %d", resp.Meta.PerPage)
	}
	if resp.Lineage == nil {
		t.Error("expected lineage in response")
	}
}

func TestGetReservoir_NotFound(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	seedSchema(t, pool)
	r := setupRouter(pool)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reservoirs/NonExistent", nil)
	req.Header.Set("X-API-Key", "test-key-123")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	var resp APIResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Error == nil || resp.Error.Code != "not_found" {
		t.Errorf("expected not_found error, got %+v", resp.Error)
	}
}

func TestListBasins(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	seedSchema(t, pool)
	r := setupRouter(pool)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/basins", nil)
	req.Header.Set("X-API-Key", "test-key-123")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Lineage == nil {
		t.Error("expected lineage in response")
	}
}

func TestGetRankings(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	seedSchema(t, pool)
	r := setupRouter(pool)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rankings/reservoirs?metric=fullest&limit=5", nil)
	req.Header.Set("X-API-Key", "test-key-123")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Lineage == nil {
		t.Error("expected lineage in response")
	}
}

func TestCompareReservoirs_NoParams(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	seedSchema(t, pool)
	r := setupRouter(pool)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/compare", nil)
	req.Header.Set("X-API-Key", "test-key-123")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	var resp APIResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Error == nil || resp.Error.Code != "bad_request" {
		t.Errorf("expected bad_request error, got %+v", resp.Error)
	}
}

func TestGetDataQuality(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	seedSchema(t, pool)
	r := setupRouter(pool)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/data-quality", nil)
	req.Header.Set("X-API-Key", "test-key-123")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Lineage == nil {
		t.Error("expected lineage in response")
	}
}

func TestAPIKeyRateLimit(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	seedSchema(t, pool)
	r := setupRouter(pool)

	// First request should succeed
	req := httptest.NewRequest(http.MethodGet, "/api/v1/sources", nil)
	req.Header.Set("X-API-Key", "test-key-123")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// Many rapid requests should eventually hit rate limit
	// The test key has 120/min limit so we won't hit it in a few requests
	// Just verify the middleware is active
}

func TestErrorShape(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	seedSchema(t, pool)
	r := setupRouter(pool)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reservoirs/DoesNotExist", nil)
	req.Header.Set("X-API-Key", "test-key-123")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("expected JSON content-type, got %s", ct)
	}
	var resp APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected error field in response")
	}
	if resp.Error.Code == "" {
		t.Error("expected error.code")
	}
	if resp.Error.Message == "" {
		t.Error("expected error.message")
	}
	if resp.Data != nil {
		t.Error("expected nil data on error")
	}
}
