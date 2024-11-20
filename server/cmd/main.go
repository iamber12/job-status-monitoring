package main

import (
	"github.com/spf13/cobra"
	"log"
	"video-translation-status/server/cmd/serve"
)

func main() {
	rootCmd := &cobra.Command{
		Use:  "video-translation-status-backend",
		Long: "video-translation-status-backend",
	}

	serveCmd := serve.NewServeCommand()
	rootCmd.AddCommand(serveCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("error running command: %v", err)
	}
}
