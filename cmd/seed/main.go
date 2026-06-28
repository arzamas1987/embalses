package main

import (
	"context"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/arzamas1987/embalses/internal/config"
	"github.com/arzamas1987/embalses/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()
	pool, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("DB connection: %v", err)
	}
	defer pool.Close()

	seed(ctx, pool.Pool)
}

func seed(ctx context.Context, pool *pgxpool.Pool) {
	log.Println("=== Embalse Seeder: Full 6-month dataset ===")

	// 1. Seed sources
	_, _ = pool.Exec(ctx, `
		INSERT INTO sources (name, organism, licence, attribution, url) VALUES
		('MITECO', 'Ministerio para la Transición Ecológica y el Reto Demográfico', 'Ley 37/2007 + RD 1495/2011', 'Fuente: MITECO', 'https://www.miteco.gob.es'),
		('SNCZI', 'Confederación Hidrográfica del Ebro', 'Datos abiertos MITECO', 'Fuente: SNCZI / MITECO', 'https://www.chebro.es'),
		('IGN', 'Instituto Geográfico Nacional', 'CC-BY 4.0', 'Fuente: IGN', 'https://www.ign.es')
		ON CONFLICT DO NOTHING
	`)

	var sourceID int
	_ = pool.QueryRow(ctx, "SELECT id FROM sources WHERE name = 'MITECO' LIMIT 1").Scan(&sourceID)
	if sourceID == 0 {
		sourceID = 1
	}

	// 2. Seed basins
	basins := []struct{ name, code string }{
		{"Ebro", "EBRO"},
		{"Duero", "DUERO"},
		{"Tajo", "TAJO"},
		{"Guadiana", "GUADIANA"},
		{"Guadalquivir", "GUADALQUIVIR"},
		{"Júcar", "JUCAR"},
		{"Miño-Sil", "MINO"},
		{"Tinto, Odiel y Piedras", "TOP"},
		{"Catalunya", "CAT"},
		{"Sur de España", "SUR"},
		{"Canarias", "CAN"},
		{"Baleares", "BAL"},
		{"Cuencas Mediterráneas Andaluzas", "CMA"},
		{"Ceuta y Melilla", "CEMEL"},
		{"Segura", "SEGURA"},
	}

	basinIDs := make(map[string]int)
	for _, b := range basins {
		var id int
		err := pool.QueryRow(ctx, `
			INSERT INTO basins (name, code) VALUES ($1, $2)
			ON CONFLICT (name) DO UPDATE SET code = EXCLUDED.code
			RETURNING id
		`, b.name, b.code).Scan(&id)
		if err != nil {
			log.Printf("Basin insert error: %v", err)
			continue
		}
		basinIDs[b.name] = id
	}

	// 3. Seed provinces
	provinces := []string{
		"Zaragoza", "Soria", "Huesca", "Teruel", "Navarra", "La Rioja",
		"Valladolid", "Zamora", "Salamanca", "Ávila", "Segovia", "Soria",
		"Toledo", "Madrid", "Guadalajara", "Cuenca", "Ciudad Real", "Albacete",
		"Badajoz", "Cáceres", "Mérida",
		"Sevilla", "Córdoba", "Jaén", "Granada", "Almería", "Málaga", "Huelva",
		"Valencia", "Alicante", "Castellón", "Murcia", "Albacete",
		"Ourense", "Lugo", "Pontevedra", "A Coruña", "León", "Asturias",
		"Barcelona", "Girona", "Lleida", "Tarragona", "Baleares",
		"Santa Cruz de Tenerife", "Las Palmas", "Ceuta", "Melilla",
	}

	provinceIDs := make(map[string]int)
	for _, p := range provinces {
		var id int
		err := pool.QueryRow(ctx, `
			INSERT INTO provinces (name) VALUES ($1)
			ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, p).Scan(&id)
		if err != nil {
			log.Printf("Province insert error: %v", err)
			continue
		}
		provinceIDs[p] = id
	}

	// 4. Seed reservoirs with realistic Spanish data
	reservoirs := []struct {
		name       string
		externalID string
		basin      string
		province   string
		capacity   float64
	}{
		{"Embalse de Mequinenza", "MEQUINENZA", "Ebro", "Zaragoza", 1530.0},
		{"Embalse de la Serena", "SERENA", "Guadiana", "Badajoz", 3232.0},
		{"Embalse de Sau", "SAU", "Catalunya", "Barcelona", 151.0},
		{"Embalse de Alcántara", "ALCANTARA", "Tajo", "Cáceres", 3167.0},
		{"Embalse de Valdecañas", "VALDECANAS", "Tajo", "Cáceres", 1488.0},
		{"Embalse de Bornos", "BORNOS", "Guadalquivir", "Cádiz", 245.0},
		{"Embalse de Canelles", "CANELLES", "Catalunya", "Lleida", 201.0},
		{"Embalse de El Atazar", "ATAZAR", "Tajo", "Madrid", 425.0},
		{"Embalse de Buendía", "BUENDIA", "Tajo", "Cuenca", 1638.0},
		{"Embalse de Rialb", "RIALB", "Catalunya", "Lleida", 402.0},
		{"Embalse de Ebro", "EBRO-RES", "Ebro", "Cantabria", 55.0},
		{"Embalse de Rules", "RULES", "Guadalquivir", "Granada", 108.0},
		{"Embalse de La Viñuela", "VINUELA", "Sur de España", "Málaga", 170.0},
		{"Embalse de San Juan", "SANJUAN", "Tajo", "Madrid", 138.0},
		{"Embalse de Gabriel y Galán", "GABYGALAN", "Tajo", "Cáceres", 900.0},
		{"Embalse de Cíjara", "CIJARA", "Guadiana", "Badajoz", 1457.0},
		{"Embalse de Orellana", "ORELLANA", "Guadiana", "Badajoz", 808.0},
		{"Embalse de García Sola", "GARCIASOLA", "Guadiana", "Badajoz", 555.0},
		{"Embalse de Zújar", "ZUJAR", "Guadiana", "Badajoz", 289.0},
		{"Embalse del Negratín", "NEGRATIN", "Guadalquivir", "Granada", 490.0},
	}

	reservoirIDs := make(map[string]int)
	for _, r := range reservoirs {
		bID := basinIDs[r.basin]
		pID := provinceIDs[r.province]
		if bID == 0 || pID == 0 {
			log.Printf("Missing basin/province for %s", r.name)
			continue
		}

		var id int
		err := pool.QueryRow(ctx, `
			INSERT INTO reservoirs (name, external_id, basin_id, province_id, capacity_hm3, source_id)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (name) DO UPDATE SET
				external_id = EXCLUDED.external_id,
				basin_id = EXCLUDED.basin_id,
				province_id = EXCLUDED.province_id,
				capacity_hm3 = EXCLUDED.capacity_hm3
			RETURNING id
		`, r.name, r.externalID, bID, pID, r.capacity, sourceID).Scan(&id)
		if err != nil {
			log.Printf("Reservoir insert error for %s: %v", r.name, err)
			continue
		}
		reservoirIDs[r.externalID] = id
		log.Printf("  Reservoir: %s (capacity: %.0f hm³)", r.name, r.capacity)
	}

	// 5. Seed dams
	for _, r := range reservoirs {
		bID := basinIDs[r.basin]
		pID := provinceIDs[r.province]
		if bID == 0 || pID == 0 {
			continue
		}
		_, _ = pool.Exec(ctx, `
			INSERT INTO dams (name, capacity_hm3, basin, province, basin_id, province_id, source_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (name) DO NOTHING
		`, "Presa de "+r.name, r.capacity, r.basin, r.province, bID, pID, sourceID)
	}

	// 6. Seed 6 months of weekly readings with seasonal patterns
	endDate := time.Now().Truncate(24 * time.Hour)
	startDate := endDate.AddDate(0, -6, 0)
	log.Printf("Generating readings from %s to %s...", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	var totalCount int
	for _, r := range reservoirs {
		rID := reservoirIDs[r.externalID]
		if rID == 0 {
			continue
		}

		// Each reservoir has a unique base fill level based on its name hash
		baseFill := 35 + float64(len(r.name)%45)
		capacity := r.capacity
		if capacity == 0 {
			capacity = 100 + rand.Float64()*500
		}

		for d := startDate; d.Before(endDate) || d.Equal(endDate); d = d.AddDate(0, 0, 7) {
			month := d.Month()
			var seasonal float64
			switch {
			case month == 12 || month == 1 || month == 2:
				seasonal = 15 + rand.Float64()*10 // Winter — moderate high
			case month == 3 || month == 4 || month == 5:
				seasonal = 30 + rand.Float64()*15 // Spring — peak fill from rain/snowmelt
			case month == 6 || month == 7 || month == 8:
				seasonal = -30 - rand.Float64()*15 // Summer — drought, low levels
			case month == 9 || month == 10 || month == 11:
				seasonal = -5 + rand.Float64()*10 // Autumn — recovering
			}

			weekly := (rand.Float64() - 0.5) * 8
			fillPct := math.Max(8, math.Min(100, baseFill+seasonal+weekly))
			volume := capacity * fillPct / 100
			variation := weekly

			_, err := pool.Exec(ctx, `
				INSERT INTO readings (reservoir_id, source_id, observed_at, volume_hm3, capacity_hm3, fill_pct, weekly_variation_hm3, is_provisional, is_official, fetched_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, false, true, NOW())
				ON CONFLICT (reservoir_id, observed_at, source_id) DO NOTHING
			`, rID, sourceID, d.Format("2006-01-02"), volume, capacity, fillPct, variation)
			if err != nil {
				log.Printf("  Insert error for %s on %s: %v", r.name, d.Format("2006-01-02"), err)
			} else {
				totalCount++
			}
		}
	}

	log.Printf("=== DONE ===")
	log.Printf("  Basins:     %d", len(basinIDs))
	log.Printf("  Provinces:  %d", len(provinceIDs))
	log.Printf("  Reservoirs: %d", len(reservoirIDs))
	log.Printf("  Readings:   %d (6 months weekly, ~26 per reservoir)", totalCount)
	log.Printf("  Date range: %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
}
