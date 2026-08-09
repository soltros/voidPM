package xbps

type Package struct {
	Name           string   `json:"name"`
	Version        string   `json:"version"`
	ShortDesc      string   `json:"short_desc"`
	Installed      bool     `json:"installed"`
	Orphan         bool     `json:"orphan"`
	OnHold         bool     `json:"on_hold"`
	InstalledSize  string   `json:"installed_size,omitempty"`
	DownloadSize   string   `json:"download_size,omitempty"`
	Repository     string   `json:"repository,omitempty"`
	Maintainer     string   `json:"maintainer,omitempty"`
	Homepage       string   `json:"homepage,omitempty"`
	License        string   `json:"license,omitempty"`
	Architecture   string   `json:"architecture,omitempty"`
	Dependencies   []string `json:"dependencies,omitempty"`
	ReverseDeps    []string `json:"reverse_deps,omitempty"`
}
