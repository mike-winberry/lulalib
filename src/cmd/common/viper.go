package common

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mike-winberry/lulalib/src/internal/template"
	"github.com/mike-winberry/lulalib/src/pkg/message"
	"github.com/spf13/viper"
)

const (
	VLogLevel  = "log_level"
	VTarget    = "target"
	VSummary   = "summary"
	VConstants = "constants"
	VVariables = "variables"
)

var (
	// Viper instance used by commands
	v *viper.Viper

	// Viper configuration error
	vConfigError error

	// Template config values
	TemplateConstants map[string]interface{}

	// Template config values
	TemplateVariables []template.VariableConfig
)

// InitViper initializes the viper singleton for the CLI
func InitViper() *viper.Viper {
	// Already initialized by some other command
	if v != nil {
		return v
	}

	v = viper.New()

	// Skip for the version command
	if isVersionCmd() {
		return v
	}

	// Specify an alternate config file
	cfgFile := os.Getenv("LULA_CONFIG")

	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile != "" {
		// Use config file from the flag.
		v.SetConfigFile(cfgFile)
	} else {
		// Search config paths in the current directory and $HOME/.lula.
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.lula")
		v.SetConfigName("lula-config")
	}

	// E.g. LULA_LOG_LEVEL=debug
	v.SetEnvPrefix("lula")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Optional, so ignore errors
	vConfigError = v.ReadInConfig()

	// Set default values for viper
	setDefaults()

	// Load template config
	constants, variables, err := GetTemplateConfig()
	if err != nil {
		panic(err)
	}
	TemplateConstants = constants
	TemplateVariables = variables

	return v
}

// GetViper returns the viper singleton
func GetViper() *viper.Viper {
	return v
}

// GetTemplateConfig loads the constants and variables from the viper config
func GetTemplateConfig() (map[string]interface{}, []template.VariableConfig, error) {
	constants := make(map[string]interface{})
	variables := make([]template.VariableConfig, 0)

	err := v.UnmarshalKey(VConstants, &constants)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to unmarshal constants into map: %v", err)
	}

	err = v.UnmarshalKey(VVariables, &variables)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to unmarshal variables into slice: %v", err)
	}

	return constants, variables, nil
}

func isVersionCmd() bool {
	args := os.Args
	return len(args) > 1 && (args[1] == "version" || args[1] == "v")
}

func setDefaults() {
	v.SetDefault(VLogLevel, "info")
	v.SetDefault(VSummary, false)
	v.SetDefault(VConstants, make(map[string]interface{}))
	v.SetDefault(VVariables, make([]interface{}, 0))
}

func printViperConfigUsed() {
	// Only print config info if viper is initialized.
	vInitialized := v != nil
	if !vInitialized {
		return
	}
	var notFoundErr viper.ConfigFileNotFoundError
	if errors.As(vConfigError, &notFoundErr) {
		return
	}
	if vConfigError != nil {
		message.WarnErrf(vConfigError, "failed to load config file: %s", vConfigError.Error())
		return
	}
	message.Notef("Using config file %s", v.ConfigFileUsed())
}
