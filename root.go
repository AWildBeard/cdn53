package main

import (
	"github.com/spf13/cobra"
	"os"
)

var (
	targetFile               *os.File = nil
	domain                   string
	specificHostedZoneDomain string
	specificSubdomain        string
	//encoding                 string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cdn53",
	Short: "Tool and library to use DNS TXT records (on route53) as a CDN",
	Long: `This tooling has the ability to;
1) Transfer a local file with encoding into DNS TXT records on route53
2) Delete those TXT records from route53 (cleanup)
3) Download the file and decode it via DNS`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		domain, _ = cmd.Flags().GetString("domain")
	},
	//Run: func(cmd *cobra.Command, args []string) {
	//
	//},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute() // Parses flags and what not
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("file", "f", "-",
		"File to read or write content to/from depending on subcommand. If using the create subcommand, this "+
			"file is read from to create the dns cdn after applying encoding. If using the download subcommand, "+
			"this file is written to with decoding applied.",
	)
	rootCmd.PersistentFlags().StringP("domain", "d", "",
		"The domain to use for various options (`[create delete download]`). With the `create` subcommand, this is the FQDN to create a "+
			"\"cdn\" with. With the `delete` subcommand, this is the FQDN to delete a \"cdn\". With the `download` "+
			"subcommand, this is the FQDN of the \"cdn\" just as you would've supplied it to the other subcommands. "+
			"If you created the \"cdn\" with x.y.z.com then the value should be x.y.z.com",
	)
	//rootCmd.PersistentFlags().StringP("encoding", "e", "base64",
	//	"The encoding system to use for encoding and decoding operations. Valid values are base64 and "+
	//		"base85 (RFC1924)",
	//)
}
