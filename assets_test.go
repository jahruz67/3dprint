package forge

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOfflinePrinterGuideAndKit(t *testing.T) {
	h := WithDocuments(http.NotFoundHandler())
	request := func(path string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	w := request("/api/docs?path=configuration/getting-started.md")
	if w.Code != 200 {
		t.Fatal("first-run guide is not bundled")
	}
	var doc struct{ Title, Content, Path string }
	if e := json.Unmarshal(w.Body.Bytes(), &doc); e != nil || doc.Title == "" || len(doc.Content) < 1000 {
		t.Fatal("guide lacks actual detailed instructions")
	}
	for _, path := range []string{"../LICENSE", "/LICENSE", "configuration/../../LICENSE"} {
		if w = request("/api/docs?path=" + path); w.Code != 400 {
			t.Fatal("documentation path escaped bundle")
		}
	}
	w = request("/api/profile-kit")
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	z, e := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
	if e != nil {
		t.Fatal(e)
	}
	names := map[string]bool{}
	for _, f := range z.File {
		names[f.Name] = true
	}
	for _, name := range []string{"README.md", "printer-template.json", "schemas/printer.schema.json", "schemas/filament.schema.json", "docs/configuration/getting-started.md", "docs/configuration/external-ai-guide.md", "docs/configuration/printer-schema.md"} {
		if !names[name] {
			t.Fatalf("profile creation kit missing %s", name)
		}
	}
}
