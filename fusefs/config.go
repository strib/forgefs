package fusefs

// Config represents the config file for a FUSE-based file system.
type Config struct {
	Debug         bool   `json:"debug,omitempty"`
	DoKAPIKey     string `json:"dok_api_key,omitempty"`
	DoKAddr       string `json:"dok_addr,omitempty"`
	SkyJAddr      string `json:"skyj_addr,omitempty"`
	DBFile        string `json:"db_file,omitempty"`
	Mountpoint    string `json:"mountpoint,omitempty"`
	ImageCacheDir string `json:"image_cache_dir,omitempty"`
}
