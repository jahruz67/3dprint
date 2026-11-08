package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	forge "github.com/forge-slicer/forge"
	"github.com/forge-slicer/forge/internal/service"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func main() {
	serve := flag.Bool("serve", false, "serve the offline UI on loopback instead of opening a native window")
	address := flag.String("listen", "127.0.0.1:8787", "loopback address for --serve")
	profile := flag.String("validate-profile", "", "validate a profile file and exit")
	kind := flag.String("kind", "printer", "profile kind: printer, filament, optimization")
	data := flag.String("data-dir", "", "override local profile storage directory")
	smoke := flag.Bool("smoke-test", false, "launch the native WebView, verify its API connection, then exit")
	flag.Parse()
	if *profile != "" {
		b, e := os.ReadFile(*profile)
		if e == nil {
			e = service.ValidateProfile(*kind, b)
		}
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		fmt.Println("Valid " + *kind + " profile")
		return
	}
	s, e := service.New(*data)
	if e != nil {
		log.Fatal(e)
	}
	defer s.Close()
	handler := forge.WithDocuments(s.Handler(forge.Assets()))
	if *serve {
		host, _, e := net.SplitHostPort(*address)
		if e != nil || (host != "127.0.0.1" && host != "localhost") {
			log.Fatal("--listen must use 127.0.0.1 or localhost")
		}
		listener, e := net.Listen("tcp", *address)
		if e != nil {
			log.Fatal(e)
		}
		server := &http.Server{Handler: handler, ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		go func() {
			<-ctx.Done()
			shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = server.Shutdown(shutdown)
		}()
		slog.Info("Forge Slicer offline server", "url", "http://"+listener.Addr().String())
		if e = server.Serve(listener); e != nil && e != http.ErrServerClosed {
			log.Print(e)
		}
		return
	}
	frontendReady := make(chan struct{}, 1)
	var nativeHandler http.Handler = handler
	if *smoke {
		nativeHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w, r)
			if r.URL.Path == "/api/state" {
				select {
				case frontendReady <- struct{}{}:
				default:
				}
			}
		})
	}
	app := application.New(application.Options{Name: "Forge Slicer", Description: "Local optimization-first FDM slicing", Assets: application.AssetOptions{Handler: nativeHandler}, Mac: application.MacOptions{ApplicationShouldTerminateAfterLastWindowClosed: true}})
	window := app.Window.NewWithOptions(application.WebviewWindowOptions{Title: "Forge Slicer", Width: 1440, Height: 940, MinWidth: 1050, MinHeight: 700, URL: "/", Hidden: *smoke})
	s.SetSaveDialog(func(name string) (string, error) {
		return app.Dialog.SaveFile().SetFilename(name).AddFilter("G-code", "*.gcode").AttachToWindow(window).PromptForSingleSelection()
	})
	smokeResult := make(chan bool, 1)
	if *smoke {
		go func() {
			select {
			case <-frontendReady:
				smokeResult <- true
			case <-time.After(20 * time.Second):
				smokeResult <- false
			}
			app.Quit()
		}()
	}
	if e = app.Run(); e != nil {
		log.Print(e)
	}
	if *smoke {
		if ok := <-smokeResult; !ok || e != nil {
			log.Print("Native desktop smoke check failed")
			s.Close()
			os.Exit(1)
		}
		log.Print("Native desktop smoke check passed: embedded UI connected to Go services")
	}
}
