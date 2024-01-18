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
	"strings"
)

const (
	// Related to CockroachDB
	crdbContainerImage string = "cockroachdb/cockroach" // CockroachDB's image name in the Docker Hub.
	crdbVersion        string = "v22.2.0"               // CockroachDB's version (Tag of container image).
	// Related to Goki
	gokiVersion             string = "Development version (latest main branch)"
	gokiResourceName        string = "goki"      // Prefix of each resource (e.g. goki-client, goki-network, goki-volume etc...)
	gokiSqlIp               string = "127.0.0.1" // IP that the first container will listen for SQL connection.
	gokiSqlPort             string = "26257"     // Port that the first contarner will listen for SQL connection.
	gokiWebUiIp             string = "127.0.0.1" // IP that the first container will listen for HTTP request (Web UI).
	gokiWebUiPort           string = "8081"      // Port that the first container will listen for HTTP request (Web UI).
	gokiNonRootUserName     string = "goki"      // Name of Non-root user.
	gokiNonRootUserPassword string = "goki"      // Password of Non-root user.
	gokiRootUserPassword    string = "gokiroot"  // Password of Root user.
	gokiResourceLabel       string = "goki"      // Label that will be specifed each docker resources.
)

func gokiIsDead(id int) (bool, error) {
	// List of running containers.
	var liveGokiList []string = []string{}
	// List of stopped (killed) containers.
	var deadGokiList []string = []string{}

	// Get list of running container that related to Goki.
	c := exec.Command("docker", "ps", "-f", "label="+gokiResourceLabel, "--format", "{{.Names}}")

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "docker ps command failed: %v\n", string(output))
		return false, err
	} else if string(output) != "" {
		liveGokiList = strings.FieldsFunc(string(output), func(r rune) bool { return r == '\n' })
	}

	// Check the specified container is running.
	if gokiContains(liveGokiList, gokiResourceName+"-"+strconv.Itoa(id)) {
		// The specified container is running.
		return false, nil
	}

	// Get list of stopped containers that related to Goki.
	c = exec.Command("docker", "ps", "-af", "label="+gokiResourceLabel, "--format", "{{.Names}}")

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "docker ps command failed: %v\n", string(output))
		return false, err
	} else if string(output) != "" {
		deadGokiList = strings.FieldsFunc(string(output), func(r rune) bool { return r == '\n' })
	}

	// Check the specified container is dead (existing).
	if gokiContains(deadGokiList, gokiResourceName+"-"+strconv.Itoa(id)) {
		// The specified container is dead.
		return true, nil
	} else {
		return false, errors.New("The specified container " + gokiResourceName + "-" + strconv.Itoa(id) + " does not exist")
	}
}
