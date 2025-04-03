package ui

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/Siddharth9890/osquery-mvp/internal/database"
)

type Handler struct {
	dbService  *database.Service
	templates  *template.Template
	apiBaseURL string
}

type PageData struct {
	SystemInfo    SystemInfo
	InstalledApps []InstalledApp
	LastUpdated   string
	Error         string
}

type SystemInfo struct {
	OSVersion      string
	OSName         string
	OSPlatform     string
	OsqueryVersion string
}

type InstalledApp struct {
	Name    string
	Version string
}

func NewHandler(dbService *database.Service, apiBaseURL string) (*Handler, error) {
	tmpl, err := template.ParseGlob("ui/templates/*.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		dbService:  dbService,
		templates:  tmpl,
		apiBaseURL: apiBaseURL,
	}, nil
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(h.apiBaseURL + "/latest_data")
	if err != nil {
		log.Printf("Error fetching data from API: %v", err)
		renderErrorPage(h.templates, w, "Failed to fetch data from API")
		return
	}
	defer resp.Body.Close()

	var apiResp struct {
		Success bool `json:"success"`
		Data    struct {
			ID             int       `json:"id"`
			OSVersion      string    `json:"os_version"`
			OsqueryVersion string    `json:"osquery_version"`
			CollectedAt    time.Time `json:"collected_at"`
			Apps           []struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"installed_apps"`
		} `json:"data"`
		Error string `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		log.Printf("Error parsing API response: %v", err)
		renderErrorPage(h.templates, w, "Failed to parse API response")
		return
	}

	if !apiResp.Success {
		renderErrorPage(h.templates, w, apiResp.Error)
		return
	}

	osName, osPlatform, err := h.dbService.GetOSDetails()
	if err != nil {
		log.Printf("Error getting OS details: %v", err)
	}

	sysInfo := SystemInfo{
		OSVersion:      apiResp.Data.OSVersion,
		OSName:         osName,
		OSPlatform:     osPlatform,
		OsqueryVersion: apiResp.Data.OsqueryVersion,
	}

	apps := make([]InstalledApp, 0, len(apiResp.Data.Apps))
	for _, app := range apiResp.Data.Apps {
		apps = append(apps, InstalledApp{
			Name:    app.Name,
			Version: app.Version,
		})
	}

	lastUpdated := apiResp.Data.CollectedAt.Format("Jan 02, 2006 15:04:05")

	data := PageData{
		SystemInfo:    sysInfo,
		InstalledApps: apps,
		LastUpdated:   lastUpdated,
	}

	if err := h.templates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) Assets(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/assets/", http.FileServer(http.Dir("ui/assets"))).ServeHTTP(w, r)
}

func renderErrorPage(tmpl *template.Template, w http.ResponseWriter, errorMsg string) {
	data := PageData{
		Error: errorMsg,
	}
	if err := tmpl.ExecuteTemplate(w, "error.html", data); err != nil {
		log.Printf("Error rendering error template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
