package geocoding

var regionTranslations = map[string]string{
	"Rhineland-Palatinate":   "Rheinland-Pfalz",
	"Bavaria":                "Bayern",
	"North Rhine-Westphalia": "Nordrhein-Westfalen",
	"Lower Saxony":           "Niedersachsen",
	"Saxony":                 "Sachsen",
	"Saxony-Anhalt":          "Sachsen-Anhalt",
}

// GetRegionTranslation returns the german translation for some regions.
// This is because google gives english names in some cases.
func GetRegionTranslation(region *string) string {
	if region == nil {
		return ""
	}

	translation, ok := regionTranslations[*region]
	if ok {
		return translation
	}
	return *region
}
