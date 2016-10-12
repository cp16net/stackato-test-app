package common

import (
	"github.com/hpcloud/hsm/generated/hsm/models"
)

// InternalServiceConfig is the set of configuration values that can be on a service
type InternalServiceConfig struct {
	Description           string   `json:"description"`
	Categories            []string `json:"categories"`
	Labels                []string `json:"labels"`
	Type                  string   `json:"type"`
	DefaultServiceVersion string   `json:"default_service_version"`
	DocumentationURL      string   `json:"documentation_url"`
	LicenseURL            string   `json:"license_url"`
	Private               bool     `json:"private"`
	IconURL               string   `json:"icon_url"`
}

// InternalConfigTemplate Config Template structure
type InternalConfigTemplate struct {
	Name           string                        `json:"name"`
	ProductVersion string                        `json:"product_version"`
	SDLVersion     string                        `json:"sdl_version"`
	Vendor         string                        `json:"vendor"`
	Parameters     []*models.ParameterDefinition `json:"parameters"`
}

// InternalService structure
type InternalService struct {
	ServiceItemResponse models.ServiceItemResponse
	IsIconURLFromConfig bool
	Path                string
	Versions            map[string]SDLConfig
}

//SDLConfig structure to store details from sdl/config.json
type SDLConfig struct {
	Disabled     *VersionStatusConfig `json:"disabled"`
	UpgradeFrom  []upgradeFrom        `json:"upgrade_from"`
	Dependencies []dependencyOn       `json:"dependencies"`
}

type upgradeFrom struct {
	ProductVersion string   `json:"product_version"`
	SdlVersions    []string `json:"sdl_versions"`
}

type dependencyOn struct {
	Name           string   `json:"name"`
	Vendor         string   `json:"vendor"`
	ProductVersion string   `json:"product_version"`
	SdlVersions    []string `json:"sdl_versions"`
}

//VersionStatusConfig internal sdl status config
type VersionStatusConfig struct {
	Vulnerability bool   `json:"vulnerability"`
	Link          string `json:"link"`
	Reason        string `json:"reason"`
}
