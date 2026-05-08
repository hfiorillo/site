package gpx

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"os"
)

type gpxFile struct {
	Tracks []track `xml:"trk"`
}

type track struct {
	Segments []segment `xml:"trkseg"`
}

type segment struct {
	Points []point `xml:"trkpt"`
}

type point struct {
	Lat float64 `xml:"lat,attr"`
	Lon float64 `xml:"lon,attr"`
	Ele float64 `xml:"ele"`
}

type Coord struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type RouteData struct {
	Coords        []Coord
	DistanceKm    float64
	ElevationGain float64
	ElevationMax  float64
	ElevationMin  float64
}

func Parse(filePath string) (*RouteData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading gpx: %w", err)
	}

	var gpx gpxFile
	if err := xml.Unmarshal(data, &gpx); err != nil {
		return nil, fmt.Errorf("parsing gpx xml: %w", err)
	}

	var rd RouteData
	rd.ElevationMin = math.MaxFloat64
	prevLat, prevLon := math.Inf(1), math.Inf(1)
	prevEle := math.Inf(1)

	for _, trk := range gpx.Tracks {
		for _, seg := range trk.Segments {
			for _, pt := range seg.Points {
				rd.Coords = append(rd.Coords, Coord{Lat: pt.Lat, Lon: pt.Lon})

				if !math.IsInf(prevLat, 1) {
					rd.DistanceKm += haversine(prevLat, prevLon, pt.Lat, pt.Lon)
				}
				prevLat, prevLon = pt.Lat, pt.Lon

				if !math.IsInf(prevEle, 1) && pt.Ele > prevEle {
					rd.ElevationGain += pt.Ele - prevEle
				}
				prevEle = pt.Ele

				if pt.Ele > rd.ElevationMax {
					rd.ElevationMax = pt.Ele
				}
				if pt.Ele < rd.ElevationMin {
					rd.ElevationMin = pt.Ele
				}
			}
		}
	}

	if rd.ElevationMin == math.MaxFloat64 {
		rd.ElevationMin = 0
	}

	return &rd, nil
}

func CoordsToJSON(coords []Coord) (string, error) {
	b, err := json.Marshal(coords)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * 6371 * math.Asin(math.Sqrt(a))
}
