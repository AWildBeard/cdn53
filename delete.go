package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/spf13/cobra"
	"os"
	"regexp"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a route53 \"CDN\"",
	Long: `Delete a "CDN" created by the 'create' subcommand. This operation uses regex and leverages the 
definable pattern that this tool creates in order to only delete TXT records associated with this tool`,
	Run: func(cmd *cobra.Command, args []string) {
		deleteCDN()
	},
}

func deleteCDN() {
	client := getClient()

	targetZone := getHostedZone(client)
	if targetZone != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Found target zone %v\n", *targetZone.Id)
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to find hosted zone with name %v\n", specificHostedZoneDomain)
		os.Exit(1)
	}

	query := route53.ListResourceRecordSetsInput{
		HostedZoneId:          targetZone.Id,
		MaxItems:              aws.Int32(5),
		StartRecordIdentifier: nil,
		StartRecordName:       aws.String(fmt.Sprintf("%s-", specificSubdomain)),
		StartRecordType:       "TXT",
	}

	changeSet := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &types.ChangeBatch{
			Changes: make([]types.Change, 0),
		},
		HostedZoneId: targetZone.Id,
	}

	output := &bytes.Buffer{}
	count := uint64(0)
	for {
		results, err := client.ListResourceRecordSets(context.Background(), &query)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to list RRS; %v\n", err)
			os.Exit(1)
		}

		for index, rrs := range results.ResourceRecordSets {
			grr := rrs
			match, err := regexp.MatchString(fmt.Sprintf("%s-[0-9]+\\.%s", specificSubdomain, specificHostedZoneDomain), *rrs.Name)
			if err != nil {
				panic(err)
			}

			if index+1 == len(results.ResourceRecordSets) {
				fmt.Fprintf(output, "\r\033[2KDeleted %v", *rrs.Name)
			}
			if match {
				changeSet.ChangeBatch.Changes = append(changeSet.ChangeBatch.Changes, types.Change{
					Action:            "DELETE",
					ResourceRecordSet: &grr,
				})
			}
		}

		if len(changeSet.ChangeBatch.Changes) <= 0 {
			if count != 0 {
				fmt.Fprintf(os.Stdout, "\nDeleted %v records\n", count)
			} else {
				fmt.Fprintf(os.Stdout, "Nothing found to delete\n")
			}

			break
		}

		_, err = client.ChangeResourceRecordSets(context.Background(), changeSet)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to delete RRS: %v\n", err)
			os.Exit(1)
		}
		count += uint64(len(changeSet.ChangeBatch.Changes))
		fmt.Fprint(os.Stdout, output.String())
		output.Reset()

		changeSet.ChangeBatch.Changes = make([]types.Change, 0)
	}
}

func init() {
	awsCmd.AddCommand(deleteCmd)
}
