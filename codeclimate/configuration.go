package codeclimate

// Model for the CodeClimate configuration that is not TFLint specific
type CodeClimateConfiguration struct {
	ExcludePatterns []string `json:"exclude_patterns,omitempty"`
	IncludePaths    []string `json:"include_paths,omitempty"`
}
