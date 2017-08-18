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

	"github.com/spf13/cobra"
)

// priceCmd represents the price command
var priceCmd = &cobra.Command{
	Use:   "price",
	Short: "Get the highest spot price for an instance type in a given region",
	Long: `
Looks for the current highest spot price for the requested instance type
in the requested region.

Example:

parsec-ec2 price --region eu-west-1 --instance-type g2.2xlarge
`,
	Run: func(cmd *cobra.Command, args []string) {
		if !isValidRegion(ec2Regions(), region) {
			fmt.Printf("%s is not a valid AWS region id.\n", region)
			os.Exit(1)
		}

		if !isValidGInstance(gInstances(), instanceType) {
			fmt.Printf("%s is not a valid EC2 GPU instance type id.\n", instanceType)
			os.Exit(1)
		}

		ec2Client, err := getEc2Client(region)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		spotPrice, err := getSpotPrice(ec2Client, instanceType)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		dollarPrice := *spotPrice.SpotPrice

		fmt.Printf("The highest spot price in the %s region for %s instances is currently $%s/hour.\n", region, instanceType, dollarPrice)
	},
}

func init() {
	RootCmd.AddCommand(priceCmd)
}
