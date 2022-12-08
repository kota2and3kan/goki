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

// Flag value of start command.
var startCmdFlags struct {
	gokiId int // Number of node (container).
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a specified container (Alias of \"revive\" command)",
	Long: `The "goki start" command starts a specified container. This command is an alias of "goki revive" command.
* By default, it starts Node 1.
    goki start
* You can specify the Node ID with -g (--goki) flag.
    goki start -g 3
`,
	RunE: func(cmd *cobra.Command, args []string) error {

		// Call gokiRevive(). The "goki start" command has the same behavior as the "goki revive" command.
		if err := gokiRevive(startCmdFlags.gokiId); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	// Flags of goki start.
	startCmd.Flags().IntVarP(&startCmdFlags.gokiId, "goki", "g", 1, "The node ID of container that goki start command will start.")
}
