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
	"github.com/hfiorillo/site/view/pages"
	"gopkg.in/yaml.v2"
)

type routeEntry struct {
	Name     string `yaml:"name"`
	Slug     string `yaml:"slug"`
	Location string `yaml:"location"`
	Date     string `yaml:"date"`
	GPXFile  string `yaml:"gpx"`
	Packlist string `yaml:"packlist"`
}

var (
	routesOnce   sync.Once
	routesList   []routeEntry
	routesCache  = map[string]*models.Route{}
	routesCoords = map[string]*gpx.RouteData{}
	routesErr    error
)

func loadRoutes() {
	raw, err := os.ReadFile("./content/routes/routes.yml")
	if err != nil {
		routesErr = fmt.Errorf("reading routes.yml: %w", err)
		return
	}
	if err := yaml.Unmarshal(raw, &routesList); err != nil {
		routesErr = fmt.Errorf("parsing routes.yml: %w", err)
		return
	}
	for i, r := range routesList {
		date, parseErr := time.Parse("2006-01-02", r.Date)
		if parseErr != nil {
			routesErr = fmt.Errorf("parsing date for %s: %w", r.Name, parseErr)
			return
		}
		rd, parseErr := gpx.Parse("." + r.GPXFile)
		if parseErr != nil {
			routesErr = fmt.Errorf("parsing gpx for %s: %w", r.Name, parseErr)
			return
		}
		cj, _ := gpx.CoordsToJSON(rd.Coords)
		routesCache[r.Slug] = &models.Route{
			Slug:          r.Slug,
			Name:          r.Name,
			Location:      r.Location,
			DistanceKm:    math.Round(rd.DistanceKm),
			ElevationGain: math.Round(rd.ElevationGain),
			ElevationMax:  math.Round(rd.ElevationMax),
			ElevationMin:  math.Round(rd.ElevationMin),
			Date:          date,
			CoordsJSON:    cj,
			GPXFile:       r.GPXFile,
			Packlist:      r.Packlist,
		}
		routesCoords[r.Slug] = rd
		routesList[i].Slug = r.Slug
	}
}

func (p PageHandler) HandleRoutes(w http.ResponseWriter, r *http.Request) error {
	routesOnce.Do(loadRoutes)
	siteOnce.Do(loadSiteMeta)
	if routesErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return pages.ErrorPage("Could not load routes.").Render(r.Context(), w)
	}

	var meta models.PageMeta
	meta = models.PageMeta{
		Title:       siteMeta.Routes.Title + " | " + siteMeta.Title,
		Description: siteMeta.Routes.Description,
		URL:         p.SiteURL + "/routes",
		Image:       p.SiteURL + siteImage(),
	}
	var list []*models.Route
	for _, entry := range routesList {
		if route, ok := routesCache[entry.Slug]; ok {
			list = append(list, route)
		}
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
	route, ok := routesCache[slug]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return pages.ErrorPage("Route not found.").Render(r.Context(), w)
	}

	siteOnce.Do(loadSiteMeta)
	meta := models.PageMeta{
		Title:       route.Name + " | " + siteMeta.Title,
		Description: route.Location,
		URL:         p.SiteURL + "/routes/" + slug,
		Image:       p.SiteURL + siteImage(),
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
	rd, ok := routesCoords[slug]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return pages.ErrorPage("Route not found.").Render(r.Context(), w)
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(rd.Coords)
}
