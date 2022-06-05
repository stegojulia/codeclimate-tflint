package tflint

import (
	"fmt"
	"log"
)

type TFLintRoot struct {
	Config TFLintConfiguration `json:"config,omitempty"`
}

type TFLintConfiguration struct {
	Config       string                        `json:"config,omitempty"`
	IgnoreModule []string                      `json:"ignore_module,omitempty"`
	EnableRule   []string                      `json:"enable_rule,omitempty"`
	DisableRule  []string                      `json:"disable_rule,omitempty"`
	Only         []string                      `json:"only,omitempty"`
	EnablePlugin []string                      `json:"enable_plugin,omitempty"`
	VarFile      string                        `json:"var_file,omitempty"`
	Var          []TFLintConfigurationVariable `json:"var"`
	Module       bool                          `json:"module,omitempty"`
}

type TFLintConfigurationVariable struct {
	Key   string
	Value string
}

// Generates na array of CLI arguments from the provided configuration
// We simply generate them without doing any kind of logic as args handling is demanded to TFLint itself
func ToCLIArguments(configuration TFLintConfiguration) []string {
	log.Println("[tflint/configuration.go/ToCLIArguments] Converting configuration file to CLI args...")
	args := []string{}

	if configuration.Config != "" {
		log.Printf("[tflint/configuration.go/ToCLIArguments] Setting --config=%v\n", configuration.Config)
		args = append(args, fmt.Sprintf("--config=%v", configuration.Config))
	}

	if configuration.IgnoreModule != nil {
		for _, item := range configuration.IgnoreModule {
			log.Printf("[tflint/configuration.go/ToCLIArguments] Setting --ignore-module=%v\n", item)
			args = append(args, fmt.Sprintf("--ignore-module=%v", item))
		}
	}

	if configuration.EnableRule != nil {
		for _, item := range configuration.EnableRule {
			log.Printf("[tflint/configuration.go/ToCLIArguments] Setting --enable-rule=%v\n", item)
			args = append(args, fmt.Sprintf("--enable-rule=%v", item))
		}
	}

	if configuration.DisableRule != nil {
		for _, item := range configuration.DisableRule {
			log.Printf("[tflint/configuration.go/ToCLIArguments] Setting --disable-rule=%v\n", item)
			args = append(args, fmt.Sprintf("--disable-rule=%v", item))
		}
	}

	if configuration.Only != nil {
		for _, item := range configuration.Only {
			log.Printf("[tflint/configuration.go/ToCLIArguments] Setting --only=%v\n", item)
			args = append(args, fmt.Sprintf("--only=%v", item))
		}
	}

	if configuration.EnablePlugin != nil {
		for _, item := range configuration.EnablePlugin {
			log.Printf("[tflint/configuration.go/ToCLIArguments] Setting --enable-plugin=%v\n", item)
			args = append(args, fmt.Sprintf("--enable-plugin=%v", item))
		}
	}

	if configuration.VarFile != "" {
		log.Printf("[tflint/configuration.go/ToCLIArguments] Setting --var-file=%v\n", configuration.VarFile)
		args = append(args, fmt.Sprintf("--var-file=%v", configuration.VarFile))
	}

	if configuration.Var != nil {
		for _, item := range configuration.Var {
			log.Printf("[tflint/configuration.go/ToCLIArguments] Setting --var='%v=%v'\n", item.Key, item.Value)
			args = append(args, fmt.Sprintf("--var='%v=%v'", item.Key, item.Value))
		}
	}

	if configuration.Module {
		log.Printf("[tflint/configuration.go/ToCLIArguments] Setting --module=%v\n", configuration.Module)
		args = append(args, "--module")
	}

	log.Println("[tflint/configuration.go/ToCLIArguments] Finished extracting arguments from TFLint configuration")
	return args
}
