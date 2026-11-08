package optimizer

import (
	"github.com/forge-slicer/forge/internal/config"
	"math"
)

// SupportMode preserves version-1 profiles which only specified Supports.
func SupportMode(p config.Preferences) string {
	if p.SupportMode != "" {
		return p.SupportMode
	}
	if p.Supports {
		return "auto"
	}
	return "off"
}

// SupportPolicy is applied after each layer-height change so contact gaps follow
// the final layer height. These are explicit starting recipes, not measured
// removal-force or strength guarantees.
func SupportPolicy(s Settings, p config.Preferences, nozzle float64) Settings {
	mode := SupportMode(p)
	s.SupportMode = mode
	if mode == "off" {
		s.Supports = false
	}
	s.SupportStyle = "organic"
	s.SupportContactZ = math.Max(.12, .65*s.LayerHeight)
	s.SupportXYDistance = math.Max(.25, .7*nozzle)
	s.SupportSpacing = 3
	s.SupportInterfaceLayers = 2
	s.SupportInterfaceSpacing = .2
	s.SupportSheath = false
	switch mode {
	case "easy_remove":
		s.SupportContactZ = math.Max(.18, .9*s.LayerHeight)
		s.SupportXYDistance = math.Max(.35, nozzle)
		s.SupportInterfaceSpacing = .3
	case "stable":
		s.SupportStyle = "grid"
		s.SupportContactZ = math.Max(.1, .5*s.LayerHeight)
		s.SupportXYDistance = math.Max(.2, .5*nozzle)
		s.SupportSpacing = 1.8
		s.SupportInterfaceLayers = 3
		s.SupportInterfaceSpacing = 0
		s.SupportSheath = true
	}
	if p.SupportStyle != "" && p.SupportStyle != "auto" {
		s.SupportStyle = p.SupportStyle
	}
	s.SupportBuildplateOnly = p.SupportPlacement == "buildplate"
	s.SupportThreshold = p.SupportAngle
	if s.SupportThreshold == 0 {
		s.SupportThreshold = 45
	}
	return s
}

// SupportVariants holds all printing settings fixed and bounds the real native
// comparison to two support shapes. Explicit shape choices are respected.
func SupportVariants(s Settings, p config.Preferences) []Settings {
	if !s.Supports || (SupportMode(p) != "optimized" && SupportMode(p) != "auto") || (p.SupportStyle != "" && p.SupportStyle != "auto") {
		return []Settings{s}
	}
	other := s
	other.SupportStyle = "snug"
	return []Settings{s, other}
}
