package forge

import (
	"archive/zip"
	"bytes"
	"embed"
	"encoding/json"
	"github.com/forge-slicer/forge/internal/config"
	"io/fs"
	"net/http"
	"strings"
)

// All application code, schemas, and documentation are bundled for offline use.
//
//go:embed frontend/dist docs schemas/*.json LICENSE THIRD_PARTY_NOTICES.md
var Files embed.FS

func Assets() fs.FS {
	f, e := fs.Sub(Files, "frontend/dist")
	if e != nil {
		panic(e)
	}
	return f
}
func WithDocuments(next http.Handler) http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /api/docs", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Query().Get("path")
		if path == "" {
			path = "configuration/getting-started.md"
		}
		if !fs.ValidPath(path) || !strings.HasSuffix(path, ".md") {
			http.Error(w, "invalid documentation path", http.StatusBadRequest)
			return
		}
		data, e := Files.ReadFile("docs/" + path)
		if e != nil {
			http.Error(w, "documentation page not found", http.StatusNotFound)
			return
		}
		content := strings.TrimPrefix(string(data), "\ufeff")
		title := strings.TrimSpace(strings.TrimPrefix(strings.SplitN(content, "\n", 2)[0], "#"))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		_ = json.NewEncoder(w).Encode(map[string]string{"title": title, "content": content, "path": path})
	})
	m.HandleFunc("GET /api/profile-kit", func(w http.ResponseWriter, r *http.Request) {
		var data bytes.Buffer
		z := zip.NewWriter(&data)
		write := func(name string, content []byte) error {
			f, e := z.Create(name)
			if e != nil {
				return e
			}
			_, e = f.Write(content)
			return e
		}
		e := fs.WalkDir(Files, "docs", func(path string, d fs.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if d.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}
			b, e := Files.ReadFile(path)
			if e != nil {
				return e
			}
			return write(path, b)
		})
		if e == nil {
			for _, name := range []string{"printer", "filament", "optimization"} {
				path := "schemas/" + name + ".schema.json"
				var b []byte
				b, e = Files.ReadFile(path)
				if e != nil {
					break
				}
				e = write(path, b)
				if e != nil {
					break
				}
			}
		}
		if e == nil {
			var b []byte
			b, e = json.MarshalIndent(config.PrinterTemplate(), "", "  ")
			if e == nil {
				e = write("printer-template.json", b)
			}
		}
		if e == nil {
			e = write("README.md", []byte("# Make Forge fit your printer\n\nStart with docs/configuration/getting-started.md. printer-template.json is an incomplete draft: replace every unknown value with verified hardware/firmware information, then validate and import it in Forge. No printer is preinstalled. The full offline field reference and external-assistant instructions are included. The application never contacts an external assistant.\n"))
		}
		closeErr := z.Close()
		if e == nil {
			e = closeErr
		}
		if e != nil {
			http.Error(w, "could not assemble profile documentation kit", 500)
			return
		}
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", `attachment; filename="forge-printer-profile-kit.zip"`)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		_, _ = w.Write(data.Bytes())
	})
	m.Handle("/docs/", http.FileServer(http.FS(Files)))
	m.Handle("/schemas/", http.FileServer(http.FS(Files)))
	m.Handle("/LICENSE", http.FileServer(http.FS(Files)))
	m.Handle("/THIRD_PARTY_NOTICES.md", http.FileServer(http.FS(Files)))
	m.Handle("/", next)
	return m
}
