package ign

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Ingest inserts IGN features into the database.
func Ingest(ctx context.Context, pool *pgxpool.Pool, basins []BasinFeature, provinces []ProvinceFeature, reservoirs []ReservoirFeature) error {
	ignSourceID := 0
	err := pool.QueryRow(ctx, "SELECT id FROM sources WHERE name = 'IGN'").Scan(&ignSourceID)
	if err != nil {
		return fmt.Errorf("lookup IGN source: %w", err)
	}

	for _, b := range basins {
		_, err := pool.Exec(ctx, `
			INSERT INTO basins (name, code, geometry, source_id)
			VALUES ($1, $2, ST_GeomFromText($3, 4326), $4)
			ON CONFLICT (name) DO UPDATE SET
				code = EXCLUDED.code,
				geometry = EXCLUDED.geometry,
				updated_at = NOW()
		`, b.Name, b.Code, b.WKT, ignSourceID)
		if err != nil {
			return fmt.Errorf("insert basin %s: %w", b.Name, err)
		}
	}

	for _, p := range provinces {
		_, err := pool.Exec(ctx, `
			INSERT INTO provinces (name, code, geometry, source_id)
			VALUES ($1, $2, ST_GeomFromText($3, 4326), $4)
			ON CONFLICT (name) DO UPDATE SET
				code = EXCLUDED.code,
				geometry = EXCLUDED.geometry,
				updated_at = NOW()
		`, p.Name, p.Code, p.WKT, ignSourceID)
		if err != nil {
			return fmt.Errorf("insert province %s: %w", p.Name, err)
		}
	}

	for _, r := range reservoirs {
		_, err := pool.Exec(ctx, `
			INSERT INTO reservoirs (name, external_id, geometry, capacity_hm3, source_id)
			VALUES ($1, $2, ST_GeomFromText($3, 4326), $4, $5)
			ON CONFLICT (name) DO UPDATE SET
				external_id = EXCLUDED.external_id,
				geometry = EXCLUDED.geometry,
				updated_at = NOW()
		`, r.Name, r.ExternalID, r.WKT, 0, ignSourceID)
		if err != nil {
			return fmt.Errorf("insert reservoir %s: %w", r.Name, err)
		}
	}

	return nil
}
