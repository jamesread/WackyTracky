package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/wacky-tracky/wacky-tracky-server/pkg/runtimeconfig"
	"github.com/wacky-tracky/wacky-tracky-server/pkg/singleFrontend"
)

var root = &cobra.Command{
	Use: "wt",
	Run: mainRoot,
}

func mainRoot(cmd *cobra.Command, args []string) {
	singleFrontend.StartServers(runtimeconfig.RuntimeConfig.DB)
}

func main() {
	log.Info("wacky-tracky")

	root.PersistentFlags().StringVarP(&runtimeconfig.RuntimeConfig.DB, "db", "D", "neo4j", "The database to use")

	root.Execute()
}
