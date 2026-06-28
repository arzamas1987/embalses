package snczi

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Ingest inserts dam features into the database.
// Coordinates are assumed to be in ETRS89 (EPSG:25830) and are reprojected to EPSG:4326 via PostGIS ST_Transform.
func Ingest(ctx context.Context, pool *pgxpool.Pool, features []DamFeature) error {
	if len(features) == 0 {
		return nil
	}

	sql := `
		INSERT INTO dams (
			name, external_id, province, basin, risk_category, river, municipality,
			geometry, basin_area_km2, capacity_hm3, nmn_elevation, dam_type, dam_height_m,
			source_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			ST_Transform(ST_SetSRID(ST_MakePoint($8, $9), 25830), 4326),
			$10, $11, $12, $13, $14,
			(SELECT id FROM sources WHERE name = 'SNCZI')
		)
		ON CONFLICT (name) DO UPDATE SET
			external_id = EXCLUDED.external_id,
			province = EXCLUDED.province,
			basin = EXCLUDED.basin,
			risk_category = EXCLUDED.risk_category,
			river = EXCLUDED.river,
			municipality = EXCLUDED.municipality,
			geometry = EXCLUDED.geometry,
			basin_area_km2 = EXCLUDED.basin_area_km2,
			capacity_hm3 = EXCLUDED.capacity_hm3,
			nmn_elevation = EXCLUDED.nmn_elevation,
			dam_type = EXCLUDED.dam_type,
			dam_height_m = EXCLUDED.dam_height_m,
			updated_at = NOW()
	`

	for _, f := range features {
		_, err := pool.Exec(ctx, sql,
			f.Name, f.ExternalID, f.Province, f.Basin, f.RiskCategory, f.River, f.Municipality,
			f.Lon, f.Lat,
			f.BasinAreaKm2, f.CapacityHM3, f.NMNElevation, f.DamType, f.DamHeightM,
		)
		if err != nil {
			return fmt.Errorf("insert dam %s: %w", f.Name, err)
		}
	}

	return nil
}
