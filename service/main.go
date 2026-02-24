package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wacky-tracky/wacky-tracky-server/internal/buildinfo"
	"github.com/wacky-tracky/wacky-tracky-server/internal/httpserver"
	"github.com/wacky-tracky/wacky-tracky-server/internal/runtimeconfig"
)

var root = &cobra.Command{
	Use: "wt",
	Run: mainRoot,
}

func mainRoot(cmd *cobra.Command, args []string) {
	httpserver.StartServer()
}

func disableLogTimestamps() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors:    false,
		DisableTimestamp: true,
	})
}

func initViperConfig() {
	viper.AutomaticEnv()
	viper.SetDefault("listenAddress", runtimeconfig.RuntimeConfig.ListenAddress)
	viper.SetDefault("database.driver", runtimeconfig.RuntimeConfig.Database.Driver)
	viper.SetDefault("database.database", runtimeconfig.RuntimeConfig.Database.Database)
	viper.BindEnv("database.driver", "DATABASE_DRIVER")
	viper.BindEnv("database.database", "DATABASE_DATABASE")
	viper.BindEnv("database.hostname", "DATABASE_HOSTNAME")
	viper.BindEnv("database.username", "DATABASE_USERNAME")
	viper.BindEnv("database.password", "DATABASE_PASSWORD")
	viper.BindEnv("database.port", "DATABASE_PORT")
	viper.SetConfigName("config.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../")
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

	log.WithFields(log.Fields{
		"version": buildinfo.Version,
		"commit":  buildinfo.Commit,
	}).Info("wacky-tracky")

	initViperConfig()

	root.Execute()
}
