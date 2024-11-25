package config

const AppConfigName = "app"

type Configuration struct {
	App AppConfig
}

type AppConfig struct {
	LogLevel     string `json:"log_level"`
	SiteID       string `json:"site_id"`
	SitesEnabled string `json:"sites_enabled"`
	// Config                         RefreshConfig   `json:"config"`
	PipelinePreOptMaxIterations int             `json:"pipeline_pre_opt_max_iterations"`
	FeatureFlags                map[string]bool `json:"feature_flags"`
	// MySQLRCBuffer                  DBConnection    `json:"mysql_rc_buffer"`
	DistributionOrderByFC          string          `json:"distribution_order_by_fc"`
	RehydrationFeatureFlagsGrouped map[string]bool `json:"rehydration_feature_flags_grouped"`
	UseCase                        struct {
		Selected string `json:"selected"`
	} `json:"use_case"`
	NodesRC   []string `json:"nodes_rc"`
	NodesRCFC []string `json:"nodes_rc_fc"`
}
