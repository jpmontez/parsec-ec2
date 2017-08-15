// Copyright © 2017 Jade Iqbal <jadeiqbal@fastmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"os"
	"os/exec"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new Parsec EC2 instance",
	Long: `
Sends a spot request for the requested EC2 instance type in the specified region.
If PARSEC_EC2_SERVER_KEY has not been exported in the shell rc file, it must be
passed to the command using the --server-key flag.

The amount to bid above the current lowest spot price for the instance is
specified using the --bid flag, so if the current lowest spot price is $0.20,
running the command with --bid 0.10 will send a spot request with a bid price
of $0.30.

If the --plan flag is used, the spot request will not be sent and instead the
'terraform plan' command will be run which will output to the console the details
of any AWS resources that will be created by running the start command.

Examples:

parsec-ec2 start --aws-region eu-west-1 --instance-type g3.4xlarge --bid 0.10
parsec-ec2 start --aws-region eu-west-2 --instance-type g2.2xlarge --bid 0.10 ---server-key xxxxx
parsec-ec2 start --aws-region eu-central-1 --instance-type g2.2xlarge --bid 0.10 --plan
`,
	Run: func(cmd *cobra.Command, args []string) {
		if !hasParsecServerKey(parsecServerKey) {
			fmt.Println(`
Either add 'export PARSEC_EC2_SERVER_KEY=xxxxx' in your shell rc file , or pass it to this command using the --server-key flag.`)
			os.Exit(0)
		}

		if !isValidAwsRegion(validAwsRegions, awsRegion) {
			fmt.Printf("\n'%s' is not a valid AWS region.\n", awsRegion)
			os.Exit(1)
		}

		session, err := session.NewSession()
		if err != nil {
			fmt.Println(err)
		}

		svc := ec2.New(session, &aws.Config{
			Region: aws.String(awsRegion),
		})

		vpcId, err := getVpcId(svc)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		spotPrice, err := getCheapestSpotPrice(svc, instanceType)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		userBid := calculateUserBid(*spotPrice.SpotPrice, bid)
		availabilityZone := *spotPrice.AvailabilityZone

		subnetId, err := getSubnetId(svc, availabilityZone)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		currentSessionVars := TfVars{
			SubnetId:        subnetId,
			VpcId:           vpcId,
			AwsRegion:       awsRegion,
			Ec2InstanceType: instanceType,
			ParsecServerKey: parsecServerKey,
			UserBid:         userBid,
		}

		err = writeSessionVars(currentSessionVars)
		if err != nil {
			fmt.Println(err)
		}

		var start *exec.Cmd

		if plan {
			start = constructTerraformCommand(currentSessionVars, []string{"plan"})
		} else {
			start = constructTerraformCommand(currentSessionVars, []string{"apply"})
		}

		err = executeTerraformCommandAndPrintOutput(start)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if !plan {
			fmt.Println(`
A EC2 spot request has been created. You will be able to see on the AWS dashboard when it has been
started and initialised. Once it has been initialised it will show as connectable in the Parsec
desktop application.`)
		}
	},
}

func calculateUserBid(cheapestSpotPrice string, bidIncrease float64) string {
	spotPrice, _ := strconv.ParseFloat(cheapestSpotPrice, 64)
	userBid := spotPrice + bidIncrease

	return fmt.Sprint(userBid)
}

var (
	bid             float64
	parsecServerKey string
	plan            bool
)

func init() {
	RootCmd.AddCommand(startCmd)
	startCmd.Flags().Float64VarP(&bid, "bid", "b", 0.00, "amount to bid above the current lowest spot price")
	startCmd.Flags().StringVarP(&parsecServerKey, "server-key", "k", "", "Parsec server key")
	startCmd.Flags().BoolVarP(&plan, "plan", "p", false, "plan out the resources to be created without creating them")
}
