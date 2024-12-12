package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"

	"github.com/SixtyAI/cli-o-mat/awsutil"
	"github.com/SixtyAI/cli-o-mat/util"
)

const (
	LaunchWaitTimeout = 30
)

// nolint: gochecknoglobals,gomnd
var launchCmd = &cobra.Command{
	Use:   "launch account template-name keypair-name [subnet-id]",
	Short: "Launch an EC2 instance from a launch template.",
	Long: `Launch an EC2 instance from a launch template.

If you don't specify a subnet-id, the default subnet from the launch template
will be used.`,
	Args: cobra.RangeArgs(3, 4), // nolint: mnd
	Run: func(_ *cobra.Command, args []string) {
		accountName := args[0]
		namePrefix := args[1]
		keypair := args[2]

		var subnetID *string
		if len(args) == 4 { // nolint: mnd
			subnetID = aws.String(args[3])
		}

		omat := loadOmatConfig(accountName)

		details := awsutil.FindAndAssumeAdminRole(omat)

		ec2Client := ec2.New(details.Session, details.Config)

		if launchVersion == "" {
			launchVersion = "$Latest"
		}

		var instanceType *string
		if launchType != "" {
			instanceType = aws.String(launchType)
		}

		templates, err := awsutil.FetchLaunchTemplates(ec2Client, nil)
		if err != nil {
			util.Fatal(1, err)
		}
		candidates := make([]string, 0)
		for _, template := range templates {
			templateName := aws.StringValue(template.LaunchTemplateName)
			if strings.HasPrefix(templateName, namePrefix) {
				candidates = append(candidates, templateName)
			}
		}

		if len(candidates) == 0 {
			fmt.Printf("Found the following launch templates, none of which match specified prefix:\n")
			for _, template := range templates {
				fmt.Printf("\t%s\n", aws.StringValue(template.LaunchTemplateName))
			}
			util.Fatalf(1, "No matching launch templates found.\n")
		} else if len(candidates) > 1 {
			fmt.Printf("Found the following launch templates matching specified prefix:\n")
			for _, candidate := range candidates {
				fmt.Printf("\t%s\n", candidate)
			}
			util.Fatalf(1, "Multiple launch templates found.\n")
		}
		name := candidates[0]
		fmt.Printf("Using launch template %s...\n", name)

		input := ec2.RunInstancesInput{
			LaunchTemplate: &ec2.LaunchTemplateSpecification{
				LaunchTemplateName: &name,
				Version:            aws.String(launchVersion),
			},
			InstanceType:                      instanceType,
			InstanceInitiatedShutdownBehavior: aws.String("terminate"),
			KeyName:                           aws.String(keypair),

			MinCount: aws.Int64(1),
			MaxCount: aws.Int64(1),
			SubnetId: subnetID,
		}
		if volumeSize > 0 {
			input.BlockDeviceMappings = []*ec2.BlockDeviceMapping{
				{
					DeviceName: aws.String("/dev/xvda"),
					Ebs: &ec2.EbsBlockDevice{
						VolumeSize: aws.Int64(volumeSize),
					},
				},
			}
		}

		resp, err := ec2Client.RunInstances(&input)
		if err != nil {
			util.Fatal(1, err)
		}

		if len(resp.Instances) != 1 {
			util.Fatalf(1, "Unable to launch EC2 instance.\n")
		}

		fmt.Printf("Launching instance %s...\n", aws.StringValue(resp.Instances[0].InstanceId))
		fmt.Printf("Waiting for instance to have a public IP...\n")

		counter := 0
		instanceIDs := []*string{resp.Instances[0].InstanceId}
		publicIP := ""

		for {
			<-time.After(1 * time.Second)

			counter++
			if counter > LaunchWaitTimeout {
				break
			}

			resp, err := ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{
				InstanceIds: instanceIDs,
			})
			if err != nil {
				util.Fatal(1, err)
			}

			if aws.StringValue(resp.Reservations[0].Instances[0].PublicIpAddress) != "" {
				publicIP = aws.StringValue(resp.Reservations[0].Instances[0].PublicIpAddress)

				break
			}
		}

		if publicIP != "" {
			fmt.Printf("Public IP: %s\n", publicIP)
		} else {
			fmt.Printf("Couldn't determine public IP.\n")
		}
	},
}

// nolint: gochecknoglobals
var (
	launchVersion string
	launchType    string
	volumeSize    int64
)

// nolint: gochecknoinits
func init() {
	rootCmd.AddCommand(launchCmd)
	launchCmd.Flags().StringVarP(&launchVersion, "version", "", "", "Version of launch template to use (default: $Latest)")
	launchCmd.Flags().StringVarP(&launchType, "type", "", "", "Instance type to launch (default from launch template)")
	launchCmd.Flags().Int64VarP(&volumeSize, "size", "", 0, "Size of EBS volume in GB (omit for default)")
}
