package geo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunSpatialJoins executes PostGIS spatial joins to link:
// - reservoirs → basins (ST_Within)
// - reservoirs → provinces (ST_Within)
// - dams → basins (by name match)
// - dams → reservoirs (by name proximity)
func RunSpatialJoins(ctx context.Context, pool *pgxpool.Pool) error {
	// Join reservoirs to basins
	_, err := pool.Exec(ctx, `
		UPDATE reservoirs r
		SET basin_id = b.id
		FROM basins b
		WHERE r.basin_id IS NULL
			AND ST_Within(r.geometry, b.geometry)
	`)
	if err != nil {
		return fmt.Errorf("join reservoirs to basins: %w", err)
	}

	// Join reservoirs to provinces
	_, err = pool.Exec(ctx, `
		UPDATE reservoirs r
		SET province_id = p.id
		FROM provinces p
		WHERE r.province_id IS NULL
			AND ST_Within(r.geometry, p.geometry)
	`)
	if err != nil {
		return fmt.Errorf("join reservoirs to provinces: %w", err)
	}

	// Join dams to basins (by basin name match)
	_, err = pool.Exec(ctx, `
		UPDATE dams d
		SET basin_id = b.id
		FROM basins b
		WHERE d.basin_id IS NULL
			AND lower(d.basin) = lower(b.name)
	`)
	if err != nil {
		return fmt.Errorf("join dams to basins: %w", err)
	}

	// Join dams to provinces (by province name match)
	_, err = pool.Exec(ctx, `
		UPDATE dams d
		SET province_id = p.id
		FROM provinces p
		WHERE d.province_id IS NULL
			AND lower(d.province) = lower(p.name)
	`)
	if err != nil {
		return fmt.Errorf("join dams to provinces: %w", err)
	}

	// Link reservoirs to dams (by name similarity)
	_, err = pool.Exec(ctx, `
		UPDATE reservoirs r
		SET dam_id = d.id
		FROM dams d
		WHERE r.dam_id IS NULL
			AND (
				lower(r.name) LIKE '%' || lower(d.name) || '%'
				OR lower(d.name) LIKE '%' || lower(r.name) || '%'
			)
	`)
	if err != nil {
		return fmt.Errorf("link reservoirs to dams: %w", err)
	}

	return nil
}

// GetReservoirWithSpatialInfo returns a reservoir with its basin and province info.
func GetReservoirWithSpatialInfo(ctx context.Context, pool *pgxpool.Pool, name string) (ReservoirSpatialInfo, error) {
	var info ReservoirSpatialInfo
	err := pool.QueryRow(ctx, `
		SELECT 
			r.name,
			b.name AS basin_name,
			p.name AS province_name,
			s.name AS source_name,
			s.attribution
		FROM reservoirs r
		LEFT JOIN basins b ON r.basin_id = b.id
		LEFT JOIN provinces p ON r.province_id = p.id
		LEFT JOIN sources s ON r.source_id = s.id
		WHERE r.name = $1
	`, name).Scan(
		&info.ReservoirName,
		&info.BasinName,
		&info.ProvinceName,
		&info.SourceName,
		&info.Attribution,
	)
	if err != nil {
		return info, fmt.Errorf("query reservoir: %w", err)
	}
	return info, nil
}

// ReservoirSpatialInfo holds joined reservoir spatial data.
type ReservoirSpatialInfo struct {
	ReservoirName string
	BasinName     string
	ProvinceName  string
	SourceName    string
	Attribution   string
}
