package config

import (
	"github.com/spf13/viper"
)

const (
	ContextDir  = ".kubecopilot"
	ContextFile = "kc-context"
	ContextType = "yaml"
)

// SetupViperConfig sets up Viper to use the shared config location.
func SetupViperConfig() {
	viper.SetConfigName(ContextFile)
	viper.SetConfigType(ContextType)
	viper.AddConfigPath(ContextDir)
}
