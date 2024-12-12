package cmd

import (
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"

	"github.com/SixtyAI/cli-o-mat/awsutil"
	"github.com/SixtyAI/cli-o-mat/config"
	"github.com/SixtyAI/cli-o-mat/util"
)

func getHostTagValues(tags []*ec2.Tag) (string, string, string, string) {
	var (
		application           string
		service               string
		asg                   string
		launchTemplateVersion string
	)

	for _, tag := range tags {
		key := aws.StringValue(tag.Key)
		val := aws.StringValue(tag.Value)

		switch key {
		case config.AppTag:
			application = val
		case config.ServiceTag:
			service = val
		case config.ASGTag:
			asg = val
		case config.LTVersionTag:
			launchTemplateVersion = val
		}
	}

	return application, service, asg, launchTemplateVersion
}

func showHosts(hosts []*ec2.Instance) {
	tableData := make([][]string, len(hosts))

	for idx, host := range hosts {
		application, service, asg, launchTemplateVersion := getHostTagValues(host.Tags)

		var stateName string

		state := host.State
		if state != nil {
			stateName = aws.StringValue(state.Name)
		}

		tableData[idx] = []string{
			aws.StringValue(host.InstanceId),
			host.LaunchTime.Format(time.RFC3339),
			aws.StringValue(host.InstanceType),
			aws.StringValue(host.Architecture),
			aws.StringValue(host.ImageId),
			stateName,
			aws.StringValue(host.PublicIpAddress),
			application,
			service,
			launchTemplateVersion,
			asg,
			aws.StringValue(host.KeyName),
		}
	}

	sort.Slice(tableData, func(i, j int) bool {
		return tableData[i][1] < tableData[j][1]
	})

	tableConfig := &util.Table{
		Columns: []util.Column{
			{Name: "ID"},
			{Name: "Launched"},
			{Name: "Type"},
			{Name: "Arch"},
			{Name: "Image"},
			{Name: "State"},
			{Name: "Public IP"},
			{Name: "Application"},
			{Name: "Service"},
			{Name: "Ver.", RightAlign: true},
			{Name: "ASG"},
			{Name: "Keypair"},
		},
	}

	tableConfig.Show(tableData)
}

// nolint: gochecknoglobals
var hostsCmd = &cobra.Command{
	Use:   "hosts account",
	Short: "List EC2 instances.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		accountName := args[0]
		omat := loadOmatConfig(accountName)

		details := awsutil.FindAndAssumeAdminRole(omat)

		ec2Client := ec2.New(details.Session, details.Config)
		nextToken := aws.String("")
		hosts := []*ec2.Instance{}

		for {
			hostResp, err := ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{
				NextToken: nextToken,
			})
			if err != nil {
				util.Fatal(1, err)
			}

			for _, res := range hostResp.Reservations {
				hosts = append(hosts, res.Instances...)
			}

			if hostResp.NextToken == nil {
				break
			}

			nextToken = hostResp.NextToken
		}

		showHosts(hosts)
	},
}

// nolint: gochecknoinits
func init() {
	rootCmd.AddCommand(hostsCmd)
}
