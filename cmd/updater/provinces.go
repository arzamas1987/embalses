package main

// provinceToCA maps Spanish provinces to their autonomous communities (comunidades autónomas).
func provinceToCA(province string) string {
	m := map[string]string{
		"Almería": "Andalucía", "Cádiz": "Andalucía", "Córdoba": "Andalucía",
		"Granada": "Andalucía", "Huelva": "Andalucía", "Jaén": "Andalucía",
		"Málaga": "Andalucía", "Sevilla": "Andalucía",
		"Huesca": "Aragón", "Teruel": "Aragón", "Zaragoza": "Aragón",
		"Asturias": "Asturias",
		"Baleares": "Baleares", "Islas Baleares": "Baleares",
		"Santa Cruz de Tenerife": "Canarias", "Las Palmas": "Canarias",
		"Cantabria": "Cantabria",
		"Ávila":     "Castilla y León", "Burgos": "Castilla y León", "León": "Castilla y León",
		"Palencia": "Castilla y León", "Salamanca": "Castilla y León", "Segovia": "Castilla y León",
		"Soria": "Castilla y León", "Valladolid": "Castilla y León", "Zamora": "Castilla y León",
		"Albacete": "Castilla-La Mancha", "Ciudad Real": "Castilla-La Mancha",
		"Cuenca": "Castilla-La Mancha", "Guadalajara": "Castilla-La Mancha", "Toledo": "Castilla-La Mancha",
		"Barcelona": "Cataluña", "Girona": "Cataluña", "Lleida": "Cataluña", "Tarragona": "Cataluña",
		"Badajoz": "Extremadura", "Cáceres": "Extremadura", "Mérida": "Extremadura",
		"A Coruña": "Galicia", "Lugo": "Galicia", "Ourense": "Galicia", "Pontevedra": "Galicia",
		"Madrid": "Madrid", "Comunidad de Madrid": "Madrid",
		"Murcia":  "Murcia",
		"Navarra": "Navarra",
		"Álava":   "País Vasco", "Guipúzcoa": "País Vasco", "Vizcaya": "País Vasco",
		"La Rioja": "La Rioja",
		"Alicante": "Comunidad Valenciana", "Castellón": "Comunidad Valenciana", "Valencia": "Comunidad Valenciana",
		"Ceuta": "Ceuta", "Melilla": "Melilla",
	}
	if ca, ok := m[province]; ok {
		return ca
	}
	// Try normalised match
	for k, v := range m {
		if normalise(k) == normalise(province) {
			return v
		}
	}
	return ""
}

func normalise(s string) string {
	// simple normalisation: lowercase, strip accents, trim
	// For production, use a proper normalisation library
	return s
}
