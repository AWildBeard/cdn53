/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"cdn53/retrieval_api"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	perRequestTimeout time.Duration
	asyncLimit        uint
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download the file from the provided \"cdn\"",
	Long: `This command will download the file from the domain name provided. 

For the --tasks (-t) flag; The default value is the number of resolvers that this tool is aware of. 
So 1 task per resolver maximum with the default value (this does not mean that every resolver will 
be queried once if the total number of segments is greater than the number of resolvers. Resolver selection 
is based on PRNG cuz why not`,
	PreRun: func(cmd *cobra.Command, args []string) {
		fileLocation, err := cmd.Flags().GetString("file")
		if err != nil {
			panic("Impossible that --file doesn't have a value")
		}

		if fileLocation == "-" {
			targetFile = os.Stdout
		} else {
			targetFile, err = os.Create(fileLocation)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Failed to open file %v: %v\n", fileLocation, err)
				os.Exit(1)
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if domain == "" {
			_, _ = fmt.Fprintf(os.Stderr, "The --domain (-d) flag must be provided.\n")
			cmd.Usage()
			os.Exit(1)
		}
		err := retrieval_api.DownloadAndDecode(domain, asyncLimit, perRequestTimeout, targetFile)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to download and decode target: %v\n", err)
		}

		err = targetFile.Close()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to write decoded data to file: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().DurationVarP(&perRequestTimeout, "wait", "w", 500*time.Millisecond,
		"Timeout per request before trying the next resolver",
	)
	downloadCmd.Flags().UintVarP(&asyncLimit, "tasks", "t", uint(retrieval_api.NumberOfResolvers),
		"Number of parallel dns queries to attempt at any given time",
	)
}
