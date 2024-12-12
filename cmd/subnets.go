package cmd

import (
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"

	"github.com/SixtyAI/cli-o-mat/awsutil"
	"github.com/SixtyAI/cli-o-mat/config"
	"github.com/SixtyAI/cli-o-mat/util"
)

func getSubnetTagValues(tags []*ec2.Tag) (string, string, string, string) {
	var (
		application string
		service     string
		netType     string
		name        string
	)

	for _, tag := range tags {
		key := aws.StringValue(tag.Key)
		val := aws.StringValue(tag.Value)

		switch key {
		case config.AppTag:
			application = val
		case config.ServiceTag:
			service = val
		case config.TypeTag:
			netType = val
		case config.NameTag:
			name = val
		}
	}

	return application, service, netType, name
}

func showSubnets(subnets []*ec2.Subnet) {
	tableData := make([][]string, len(subnets))

	//nolint: varnamelen
	sort.SliceStable(subnets, func(i, j int) bool {
		switch strings.Compare(aws.StringValue(subnets[i].VpcId), aws.StringValue(subnets[j].VpcId)) {
		case -1:
			return true
		case 1:
			return false
		default:
			return aws.StringValue(subnets[i].SubnetId) < aws.StringValue(subnets[j].SubnetId)
		}
	})

	for idx, subnet := range subnets {
		application, service, netType, name := getSubnetTagValues(subnet.Tags)

		tableData[idx] = []string{
			aws.StringValue(subnet.VpcId),
			aws.StringValue(subnet.SubnetId),
			name,
			netType,
			application,
			service,
			awsutil.DefaultToString(subnet.DefaultForAz),
			aws.StringValue(subnet.AvailabilityZone),
			aws.StringValue(subnet.AvailabilityZoneId),
			aws.StringValue(subnet.CidrBlock),
			aws.StringValue(subnet.State),
		}
	}

	tableConfig := &util.Table{
		Columns: []util.Column{
			{Name: "VPC"},
			{Name: "ID"},
			{Name: "Name"},
			{Name: "Type"},
			{Name: "Application"},
			{Name: "Service"},
			{Name: "Default?"},
			{Name: "Zone"},
			{Name: "Zone ID"},
			{Name: "CIDR"},
			{Name: "State"},
		},
	}

	tableConfig.Show(tableData)
}

// nolint: gochecknoglobals
var subnetsCmd = &cobra.Command{
	Use:   "subnets account",
	Short: "Show details about the available subnets.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		accountName := args[0]
		omat := loadOmatConfig(accountName)

		details := awsutil.FindAndAssumeAdminRole(omat)

		ec2Client := ec2.New(details.Session, details.Config)

		subnets, err := awsutil.FetchSubnets(ec2Client)
		if err != nil {
			util.Fatal(1, err)
		}

		showSubnets(subnets)
	},
}

// nolint: gochecknoinits
func init() {
	rootCmd.AddCommand(subnetsCmd)
}
