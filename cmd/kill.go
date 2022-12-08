// Copyright 2022 kota2and3kan
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
	"github.com/spf13/cobra"
)

// Flag value of kill command.
var killCmdFlags struct {
	gokiId int // Number of node (container).
}

// killCmd represents the kill command
var killCmd = &cobra.Command{
	Use:   "kill",
	Short: "Kill a specified container (Alias of \"jet\" command)",
	Long: `The "goki kill" command kills a specified container. This command is an alias of "goki jet" command.
* By default, it kills Node 1.
    goki kill
* You can specify the Node ID with -g (--goki) flag.
    goki kill -g 3
`,
	RunE: func(cmd *cobra.Command, args []string) error {

		// Call gokiJet(). The "goki kill" command has the same behavior as the "goki jet" command.
		if err := gokiJet(killCmdFlags.gokiId); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(killCmd)
	// Flags of goki kill.
	killCmd.Flags().IntVarP(&killCmdFlags.gokiId, "goki", "g", 1, "The node ID of container that goki kill command will kill.")
}
