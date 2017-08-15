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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the initialisation status of a launched EC2 instance",
	Long: `
Queries the launched instance and gets the current initialisation status.

Once an instance is reporting a status of initialised, it may still take
some time for the instance to show up in the Parsec desktop application
as time is still required for the provisioning script to run on instance
which is what will allow the Parsec application to launch and log in with
the provided Parsec server key.
`,
	Run: func(cmd *cobra.Command, args []string) {
		currentSessionFile := fmt.Sprintf("%s/currentSession.json", appFolder)

		bytes, err := ioutil.ReadFile(currentSessionFile)
		if err != nil {
			fmt.Println(`
There are no sessions currently running.`)
			os.Exit(0)
		}

		var currentSessionVars TfVars

		err = json.Unmarshal(bytes, &currentSessionVars)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		session, err := session.NewSession()
		if err != nil {
			fmt.Println(err)
		}

		svc := ec2.New(session, &aws.Config{
			Region: aws.String(currentSessionVars.AwsRegion),
		})

		refresh := constructTerraformCommand(currentSessionVars, []string{"refresh"})

		err = executeTerraformCommandAndSwallowOutput(refresh)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		output := constructTerraformCommand(currentSessionVars, []string{"output", "spot_instance_id"})

		spotInstanceId, err := executeTerraformCommandAndReturnOutput(output)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if len(spotInstanceId) < 1 {
			fmt.Println(`
The spot instance request has not yet been filled.`)
			os.Exit(0)
		}

		instanceIds := []*string{&spotInstanceId}

		describeInstanceStatusInput := ec2.DescribeInstanceStatusInput{
			InstanceIds: instanceIds,
		}

		describeInstanceStatusOutput, err := svc.DescribeInstanceStatus(&describeInstanceStatusInput)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		instanceStatuses := describeInstanceStatusOutput.InstanceStatuses

		if len(instanceStatuses) < 1 {
			fmt.Println(`
The spot instance request has been filled but initialisation status is
not yet available.`)
			os.Exit(0)
		}

		instanceStatus := instanceStatuses[0].InstanceStatus.Status

		if *instanceStatus == "ok" {
			fmt.Println(`
The spot instance has finished initialising and should either already or
very shortly be visible on the Parsec desktop application.`)
			os.Exit(0)
		} else {
			fmt.Println(`
The spot instance has not yet finished initialising. It will not yet be
visible on the Parsec desktop application.`)
			os.Exit(0)
		}
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
