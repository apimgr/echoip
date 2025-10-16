package server

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

var templates *template.Template

// InitTemplates initializes the HTML templates
func InitTemplates() error {
	var err error
	templates, err = template.ParseFS(templateFS, "templates/*.html")
	return err
}

// StaticHandler serves embedded static files
func StaticHandler() http.Handler {
	// Strip the "static" prefix from the embedded FS
	staticContent, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(staticContent))
}
