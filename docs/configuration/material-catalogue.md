# Material starting points and sources

Forge includes six selectable generic material recipes: PLA, PETG, ABS, ASA, TPU 95A and PA12. They are editable starting points, not calibrated settings for every brand, color, spool, nozzle or machine. None supplies a printer profile or raises a printer's verified limits. Review the spool manufacturer's instructions, duplicate the closest material with a new identifier, and adjust it for the actual combination before printing.

It also includes the separately labeled **Custom PLA+ 115 C nozzle - unverified** entry, ID `user-pla-plus-115`. The 115 C **nozzle** setpoint is the user's explicit custom request, not a manufacturer recommendation and not a bed-temperature interpretation. Its material notes are displayed for review. eSUN's [official PLA+ documentation](https://www.esun3d.com/pla-pro-product/) instead lists nozzle 210–230 C, bed 45–60 C, density 1.24 g/cm3 and full cooling. Those manufacturer data explain the contrast and supply illustrative density/bed/cooling inputs; they provide **no evidence of printable extrusion at 115 C**. The custom 2 mm3/s cap is likewise uncalibrated. The printer's verified minimum temperature and firmware cold-extrusion protection remain unchanged, so composition can correctly reject this material for a machine that cannot print at the requested temperature. Never lower a hardware limit merely to silence that error.

## Source record

Primary manufacturer documentation was reviewed on 2026-11-01. The table preserves the relevant source facts locally so the catalogue remains understandable offline; opening source links is optional. Setpoints are engineering selections inside the cited product ranges, rather than claims of a universally optimum temperature. Density is an illustrative mass-estimation input, not a certified property of an unknown spool.

| Starter ID | Nozzle / bed chosen, C | Density, g/cm3 | Source facts and primary reference |
|---|---|---|---|
| `generic-pla` | 210 / 60 | 1.24 | Fused Materials publishes PLA nozzle 190–230 C, bed 60–80 C and density 1.20–1.30. The chosen density lies inside that range. [Manufacturer fact sheet](https://fusedmaterials.com/pages/filament-fact-sheet/). |
| `generic-petg` | 230 / 80 | 1.27 | Fused Materials publishes PETG nozzle 200–240 C, bed 75–85 C and density 1.27. [Manufacturer fact sheet](https://fusedmaterials.com/pages/filament-fact-sheet/). |
| `generic-abs` | 245 / 90 | 1.05 | Fused Materials publishes ABS nozzle 230–260 C, bed 80–100 C and density 1.05. [Manufacturer fact sheet](https://fusedmaterials.com/pages/filament-fact-sheet/). |
| `generic-asa` | 250 / 85 | 1.13 | Polymaker's PolyLite ASA specifies nozzle 240–260 C, bed 75–95 C and fan off; its data sheet gives density 1.13. [Manufacturer printing conditions](https://polymaker.com/product/polylite-asa), [manufacturer TDS](https://polymaker.com/wp-content/uploads/lana-downloads/PolyLite_ASA_TDS_EN_V5.4.pdf). |
| `generic-tpu` | 220 / 50 | 1.22 | Polymaker's PolyFlex TPU95 specifies nozzle 210–230 C, bed 25–60 C, fan on and density 1.20–1.24; chosen density is the midpoint. [Manufacturer technical data](https://wiki.polymaker.com/polymaker-products/more-about-our-products/documents/technical-data-sheets/tpu/polyflex-tm-tpu95). |
| `generic-pa` | 240 / 100 | 1.04 | Fused Materials publishes PA12 nozzle 230–250 C, bed 90–110 C and density 1.04. This starter is for PA12; do not treat its temperatures as a PA6 or composite recipe. [Manufacturer fact sheet](https://fusedmaterials.com/pages/filament-fact-sheet/). |

The product-derived numbers do not imply an endorsement, product identification or compatibility claim for an arbitrary generic spool. Manufacturers can change formulations. Where exact product documentation differs, use the actual spool's documentation and retain that evidence with the exported profile.

## Explicit engineering assumptions

The starter throughput caps are **uncalibrated engineering choices**, not sourced hot-end ratings: PLA 6, PETG 5, ABS 6, ASA 6, TPU 2 and PA12 4 mm3/s. They constrain estimated extrusion demand but are not guaranteed printable on unknown hardware. The effective limit is the smaller of the material cap and the imported printer's configured flow cap. You may tune these caps for the actual material/nozzle/temperature; calibration is optional. Never copy a manufacturer's melt-flow index in g/10min into a mm3/s hot-end flow field: these are different measurements.

The shared starter layer-height interval, 0.08–0.30 mm, is an engineering envelope commonly usable with a 0.4 mm nozzle; the actual optimizer intersects it with the imported nozzle's 0.2–0.8 diameter interval. It is not sourced as a property of every material. Review it for a changed nozzle or spool.

Cooling choices are PLA/TPU 100%, PETG 20%, ABS/ASA/PA12 0%. ASA off and TPU on follow the cited product recommendations; the remaining choices are initial engineering assumptions to review. Polymaker separately recommends low cooling up to 20% for [PolyLite PETG](https://polymaker.com/product/polylite-petg), supporting the low-cooling starting approach without implying that all PETG brands are identical. Positive cooling requires a verified machine fan.

Material definitions do not currently model enclosure temperature, drying state, fumes, chamber control, brand-specific adhesion treatments or material-specific retraction. Those remain human setup responsibilities; required hardware and conditions must be verified independently. A successful schema check establishes data validity, not a complete process qualification. The catalogue is intentionally easy to export, duplicate and replace with spool-specific profiles.
