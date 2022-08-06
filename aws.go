/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var (
	profile string
	region  string
)

func getClient() *route53.Client {
	if cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigProfile(profile), config.WithDefaultRegion(region)); err == nil {
		if cfg.Region == "" {
			_, _ = fmt.Fprintf(os.Stderr, "Either; \n 1) You did not provide a profile (--profile)"+
				"\n 2) The profile you provided does not exist "+
				"\n 3) The profile you provided does not have a region set "+
				"\n 4) The default profile does not exist (used when --profile isn't set)"+
				"\n 5) The default profile does not have a region set "+
				"\n 6) You did not provide a region (--region). "+
				"\n\nIf you do not have a profile configured in aws config files, you must pass --region "+
				"(assuming you're providing aws credentials outside of the aws config files & directories "+
				"via aws SDK environment variables, etc). If you have a profile (and aws credentials, "+
				"etc specified in the aws config files); the profile must have a configured region "+
				"which you can set by passing --region or by setting one in the aws config file.\n",
			)
			os.Exit(1)
		}

		return route53.NewFromConfig(cfg)
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "failed to load aws config: %v\n", err)
		os.Exit(1)
	}

	return nil
}

func getHostedZone(client *route53.Client) *types.HostedZone {
	zoneList, err := client.ListHostedZones(context.Background(), nil)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to list hosted zones with dnsname %v: %v\n", domain, err)
		os.Exit(1)
	}

	var targetZone *types.HostedZone
	for _, zone := range zoneList.HostedZones {
		if *zone.Name == specificHostedZoneDomain+"." {
			targetZone = &zone
			break
		}
	}

	return targetZone
}

// awsCmd represents the aws command
var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Create or Delete \"cdn\"s on route53.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Parent().PersistentPreRun != nil {
			cmd.Parent().PersistentPreRun(cmd.Parent(), args)
		}

		profile, _ = cmd.Flags().GetString("profile")
		region, _ = cmd.Flags().GetString("region")

		specificHostedZoneDomain, _ = cmd.Flags().GetString("hosted-zone-domain")
		specificSubdomain, _ = cmd.Flags().GetString("sub-domain")
		if domain != "" && specificSubdomain != "" && specificHostedZoneDomain != "" {
			_, _ = fmt.Fprintf(os.Stderr, "You cannot specify the --domain (-d) flag with --sub-domain "+
				"and/or --hosted-zone-domain\n",
			)
		} else if domain != "" {
			splitDomain := strings.SplitN(domain, ".", 2)
			specificSubdomain = splitDomain[0]
			specificHostedZoneDomain = splitDomain[1]
		} else if specificSubdomain != "" && specificHostedZoneDomain != "" {
			domain = specificSubdomain + specificHostedZoneDomain
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "either the --domain flag was not provided or; "+
				"the --hosted-zone-domain flag and --sub-domain flag were not provided\n",
			)
			_ = cmd.Help()
			os.Exit(1)
		}
	},
}

func init() {
	awsCmd.PersistentFlags().StringP("profile", "p", "",
		"AWS profile to use",
	)
	awsCmd.PersistentFlags().StringP("region", "r", "",
		"AWS region to use",
	)
	awsCmd.PersistentFlags().String("hosted-zone-domain", "", "Set "+
		"the specific name of the hosted zone to build a \"cdn\" under. Must be provided with --sub-domain",
	)
	awsCmd.PersistentFlags().String("sub-domain", "", "Set the specific "+
		"subdomain to build a \"cdn\" with. Must be provided with --hosted-zone-domain",
	)
	awsCmd.MarkFlagsRequiredTogether("hosted-zone-domain", "sub-domain")

	rootCmd.AddCommand(awsCmd)
}
