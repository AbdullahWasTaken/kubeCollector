package cmd

import (
	"os"

	"github.com/AbdullahWasTaken/kubeCollector/collector"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var outDir string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kubeCollector <kubeconfig>",
	Short: "A kubernetes cluster state collector",
	Long: `KubeCollector is a general purpose state collector for kubernetes clusters 
	without any restrictions on the third-party resources installed on the server.

	Examples:
		# Collect the state into default location
		kubeCollector /.kube/config
	
		# Collect the state into specified location
		kubeCollector /.kube/config -out /outputDir`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		collector.Collect(args[0], outDir)
		log.Info("Kubernetes cluster state saved to ", outDir)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&outDir, "out", "out", "output directory")
}
