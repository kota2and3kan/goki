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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
)

// Flag value of revive command.
var reviveCmdFlags struct {
	gokiId int // Number of node (container).
}

// reviveCmd represents the revive command
var reviveCmd = &cobra.Command{
	Use:   "revive",
	Short: "Revive a specified container",
	Long: `The "goki revive" command revives a specified container.
* By default, it revives Node 1.
    goki revive
* You can specify the Node ID with -g (--goki) flag.
    goki revive -g 3
`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if err := gokiRevive(reviveCmdFlags.gokiId); err != nil {
			return err
		}

		return nil
	},
}

// gokiRevive() is called in the "goki revive" and "goki start" command.
func gokiRevive(id int) error {
	// Check the specified container is dead.
	ok, err := gokiIsDead(id)
	if err != nil {
		return err
	} else if !ok {
		fmt.Fprintln(os.Stderr, gokiResourceName+"-"+strconv.Itoa(id)+" is running")
		return errors.New("the specified container is running")
	}

	// Revive specified container.
	if err := reviveGokiContainer(id); err != nil {
		return err
	}

	return nil
}

func reviveGokiContainer(id int) error {
	// Set container name.
	container := gokiResourceName + "-" + strconv.Itoa(id)

	// Revive specified container using "docker start" command.
	c := exec.Command("docker", "start", container)
	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "docker start command failed: %v\n", string(output))
		return err
	}

	// Show revived container.
	fmt.Println("The container " + gokiResourceName + "-" + strconv.Itoa(id) + " was revived.")

	return nil
}

func init() {
	rootCmd.AddCommand(reviveCmd)
	// Flags of goki revive.
	reviveCmd.Flags().IntVarP(&reviveCmdFlags.gokiId, "goki", "g", 1, "The node ID of container that goki revive command will revive.")
}
