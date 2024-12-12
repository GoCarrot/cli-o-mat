package cmd

import (
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"

	"github.com/SixtyAI/cli-o-mat/awsutil"
	"github.com/SixtyAI/cli-o-mat/config"
	"github.com/SixtyAI/cli-o-mat/util"
)

func getLaunchTemplateTagValues(tags []*ec2.Tag) (string, string) {
	var (
		application string
		service     string
	)

	for _, tag := range tags {
		key := aws.StringValue(tag.Key)
		val := aws.StringValue(tag.Value)

		switch key {
		case config.AppTag:
			application = val
		case config.ServiceTag:
			service = val
		}
	}

	return application, service
}

func showLaunchTemplates(templates []*ec2.LaunchTemplate) {
	tableData := make([][]string, len(templates))

	for idx, template := range templates {
		application, service := getLaunchTemplateTagValues(template.Tags)

		tableData[idx] = []string{
			aws.StringValue(template.LaunchTemplateId),
			aws.StringValue(template.LaunchTemplateName),
			application,
			service,
			strconv.FormatInt(aws.Int64Value(template.DefaultVersionNumber), 10),
			strconv.FormatInt(aws.Int64Value(template.LatestVersionNumber), 10),
			template.CreateTime.Format(time.RFC3339),
		}
	}

	sort.Slice(tableData, func(i, j int) bool {
		return tableData[i][1] < tableData[j][1]
	})

	tableConfig := &util.Table{
		Columns: []util.Column{
			{Name: "ID"},
			{Name: "Name"},
			{Name: "Application"},
			{Name: "Service"},
			{Name: "Default", RightAlign: true},
			{Name: "Latest", RightAlign: true},
			{Name: "Created"},
		},
	}

	tableConfig.Show(tableData)
}

// nolint: gochecknoglobals
var launchTemplatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "List launch templates.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		accountName := args[0]
		omat := loadOmatConfig(accountName)

		details := awsutil.FindAndAssumeAdminRole(omat)

		ec2Client := ec2.New(details.Session, details.Config)
		templates, err := awsutil.FetchLaunchTemplates(ec2Client, nil)
		if err != nil {
			util.Fatal(1, err)
		}

		showLaunchTemplates(templates)
	},
}

// nolint: gochecknoinits
func init() {
	rootCmd.AddCommand(launchTemplatesCmd)
}
