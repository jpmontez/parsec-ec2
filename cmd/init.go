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

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialisation command to be run after initial installs and subsequent upgrades",
	Long: `
Copies the terraform and Windows provisioning user data template to the
$HOME/.parsec-ec2 directory and runs 'terraform init' to initialise all
the things required by the terraform template.

If upgrading an existing installation, the $HOME/.parsec-ec2 folder should
be manually removed before running the init command run again.

After running this command it is recommended to export the environment
variable PARSEC_EC2_SERVER_KEY in your shell rc file so that it doesn't
need to be passed in manually every time the start command is run.
`,
	Run: func(cmd *cobra.Command, args []string) {
		terraformFilePath := fmt.Sprintf("%s/%s", projectFolder, ParsecTemplate)
		userDataFilePath := fmt.Sprintf("%s/%s", projectFolder, UserDataTemplate)

		fileInfo, _ := os.Stat(appFolder)

		if fileInfo != nil {
			fmt.Println(`
The init command has already been run on this machine. If you wish to run
it again you must manually delete the $HOME/.parsec-ec2 folder first.`)
			os.Exit(0)
		}

		err := os.Mkdir(appFolder, 0755)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = copyFile(terraformFilePath, ParsecTemplate, appFolder)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = copyFile(userDataFilePath, UserDataTemplate, appFolder)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		init := exec.Command(Terraform, Init)
		init.Dir = appFolder

		err = executeTerraformCommandAndPrintOutput(init)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}
