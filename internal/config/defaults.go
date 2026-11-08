package config

import "github.com/forge-slicer/forge/examples/profiles"

// Defaults starts with no printer. Users must import their own verified hardware
// profile before slicing. Materials are editable generic starting points, not
// calibrations for a particular printer or spool; see the material catalogue docs.
func Defaults() ([]Printer, []Filament) {
	filaments := make([]Filament, 0, 7)
	for _, name := range []string{"generic-pla.json", "generic-petg.json", "generic-abs.json", "generic-asa.json", "generic-tpu.json", "generic-pa.json", "user-pla-plus-115.json"} {
		data, err := profiles.Files.ReadFile(name)
		if err != nil {
			panic(err)
		}
		f, err := ParseFilament(data)
		if err != nil {
			panic(err)
		}
		filaments = append(filaments, f)
	}
	return []Printer{}, filaments
}

// ReferenceFixture is only for deterministic development tests. It is never
// inserted into the application catalogue or offered as a configured printer.
func ReferenceFixture() ([]Printer, []Filament) {
	pb, err := profiles.Files.ReadFile("reference-cartesian.json")
	if err != nil {
		panic(err)
	}
	fb, err := profiles.Files.ReadFile("reference-pla.json")
	if err != nil {
		panic(err)
	}
	p, err := ParsePrinter(pb)
	if err != nil {
		panic(err)
	}
	f, err := ParseFilament(fb)
	if err != nil {
		panic(err)
	}
	return []Printer{p}, []Filament{f}
}
