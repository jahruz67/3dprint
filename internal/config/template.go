package config

// PrinterTemplate returns a fresh, intentionally incomplete JSON draft. Unknown
// hardware values are null rather than invented capabilities. Nulls and empty
// identity/evidence fields must be replaced before profile validation can pass.
// false optional capability flags mean "not verified / do not use".
// Flow is a configurable operating cap, prefilled at 4 mm³/s, with no certification step.
func PrinterTemplate() map[string]any {
	return map[string]any{
		"schemaVersion": SchemaVersion,
		"id":            "",
		"name":          "",
		"manufacturer":  "",
		"model":         "",
		"firmware": map[string]any{
			"family": nil, "version": nil, "gcodeFlavor": nil, "extrusionMode": nil,
			"supportedCommands": []string{}, "unsupportedCommands": []string{},
			"startGcode": []string{}, "endGcode": []string{},
		},
		"buildVolume": map[string]any{
			"width": nil, "depth": nil, "height": nil, "shape": nil, "origin": nil,
			"excludedRegions": []any{},
		},
		"motion": map[string]any{
			"maxSpeed": nil, "maxTravelSpeed": nil, "maxAcceleration": nil,
		},
		"extruders": []any{map[string]any{
			"nozzleDiameter": nil, "filamentDiameter": nil, "maxVolumetricFlow": 4.0,
			"retractionLength": nil, "retractionSpeed": nil, "type": nil,
		}},
		"thermal": map[string]any{
			"minNozzle": nil, "maxNozzle": nil, "minBed": nil, "maxBed": nil,
			"hasHeatedBed": nil, "partCooling": nil,
		},
		"capabilities": map[string]any{
			"autoBedLeveling": false, "firmwareRetraction": false, "pressureAdvance": false,
		},
		"provenance": map[string]any{
			"buildVolume": "", "motion": "", "thermal": "",
		},
	}
}
