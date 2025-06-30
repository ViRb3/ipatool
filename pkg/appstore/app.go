package appstore

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog"
)

type App struct {
	ID       int64   `json:"trackId,omitempty"`
	BundleID string  `json:"bundleId,omitempty"`
	Name     string  `json:"trackName,omitempty"`
	Version  string  `json:"version,omitempty"`
	Price    float64 `json:"price,omitempty"`
}

type VersionHistoryInfo struct {
	App                App
	LatestVersion      string
	VersionIdentifiers []string
}

type VersionDetails struct {
	VersionID     string
	VersionString string
	Success       bool
	Error         string
}

type Apps []App

func (apps Apps) MarshalZerologArray(a *zerolog.Array) {
	for _, app := range apps {
		a.Object(app)
	}
}

func (a App) MarshalZerologObject(event *zerolog.Event) {
	event.
		Int64("id", a.ID).
		Str("bundleID", a.BundleID).
		Str("name", a.Name).
		Str("version", a.Version).
		Float64("price", a.Price)
}

func (a *App) LoadMetadata(metadata map[string]interface{}) {
	for prop, source := range map[*string]string{
		// there is also bundleDisplayName, but this matches lookup name
		&a.Name:     "itemName",
		&a.BundleID: "softwareVersionBundleId",
		// this is the API version (externalVersionID), which is always present even for unlisted apps
		&a.Version: "bundleVersion",
	} {
		if val, ok := metadata[source]; ok {
			switch val := (val).(type) {
			case string:
				*prop = val
			}
		}
	}
	// if present, this is the public-facing version (displayVersion), so use it where possible
	if version, ok := metadata["bundleShortVersionString"]; ok {
		switch version := (version).(type) {
		case string:
			a.Version = version
		}
	}
}

func (a App) GetIPAName() string {
	return fmt.Sprintf("%s+%s+%d+%s.ipa",
		a.cleanName(a.BundleID),
		a.cleanName(a.Name),
		a.ID,
		a.cleanName(a.Version))
}

var cleanRegex1 = regexp.MustCompile("[^-\\w.]")
var cleanRegex2 = regexp.MustCompile("\\s+")

func (a App) cleanName(name string) string {
	name = cleanRegex1.ReplaceAllString(name, " ")
	name = strings.TrimSpace(name)
	name = cleanRegex2.ReplaceAllString(name, "_")
	return name
}
