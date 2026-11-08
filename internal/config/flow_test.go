package config

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDefaultFlowDoesNotRequireVerificationOrEvidence(t *testing.T) {
	printers, filaments := ReferenceFixture()
	p := printers[0]
	p.Extruders[0].FlowVerified = false
	p.Extruders[0].MaxVolumetricFlow = 4
	delete(p.Provenance, "maxVolumetricFlow")
	pref := Preferences{SchemaVersion: "1.0", Mode: "balanced", Budget: 64, MinWalls: 2, Supports: true}
	for _, metadata := range []string{"", `,"flowVerified":false`, `,"flowVerified":true`} {
		raw, _ := json.Marshal(p)
		// Go omits false legacy metadata. Exercise all historical JSON forms.
		raw = []byte(strings.Replace(string(raw), `"type":"direct"`, `"type":"direct"`+metadata, 1))
		parsed, err := ParsePrinter(raw)
		if err != nil {
			t.Fatalf("default-flow profile rejected (%s): %v", metadata, err)
		}
		if err = Compose(parsed, filaments[0], pref); err != nil {
			t.Fatal(err)
		}
	}
	for _, invalid := range []float64{0, -1, 101} {
		p.Extruders[0].MaxVolumetricFlow = invalid
		if ValidatePrinter(p) == nil {
			t.Fatalf("invalid numeric flow cap accepted: %v", invalid)
		}
	}
}
