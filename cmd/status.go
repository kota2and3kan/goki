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
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show container status",
	Long: `Show container status.
You can kill/start containers of Goki using "goki jet (or kill)" and "goki revive (or start)" command.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if n, err := getNumberOfContainers(); err != nil {
			return err
		} else if n == 0 {
			fmt.Fprintln(os.Stderr, "There is no containers of Goki.")
			return nil
		} else if n > 0 {
			if err = showContainerStatus(n); err != nil {
				return err
			}
		}

		return nil
	},
}

func getNumberOfContainers() (int, error) {
	// List of all containers that includes stopped (killed) containers.
	var gokiList []string = []string{}

	// Get list of stopped containers to remove them.
	c := exec.Command("docker", "ps", "-aqf", "label="+gokiResourceLabel, "--format", "{{.Names}}")

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "docker ps command failed: %v\n", string(output))
		return -1, err
	} else if string(output) != "" {
		gokiList = strings.FieldsFunc(string(output), func(r rune) bool { return r == '\n' })
	}

	return len(gokiList), nil
}

func showContainerStatus(n int) error {
	// List of running containers.
	var liveGokiList []string = []string{}
	// List of stopped (killed) containers.
	var deadGokiList []string = []string{}

	// getNumberOfContainers() returns the number of all containers include "goki-client".
	// We need to check from "goki-1" to "goki-n" other tan "goki-client".
	for i := 1; i <= n-1; i++ {
		if dead, err := gokiIsDead(i); err != nil {
			return err
		} else if dead {
			deadGokiList = append(deadGokiList, gokiResourceName+"-"+strconv.Itoa(i))
		} else if !dead {
			liveGokiList = append(liveGokiList, gokiResourceName+"-"+strconv.Itoa(i))
		}
	}

	// Show alive containers.
	fmt.Fprintln(os.Stdout, "Alive containers:")
	if len(liveGokiList) == 0 {
		fmt.Fprintln(os.Stdout, "  Nothing")
	}
	for i := 0; i < len(liveGokiList); i++ {
		fmt.Fprintln(os.Stdout, "  "+liveGokiList[i])
	}

	// Show dead containers.
	fmt.Fprintln(os.Stdout, "Dead containers:")
	if len(deadGokiList) == 0 {
		fmt.Fprintln(os.Stdout, "  Nothing")
	}
	for i := 0; i < len(deadGokiList); i++ {
		fmt.Fprintln(os.Stdout, "  "+deadGokiList[i])
	}

	return nil
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
