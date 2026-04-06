// Package helm provides parsers for helm command output.
// This package is in the adapters layer and implements parsers for
// converting raw helm command output into structured domain types.
package helm

// Release represents a Helm release from helm list output.
type Release struct {
	// Name is the release name.
	Name string `json:"name"`

	// Namespace is the Kubernetes namespace where the release is installed.
	Namespace string `json:"namespace"`

	// Revision is the release revision number.
	Revision int `json:"revision"`

	// Updated is the timestamp of the last update (e.g., "2024-01-15 10:30:45.123456789 +0000 UTC").
	Updated string `json:"updated"`

	// Status is the release status (e.g., "deployed", "failed", "pending-install").
	Status string `json:"status"`

	// Chart is the chart name and version (e.g., "nginx-1.0.0").
	Chart string `json:"chart"`

	// AppVersion is the application version from the chart.
	AppVersion string `json:"app_version"`
}

// ListResult represents the structured output of 'helm list'.
type ListResult struct {
	// Releases is the list of Helm releases.
	Releases []Release `json:"releases"`
}

// Revision represents a release revision from helm history output.
type Revision struct {
	// Revision is the revision number.
	Revision int `json:"revision"`

	// Updated is the timestamp of this revision.
	Updated string `json:"updated"`

	// Status is the status at this revision (e.g., "deployed", "superseded").
	Status string `json:"status"`

	// Chart is the chart name and version.
	Chart string `json:"chart"`

	// AppVersion is the application version from the chart.
	AppVersion string `json:"app_version"`

	// Description is the revision description (e.g., "Install complete", "Upgrade complete").
	Description string `json:"description"`
}

// HistoryResult represents the structured output of 'helm history'.
type HistoryResult struct {
	// Revisions is the list of revisions for the release.
	Revisions []Revision `json:"revisions"`
}

// ChartInfo represents chart information from helm search or helm show.
type ChartInfo struct {
	// Name is the chart name.
	Name string `json:"name"`

	// ChartVersion is the version of the chart.
	ChartVersion string `json:"chart_version"`

	// AppVersion is the version of the application in the chart.
	AppVersion string `json:"app_version"`

	// Description is the chart description.
	Description string `json:"description"`
}

// SearchResult represents the structured output of 'helm search repo'.
type SearchResult struct {
	// Charts is the list of charts matching the search.
	Charts []ChartInfo `json:"charts"`
}

// ReleaseResource represents a Kubernetes resource that is part of a release.
type ReleaseResource struct {
	// Kind is the Kubernetes resource kind (e.g., "Deployment", "Service").
	Kind string `json:"kind"`

	// Name is the resource name.
	Name string `json:"name"`

	// Namespace is the resource namespace.
	Namespace string `json:"namespace,omitempty"`
}

// StatusResult represents the structured output of 'helm status'.
type StatusResult struct {
	// Name is the release name.
	Name string `json:"name"`

	// Namespace is the release namespace.
	Namespace string `json:"namespace"`

	// Revision is the current revision number.
	Revision int `json:"revision"`

	// Status is the release status.
	Status string `json:"status"`

	// LastDeployed is the timestamp of the last deployment.
	LastDeployed string `json:"last_deployed"`

	// Description is the status description.
	Description string `json:"description,omitempty"`

	// Notes is the NOTES.txt output from the chart.
	Notes string `json:"notes,omitempty"`

	// Resources is the list of Kubernetes resources in the release.
	Resources []ReleaseResource `json:"resources,omitempty"`

	// currentResourceKind is used during parsing to track the current resource kind.
	currentResourceKind string `json:"-"`
}

// ChartValue represents a single chart value from helm show values.
type ChartValue struct {
	// Key is the value path (e.g., "image.repository").
	Key string `json:"key"`

	// Value is the default value.
	Value any `json:"value"`

	// Description is an optional description from comments.
	Description string `json:"description,omitempty"`
}

// ShowValuesResult represents the structured output of 'helm show values'.
type ShowValuesResult struct {
	// Values is the list of chart values.
	Values []ChartValue `json:"values,omitempty"`

	// Raw is the raw YAML output for complex nested values.
	Raw string `json:"raw"`
}
