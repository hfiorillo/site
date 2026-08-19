package handler

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hfiorillo/site/internal/gpx"
	"github.com/hfiorillo/site/models"
	"github.com/hfiorillo/site/paths"
	"github.com/hfiorillo/site/view/pages"
	"gopkg.in/yaml.v2"
)

type routeEntry struct {
	Name         string  `yaml:"name"`
	Slug         string  `yaml:"slug"`
	Location     string  `yaml:"location"`
	Date         string  `yaml:"date"`
	GPXFile      string  `yaml:"gpx"`
	Packlist     string  `yaml:"packlist"`
	DistanceKm   float64 `yaml:"distance_km"`
	ElevationGain float64 `yaml:"elevation_gain"`
}

var (
	routesOnce   sync.Once
	routesList   []routeEntry
	routesCache  = map[string]*models.Route{}
	routesCoords = map[string]*gpx.RouteData{}
	routesErr    error
	routesMu     sync.RWMutex
)

func loadRoutes() {
	raw, err := os.ReadFile(paths.RoutesYAML)
	if err != nil {
		routesErr = fmt.Errorf("reading routes.yml: %w", err)
		return
	}
	if err := yaml.Unmarshal(raw, &routesList); err != nil {
		routesErr = fmt.Errorf("parsing routes.yml: %w", err)
		return
	}
	for i := range routesList {
		r := &routesList[i]
		date, parseErr := time.Parse("2006-01-02", r.Date)
		if parseErr != nil {
			routesErr = fmt.Errorf("parsing date for %s: %w", r.Name, parseErr)
			return
		}
		routesCache[r.Slug] = &models.Route{
			Slug:          r.Slug,
			Name:          r.Name,
			Location:      r.Location,
			Date:          date,
			GPXFile:       r.GPXFile,
			Packlist:      r.Packlist,
			DistanceKm:    r.DistanceKm,
			ElevationGain: r.ElevationGain,
		}
	}
}

func ensureRouteData(slug string) (*models.Route, *gpx.RouteData, error) {
	routesMu.RLock()
	route := routesCache[slug]
	rd := routesCoords[slug]
	routesMu.RUnlock()
	if route == nil {
		return nil, nil, fmt.Errorf("route not found: %s", slug)
	}
	if rd != nil {
		return route, rd, nil
	}

	rd, err := gpx.Parse("." + route.GPXFile)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing gpx for %s: %w", slug, err)
	}
	cj, _ := gpx.CoordsToJSON(rd.Coords)

	routesMu.Lock()
	if route.DistanceKm == 0 {
		route.DistanceKm = math.Round(rd.DistanceKm)
	}
	if route.ElevationGain == 0 {
		route.ElevationGain = math.Round(rd.ElevationGain)
	}
	route.ElevationMax = math.Round(rd.ElevationMax)
	route.ElevationMin = math.Round(rd.ElevationMin)
	route.CoordsJSON = cj
	routesCoords[slug] = rd
	routesMu.Unlock()

	return route, rd, nil
}

func (p PageHandler) HandleRoutes(w http.ResponseWriter, r *http.Request) error {
	routesOnce.Do(loadRoutes)
	siteOnce.Do(loadSiteMeta)
	if routesErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return pages.ErrorPage("Could not load routes.").Render(r.Context(), w)
	}

	var list []*models.Route
	for _, entry := range routesList {
		routesMu.RLock()
		route := routesCache[entry.Slug]
		routesMu.RUnlock()
		if route != nil {
			list = append(list, route)
		}
	}

	meta := models.PageMeta{
		Title:          siteMeta.Routes.Title + " | " + siteMeta.Title,
		Description:    siteMeta.Routes.Description,
		URL:            p.SiteURL + paths.Routes,
		Canonical:      p.SiteURL + paths.Routes,
		Image:          p.SiteURL + siteImage(),
		StructuredData: personJSON(p.SiteURL),
	}
	return pages.Routes(list, meta).Render(r.Context(), w)
}

func (p PageHandler) HandleRoute(w http.ResponseWriter, r *http.Request) error {
	routesOnce.Do(loadRoutes)
	if routesErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return pages.ErrorPage("Could not load route.").Render(r.Context(), w)
	}

	slug := chi.URLParam(r, "slug")
	route, rd, err := ensureRouteData(slug)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return pages.ErrorPage("Route not found.").Render(r.Context(), w)
	}
	_ = rd

	siteOnce.Do(loadSiteMeta)
	meta := models.PageMeta{
		Title:          route.Name + " | " + siteMeta.Title,
		Description:    route.Location,
		URL:            p.SiteURL + paths.Routes + "/" + slug,
		Canonical:      p.SiteURL + paths.Routes + "/" + slug,
		Image:          p.SiteURL + siteImage(),
		StructuredData: personJSON(p.SiteURL),
	}
	return pages.RoutePage(route, slug, meta).Render(r.Context(), w)
}

func (p PageHandler) HandleRouteCoords(w http.ResponseWriter, r *http.Request) error {
	routesOnce.Do(loadRoutes)
	if routesErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return pages.ErrorPage("Could not load route.").Render(r.Context(), w)
	}

	slug := chi.URLParam(r, "slug")
	_, rd, err := ensureRouteData(slug)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return pages.ErrorPage("Route not found.").Render(r.Context(), w)
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(rd.Coords)
}
