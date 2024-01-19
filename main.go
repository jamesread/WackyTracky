package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wacky-tracky/wacky-tracky-server/pkg/runtimeconfig"
	"github.com/wacky-tracky/wacky-tracky-server/pkg/singleFrontend"
)

var root = &cobra.Command{
	Use: "wt",
	Run: mainRoot,
}

func mainRoot(cmd *cobra.Command, args []string) {
	singleFrontend.StartServers()
}

func disableLogTimestamps() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors:    false,
		DisableTimestamp: true,
	})
}

func initViperConfig() {
	viper.AutomaticEnv()
	viper.SetConfigName("config.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/config") // For containers.

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Infof("Config file not found %v", err)
		} else {
			log.Errorf("Config file error at startup. %s", err)
			os.Exit(1)
		}
	} else {
		log.WithFields(log.Fields{
			"file": viper.ConfigFileUsed(),
		}).Infof("Starting to read a config")
	}

	if err := viper.UnmarshalExact(&runtimeconfig.RuntimeConfig); err != nil {
		log.Fatalf("Config unmarshal error %v", err)
	}
}

func main() {
	disableLogTimestamps()

	log.Info("wacky-tracky")

	initViperConfig()

	root.Execute()
}
