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

// Flag value of jet command.
var jetCmdFlags struct {
	gokiId int // Number of node (container).
}

// jetCmd represents the jet command.
var jetCmd = &cobra.Command{
	Use:   "jet",
	Short: "Kill a specified container",
	Long: `The "goki jet" command kills a specified container.
* By default, it kills Node 1.
    goki jet
* You can specify the Node ID with -g (--goki) flag.
    goki jet -g 3
`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if err := gokiJet(jetCmdFlags.gokiId); err != nil {
			return err
		}

		return nil
	},
}

// gokiJet() is called in the "goki jet" and "goki kill" command.
func gokiJet(id int) error {
	// Check the specified container is running.
	ok, err := gokiIsDead(id)
	if err != nil {
		return err
	} else if ok {
		fmt.Fprintln(os.Stderr, gokiResourceName+"-"+strconv.Itoa(id)+" is not running")
		return errors.New("the specified container is not running")
	}

	// Kill specified container.
	if err := killGokiContainer(id); err != nil {
		return err
	}

	return nil
}

func gokiContains(list []string, name string) bool {
	for _, v := range list {
		if name == v {
			return true
		}
	}
	return false
}

func killGokiContainer(id int) error {
	// Set container name.
	container := gokiResourceName + "-" + strconv.Itoa(id)

	// Kill specified container using "docker kill" command.
	c := exec.Command("docker", "kill", container)
	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "docker kill command failed: %v\n", string(output))
		return err
	}

	// Show killed container.
	fmt.Println("The container " + gokiResourceName + "-" + strconv.Itoa(id) + " was killed.")

	return nil
}

func init() {
	rootCmd.AddCommand(jetCmd)
	// Flags of goki jet.
	jetCmd.Flags().IntVarP(&jetCmdFlags.gokiId, "goki", "g", 1, "The node ID of container that goki jet command will kill.")
}
