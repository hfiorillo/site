package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hfiorillo/site/models"
	"github.com/hfiorillo/site/paths"
	"github.com/hfiorillo/site/view/pages"
	"gopkg.in/yaml.v2"
)

type routeEntry struct {
	RouteURL      string  `yaml:"route_url"`
	Name          string  `yaml:"name"`
	Slug          string  `yaml:"slug"`
	Location      string  `yaml:"location"`
	Date          string  `yaml:"date"`
	Packlist      string  `yaml:"packlist"`
	DistanceKm    float64 `yaml:"distance_km"`
	ElevationGain float64 `yaml:"elevation_gain"`
}

var (
	routesOnce  sync.Once
	routesList  []routeEntry
	routesCache = map[string]*models.Route{}
	routesErr   error
)

// Load once and never mutate the routes after publishing the cache.
func loadRoutes() {
	raw, err := os.ReadFile(paths.RoutesYAML)
	if err != nil {
		routesErr = err
		return
	}
	if err := yaml.Unmarshal(raw, &routesList); err != nil {
		routesErr = err
		return
	}
	for _, r := range routesList {
		target, err := url.Parse(r.RouteURL)
		if err != nil || target.Scheme != "https" || (target.Host != "www.komoot.com" && target.Host != "komoot.com") {
			routesErr = fmt.Errorf("invalid Komoot URL for %s", r.Name)
			return
		}
		var date time.Time
		if r.Date != "" {
			date, err = time.Parse("2006-01-02", r.Date)
			if err != nil {
				routesErr = err
				return
			}
		}
		routesCache[r.Slug] = &models.Route{RouteURL: r.RouteURL, Slug: r.Slug, Name: r.Name, Location: r.Location, Date: date, Packlist: r.Packlist, DistanceKm: r.DistanceKm, ElevationGain: r.ElevationGain}
	}
}

func (p PageHandler) HandleRoutes(w http.ResponseWriter, r *http.Request) error {
	routesOnce.Do(loadRoutes)
	if routesErr != nil {
		return routesErr
	}
	siteOnce.Do(loadSiteMeta)
	list := make([]*models.Route, 0, len(routesList))
	for _, entry := range routesList {
		list = append(list, routesCache[entry.Slug])
	}
	meta := models.PageMeta{Title: siteMeta.Routes.Title + " | " + siteMeta.Title, Description: siteMeta.Routes.Description, URL: p.SiteURL + paths.Routes, Canonical: p.SiteURL + paths.Routes, Image: p.SiteURL + siteImage(), StructuredData: personJSON(p.SiteURL)}
	return pages.Routes(list, meta).Render(r.Context(), w)
}

// Preserve existing bookmarks while the listing links directly to Komoot.
func (p PageHandler) HandleRoute(w http.ResponseWriter, r *http.Request) error {
	routesOnce.Do(loadRoutes)
	if routesErr != nil {
		return routesErr
	}
	route := routesCache[chi.URLParam(r, "slug")]
	if route == nil {
		http.NotFound(w, r)
		return nil
	}
	http.Redirect(w, r, route.RouteURL, http.StatusTemporaryRedirect)
	return nil
}
