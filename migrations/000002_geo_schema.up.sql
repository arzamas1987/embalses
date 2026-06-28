CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS sources (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    organism TEXT NOT NULL,
    licence TEXT NOT NULL,
    attribution TEXT NOT NULL,
    url TEXT,
    last_fetched_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS basins (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    code TEXT,
    geometry GEOMETRY(MultiPolygon, 4326),
    source_id INTEGER REFERENCES sources(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS provinces (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    code TEXT,
    geometry GEOMETRY(MultiPolygon, 4326),
    source_id INTEGER REFERENCES sources(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS dams (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    external_id TEXT,
    risk_category TEXT,
    river TEXT,
    municipality TEXT,
    province TEXT,
    basin TEXT,
    basin_id INTEGER REFERENCES basins(id),
    province_id INTEGER REFERENCES provinces(id),
    geometry GEOMETRY(Point, 4326),
    basin_area_km2 NUMERIC,
    capacity_hm3 NUMERIC,
    nmn_elevation NUMERIC,
    dam_type TEXT,
    dam_height_m NUMERIC,
    source_id INTEGER REFERENCES sources(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS reservoirs (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    external_id TEXT,
    basin_id INTEGER REFERENCES basins(id),
    province_id INTEGER REFERENCES provinces(id),
    dam_id INTEGER REFERENCES dams(id),
    geometry GEOMETRY(Polygon, 4326),
    capacity_hm3 NUMERIC,
    source_id INTEGER REFERENCES sources(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dams_basin ON dams(basin_id);
CREATE INDEX IF NOT EXISTS idx_dams_province ON dams(province_id);
CREATE INDEX IF NOT EXISTS idx_dams_geometry ON dams USING GIST(geometry);
CREATE INDEX IF NOT EXISTS idx_basins_geometry ON basins USING GIST(geometry);
CREATE INDEX IF NOT EXISTS idx_provinces_geometry ON provinces USING GIST(geometry);
CREATE INDEX IF NOT EXISTS idx_reservoirs_basin ON reservoirs(basin_id);
CREATE INDEX IF NOT EXISTS idx_reservoirs_province ON reservoirs(province_id);
CREATE INDEX IF NOT EXISTS idx_reservoirs_geometry ON reservoirs USING GIST(geometry);

INSERT INTO sources (name, organism, licence, attribution, url) VALUES
    ('SNCZI', 'MITECO - Subdirección General de Dominio Público Hidráulico e Infraestructuras', 'Ley 37/2007 + RD 1495/2011 (PSI reuse)', 'SNCZI – Inventario de Presas y Emballes, MITECO', 'https://www.miteco.gob.es/es/cartografia-y-sig/ide/descargas/agua/inventario-presas-embalses.html'),
    ('IGN', 'Instituto Geográfico Nacional (IGN) / Centro Nacional de Información Geográfica (CNIG)', 'CC-BY 4.0 (Orden FOM/2807/2015)', 'CC-BY 4.0 scne.es / © Instituto Geográfico Nacional', 'https://centrodedescargas.cnig.es/CentroDescargas/home')
ON CONFLICT (name) DO NOTHING;
