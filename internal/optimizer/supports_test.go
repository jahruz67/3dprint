package optimizer

import (
	"github.com/forge-slicer/forge/internal/config"
	"testing"
)

func TestSupportPolicyAndBoundedVariants(t *testing.T) {
	s := Settings{LayerHeight: .2, Supports: true}
	p := config.Preferences{Supports: true, SupportMode: "auto"}
	normal := SupportPolicy(s, p, .4)
	if len(SupportVariants(normal, p)) != 2 {
		t.Fatal("auto must compare two native support shapes")
	}
	p.SupportMode = "easy_remove"
	easy := SupportPolicy(s, p, .4)
	p.SupportMode = "stable"
	stable := SupportPolicy(s, p, .4)
	if easy.SupportContactZ <= stable.SupportContactZ || easy.SupportXYDistance <= stable.SupportXYDistance || stable.SupportSpacing >= easy.SupportSpacing || !stable.SupportSheath {
		t.Fatal("recipes do not express removal/stability tradeoffs")
	}
	p.SupportStyle = "snug"
	p.SupportMode = "optimized"
	explicit := SupportPolicy(s, p, .4)
	if len(SupportVariants(explicit, p)) != 1 || explicit.SupportStyle != "snug" {
		t.Fatal("explicit shape ignored")
	}
	p.SupportMode = "off"
	if SupportPolicy(s, p, .4).Supports {
		t.Fatal("off enables support")
	}
}
