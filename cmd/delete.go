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
	"strings"

	"github.com/spf13/cobra"
)

// Flag value of delete command.
var deleteCmdFlags struct {
	volume bool // If true, delete all docker volumes related to Goki.
}

// deleteCmd represents the delete command.
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the CockroachDB Local Cluster",
	Long: `The "goki delete" command deletes the CockroachDB Cluster in your local environment.
* By default, it deletes Containers and Docker Network (Docker Volume that includes DB Data will not be deleted).
    goki delete
* You can delete Docker Volume (DB data) with -v (--volume) flag.
    goki delete -v
`,
	RunE: func(cmd *cobra.Command, args []string) error {

		fmt.Println("*** Start Deleting CockroachDB Local Cluster ***")
		fmt.Println("INFO: Delete the following resources.")

		// Delete CockroachDB Local Cluster.
		if err := deleteGokiCluster(); err != nil {
			return err
		}

		// Deleting CockroachDB Local Cluster done.
		fmt.Printf("\n*** Deleting CockroachDB Local Cluster done ***")
		if deleteCmdFlags.volume {
			fmt.Printf("\nINFO: All docker resources (include docker volume) are deleted.\n")
		} else {
			fmt.Printf("\nINFO: You can re-create the cluster with DB Data of deleted cluster (you can re-use deleted cluster's Data).")
			fmt.Printf("\n      If you want to re-use old Data, re-run \"goki create\" command.")
			fmt.Printf("\nINFO: Docker Volumes are not deleted.")
			fmt.Printf("\n      If you don't need the DB Data of deleted cluster, run \"goki delete\" command wiht -v (--volume) flag.\n")
		}

		return nil
	},
}

func deleteGokiCluster() error {
	// Delete all containers.
	if err := deleteGokiContainers(); err != nil {
		return err
	}

	// Delete docker network.
	if err := deleteGokiNetwork(); err != nil {
		return err
	}

	// Delete docker volume, if --volume (-v) specified.
	if deleteCmdFlags.volume {
		if err := deleteGokiVolume(); err != nil {
			return err
		}
	}

	return nil
}

func deleteGokiContainers() error {
	// List of running containers.
	var liveGokiList []string = []string{}
	// List of stopped (killed) containers.
	var deadGokiList []string = []string{}

	// Get list of running container that related to Goki.
	c := exec.Command("docker", "ps", "-f", "label="+gokiResourceLabel, "--format", "{{.Names}}")

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "docker ps command failed: %v\n", string(output))
		return err
	} else if string(output) != "" {
		liveGokiList = strings.FieldsFunc(string(output), func(r rune) bool { return r == '\n' })
	}

	// First, kill all containers. If we use "docker stop", it take a long time.
	// So, use "docker kill" instead of "docker stop".
	if len(liveGokiList) != 0 {
		for _, container := range liveGokiList {
			c := exec.Command("docker", "kill", container)
			if output, err := c.CombinedOutput(); err != nil {
				fmt.Fprintf(os.Stderr, "docker kill command failed: %v\n", string(output))
				return err
			}
		}
	}

	// Get list of stopped containers to remove them.
	c = exec.Command("docker", "ps", "-af", "label="+gokiResourceLabel, "--format", "{{.Names}}")

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "docker ps command failed: %v\n", string(output))
		return err
	} else if string(output) != "" {
		deadGokiList = strings.FieldsFunc(string(output), func(r rune) bool { return r == '\n' })
	}

	// Remove all containers that related to Goki.
	if len(deadGokiList) != 0 {
		for _, container := range deadGokiList {
			c := exec.Command("docker", "rm", container)
			if output, err := c.CombinedOutput(); err != nil {
				fmt.Fprintf(os.Stderr, "docker rm command failed: %v\n", string(output))
				return err
			}
		}
	}

	// Show deleted container name.
	fmt.Println("  Docker Containers:")
	if len(deadGokiList) != 0 {
		for i := 0; i < len(deadGokiList); i++ {
			fmt.Println("    " + deadGokiList[i])
		}
	} else if len(deadGokiList) == 0 {
		fmt.Println("    Nothing")
	}

	return nil
}

func deleteGokiNetwork() error {
	var gokiNetworkName string

	// Get goki network name.
	c := exec.Command("docker", "network", "ls", "-f", "label="+gokiResourceLabel, "--format", "{{.Name}}")

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "docker network ls command failed: %v\n", string(output))
		return err
	} else if string(output) != "" {
		gokiNetworkName = string(output)
		gokiNetworkName = strings.TrimRight(gokiNetworkName, "\n")
	}

	// Remove docker network that related to Goki.
	if gokiNetworkName != "" {
		c := exec.Command("docker", "network", "rm", gokiNetworkName)
		if output, err := c.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "docker network rm command failed: %v\n", string(output))
			return err
		}
	}

	// Show deleted docker network name.
	fmt.Println("  Docker Network:")
	if gokiNetworkName != "" {
		fmt.Println("    " + gokiNetworkName)
	} else if gokiNetworkName == "" {
		fmt.Println("    Nothing")
	}

	return nil
}

func deleteGokiVolume() error {
	var gokiVolumeList []string = []string{}

	// Get list of goki volume.
	c := exec.Command("docker", "volume", "ls", "-f", "label="+gokiResourceLabel, "--format", "{{.Name}}")

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "docker volume ls command failed: %v\n", string(output))
		return err
	} else if string(output) != "" {
		gokiVolumeList = strings.FieldsFunc(string(output), func(r rune) bool { return r == '\n' })
	}

	// Remove all volumes that related to Goki.
	if len(gokiVolumeList) != 0 {
		for _, volume := range gokiVolumeList {
			c := exec.Command("docker", "volume", "rm", volume)
			if output, err := c.CombinedOutput(); err != nil {
				fmt.Fprintf(os.Stderr, "docker volume rm command failed: %v\n", string(output))
				return err
			}
		}
	}

	// Show deleted docker volume name.
	fmt.Println("  Docker Volumes:")
	if len(gokiVolumeList) != 0 {
		for i := 0; i < len(gokiVolumeList); i++ {
			fmt.Println("    " + gokiVolumeList[i])
		}
	} else if len(gokiVolumeList) == 0 {
		fmt.Println("    Nothing")
	}

	return nil
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	// Flags of goki delete.
	deleteCmd.Flags().BoolVarP(&deleteCmdFlags.volume, "volume", "v", false, "Delete docker volume")
}
