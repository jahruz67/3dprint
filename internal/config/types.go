// Package config implements portable, strictly validated machine, material and preference profiles.
package config

type Printer struct {
	SchemaVersion string            `json:"schemaVersion"`
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Manufacturer  string            `json:"manufacturer"`
	Model         string            `json:"model"`
	Firmware      Firmware          `json:"firmware"`
	BuildVolume   BuildVolume       `json:"buildVolume"`
	Motion        Motion            `json:"motion"`
	Extruders     []Extruder        `json:"extruders"`
	Thermal       Thermal           `json:"thermal"`
	Capabilities  Capabilities      `json:"capabilities"`
	Provenance    map[string]string `json:"provenance"`
}
type Firmware struct {
	Family              string   `json:"family"`
	Version             string   `json:"version"`
	GCodeFlavor         string   `json:"gcodeFlavor"`
	ExtrusionMode       string   `json:"extrusionMode"`
	SupportedCommands   []string `json:"supportedCommands"`
	UnsupportedCommands []string `json:"unsupportedCommands"`
	StartGCode          []string `json:"startGcode"`
	EndGCode            []string `json:"endGcode"`
}
type BuildVolume struct {
	Width           float64  `json:"width"`
	Depth           float64  `json:"depth"`
	Height          float64  `json:"height"`
	Shape           string   `json:"shape"`
	Origin          string   `json:"origin"`
	ExcludedRegions []Region `json:"excludedRegions"`
}
type Region struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Width float64 `json:"width"`
	Depth float64 `json:"depth"`
}
type Motion struct {
	MaxSpeed        float64 `json:"maxSpeed"`
	MaxTravelSpeed  float64 `json:"maxTravelSpeed"`
	MaxAcceleration float64 `json:"maxAcceleration"`
}
type Extruder struct {
	NozzleDiameter    float64 `json:"nozzleDiameter"`
	FilamentDiameter  float64 `json:"filamentDiameter"`
	MaxVolumetricFlow float64 `json:"maxVolumetricFlow"`
	RetractionLength  float64 `json:"retractionLength"`
	RetractionSpeed   float64 `json:"retractionSpeed"`
	FlowVerified      bool    `json:"flowVerified,omitempty"` // Legacy metadata only; never gates import or slicing.
	Type              string  `json:"type"`
}
type Thermal struct {
	MinNozzle    float64 `json:"minNozzle"`
	MaxNozzle    float64 `json:"maxNozzle"`
	MinBed       float64 `json:"minBed"`
	MaxBed       float64 `json:"maxBed"`
	HasHeatedBed bool    `json:"hasHeatedBed"`
	PartCooling  bool    `json:"partCooling"`
}
type Capabilities struct {
	AutoBedLeveling    bool `json:"autoBedLeveling"`
	FirmwareRetraction bool `json:"firmwareRetraction"`
	PressureAdvance    bool `json:"pressureAdvance"`
}
type Filament struct {
	SchemaVersion     string   `json:"schemaVersion"`
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Material          string   `json:"material"`
	Density           float64  `json:"density"`
	NozzleTemperature float64  `json:"nozzleTemperature"`
	BedTemperature    float64  `json:"bedTemperature"`
	MaxVolumetricFlow float64  `json:"maxVolumetricFlow"`
	MinLayerHeight    float64  `json:"minLayerHeight"`
	MaxLayerHeight    float64  `json:"maxLayerHeight"`
	FanPercent        int      `json:"fanPercent"`
	Notes             []string `json:"notes,omitempty"`
}
type Preferences struct {
	SchemaVersion    string     `json:"schemaVersion"`
	Mode             string     `json:"mode"`
	Budget           int        `json:"budget"`
	MinWalls         int        `json:"minWalls"`
	MinInfill        float64    `json:"minInfill"`
	Supports         bool       `json:"supports"`
	LoadDirection    [3]float64 `json:"loadDirection,omitempty"`
	SupportMode      string     `json:"supportMode,omitempty"`
	SupportStyle     string     `json:"supportStyle,omitempty"`
	SupportPlacement string     `json:"supportPlacement,omitempty"`
	SupportAngle     float64    `json:"supportAngle,omitempty"`
}
