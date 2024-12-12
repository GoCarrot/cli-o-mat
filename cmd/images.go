package cmd

import (
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"

	"github.com/SixtyAI/cli-o-mat/awsutil"
	"github.com/SixtyAI/cli-o-mat/config"
	"github.com/SixtyAI/cli-o-mat/util"
)

func showImages(shortHashes bool, images []*ec2.Image) {
	hashLen := 40
	if shortHashes {
		hashLen = 7
	}

	tableData := make([][]string, len(images))

	for idx, image := range images {
		var commit string

		for _, tag := range image.Tags {
			if aws.StringValue(tag.Key) == config.CommitTag {
				commit = aws.StringValue(tag.Value)

				if shortHashes && commit != "" {
					commit = commit[0:hashLen]
				}

				break
			}
		}

		tableData[idx] = []string{
			aws.StringValue(image.ImageId),
			aws.StringValue(image.Architecture),
			aws.StringValue(image.State),
			commit,
			aws.StringValue(image.Name),
			strconv.FormatInt(*image.BlockDeviceMappings[0].Ebs.VolumeSize, 10),
			aws.StringValue(image.CreationDate),
		}
	}

	sort.Slice(tableData, func(i, j int) bool {
		return tableData[i][5] < tableData[j][5]
	})

	tableConfig := &util.Table{
		Columns: []util.Column{
			{Name: "ID"},
			{Name: "Arch"},
			{Name: "State"},
			{Name: "Commit"},
			{Name: "Name"},
			{Name: "Size"},
			{Name: "Created"},
		},
	}

	tableConfig.Show(tableData)
}

// nolint: gochecknoglobals
var imagesCmd = &cobra.Command{
	Use:   "images account",
	Short: "List AMIs.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		accountName := args[0]
		omat := loadOmatConfig(accountName)

		details := awsutil.FindAndAssumeAdminRole(omat)

		ec2Client := ec2.New(details.Session, details.Config)

		imageList, err := ec2Client.DescribeImages(&ec2.DescribeImagesInput{
			Owners: []*string{aws.String("self")},
		})
		if err != nil {
			util.Fatal(1, err)
		}

		showImages(imagesShortHashes, imageList.Images)
	},
}

// nolint: gochecknoglobals
var imagesShortHashes bool

// nolint: gochecknoinits
func init() {
	rootCmd.AddCommand(imagesCmd)

	imagesCmd.Flags().BoolVarP(&imagesShortHashes, "short", "", false, "Shorten git commit SHAs")
}
