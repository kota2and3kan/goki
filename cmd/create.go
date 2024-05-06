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
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

var (
	gokiVolumeAlreadyExist bool     // If Goki's docker volume already exist, re-use it and skip cluster initializing.
	gokiRegionId           int  = 0 // The ID of a region that a node will be deployed. Goki uses region-1, region-2, or region-3.
	gokiZoneId             int  = 0 // The ID of a zone that a node will be deployed. Goki uses zone-1, zone-2, or zone-3.
)

// Flag value of create command.
var createCmdFlags struct {
	node        int    // Number of node (container).
	crdbVersion string // Version of CockroachDB that specified to tag of container image.
	locality    bool   // Whether set --locality flag or not.
}

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create the CockroachDB Local Cluster",
	Long: `The "goki create" command creates the CockroachDB Cluster in your local environment utilize Docker.
* By default, it creates 3 node cluster.
    goki create
* You can specify the number of node between 1 and 9 with -n (--node) flag.
    goki create -n 9
* You can specify the version of CockroachDB with --crdb-version flag.
    goki create --crdb-version v21.2.7
`,
	RunE: func(cmd *cobra.Command, args []string) error {

		// Check the docker command exist or not.
		if err := checkDocker(); err != nil {
			return err
		}

		// Check the already running Goki Container.
		if err := checkGokiContainer(); err != nil {
			return err
		}

		// Check the already exist Goki Network.
		if err := checkGokiNetwork(); err != nil {
			return err
		}

		// Check the already exist Goki Volume.
		if err := checkGokiVolume(); err != nil {
			return err
		}

		// Check the value of --node flag (number of cockroaches).
		if err := checkNumOfNode(); err != nil {
			return err
		}

		// Create CockroachDB Local Cluster.
		fmt.Println("INFO: The number of cockroaches in the cluster is", createCmdFlags.node, ".")
		if createCmdFlags.locality {
			fmt.Println("INFO: The --set-locality is true. Set region and zone information in each node.")
		}
		fmt.Println("INFO: *** Start Creating CockroachDB Local Cluster ***")

		// Create Docker Network that each cockroach and client will join.
		if err := createGokiNetwork(); err != nil {
			return err
		}

		// Create Docker Volume for each cockroach.
		if err := createGokiVolume(); err != nil {
			return err
		}

		// Create client container as a client of CockroachDB Cluster.
		// And, this container will be used to create cert files.
		if err := createClientContainer(); err != nil {
			return err
		}

		// Create cert files for secure cluster.
		if err := createCertFile(); err != nil {
			return err
		}

		// Create first node.
		if err := createFirstGoki(); err != nil {
			return err
		}

		// Create Cluster (Run containers).
		if err := createGokiCluster(); err != nil {
			return err
		}

		// Check the CockroachDB is ready to accept SQL connections.
		if err := checkSqlConnectionAcceptance(); err != nil {
			return err
		}

		// Confrim and show the Cluster Status by "cockroach node status" command.
		if err := confirmClusterStatus(); err != nil {
			return err
		}

		// If the Cluster already initialized, we don't need to load intro DB.
		if !gokiVolumeAlreadyExist {
			// Load intro DB.
			if err := loadIntroDB(); err != nil {
				return err
			}
		}

		// Show Cockroach AA of intro DB.
		if err := showCockroach(); err != nil {
			return err
		}

		// Show the way to access DB and Web UI.
		showAccessCommand()

		return nil
	},
}

func checkDocker() error {
	c := exec.Command("docker", "version")

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Goki needs Docker. Please install Docker.")
		fmt.Fprintf(os.Stderr, "HINT: docker version command failed: %v\n", string(output))
		return err
	}
	return nil
}

func checkGokiContainer() error {
	c := exec.Command("docker", "ps", "-af", "label="+gokiResourceLabel, "--format", "{{.ID}} : {{.Names}}")

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "docker ps commad failed: %v\n", string(output))
		return err
	} else if len(output) != 0 { // docker ps command succeeded. But, some container exists that its label is goki.
		fmt.Println(string(output))
		fmt.Fprintln(os.Stderr, "ERROR: Maybe the Goki Cluster is already running.")
		fmt.Fprintln(os.Stderr, "HINT: The following Goki Container exists.")
		fmt.Fprintf(os.Stderr, "%s\n", string(output))
		return errors.New("maybe the Goki Cluster is already running")
	}
	return nil
}

func checkGokiNetwork() error {
	c := exec.Command("docker", "network", "ls", "-f", "label="+gokiResourceLabel, "--format", "{{.ID}} : {{.Name}}")

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "docker network ls commad failed: %v\n", string(output))
		return err
	} else if len(output) != 0 { // docker network command succeeded. But, some network exists that its label is goki.
		fmt.Fprintln(os.Stderr, "ERROR: Maybe the Goki Cluster is already running.")
		fmt.Fprintln(os.Stderr, "HINT: The following Goki Network exists.")
		fmt.Fprintf(os.Stderr, "%s\n", string(output))
		return errors.New("maybe the Goki Cluster is already running")
	}
	return nil
}

func checkGokiVolume() error {
	c := exec.Command("docker", "volume", "ls", "-f", "label="+gokiResourceLabel, "--format", "{{.Name}}")

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "docker volume ls command failed: %v\n", string(output))
		return err
	} else if string(output) != "" {
		gokiVolumeAlreadyExist = true
	} else if string(output) == "" {
		gokiVolumeAlreadyExist = false
	}
	return nil
}

func checkNumOfNode() error {
	if createCmdFlags.node < 1 || 9 < createCmdFlags.node {
		fmt.Fprintln(os.Stderr, "ERROR: Invalid argument. Please specify the number of cockroaches between 1 and 9.")
		fmt.Fprintln(os.Stderr, "HINT: The max number of cockroaches of Goki is 9.")
		return errors.New("invalid argument. Please specify the number of cockroaches between 1 and 9")
	}
	return nil
}

func createGokiNetwork() error {
	fmt.Println("INFO: Creating Docker Network " + gokiResourceName + "-net start.")

	c := exec.Command("docker", "network", "create", "-d", "bridge", gokiResourceName+"-net",
		"-o", "com.docker.network.bridge.name="+gokiResourceName+"-net",
		"--label="+gokiResourceLabel)

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: docker network create command failed.\n Error is: %v\n", string(output))
		return err
	} else {
		fmt.Printf("INFO: Created docker network is: %s", string(output))
		fmt.Println("INFO: Creating Docker Network " + gokiResourceName + "-net done.")
	}
	return nil
}

func createGokiVolume() error {
	fmt.Println("INFO: Creating Docker Volume start.")

	// For client container. This volume includes cert files.
	c := exec.Command("docker", "volume", "create", gokiResourceName+"-volume-client",
		"--label="+gokiResourceLabel)

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: docker volume create command failed.\n Error is: %v\n", string(output))
		return err
	} else {
		fmt.Printf("INFO: Created docker volume is: %s", string(output))
	}

	// For each cockroach. This volume includes data file of DB.
	for i := 1; i <= createCmdFlags.node; i++ {
		c := exec.Command("docker", "volume", "create", gokiResourceName+"-volume-"+strconv.Itoa(i),
			"--label="+gokiResourceLabel)

		if output, err := c.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: docker volume create command failed.\n Error is: %v\n", string(output))
			return err
		} else {
			fmt.Printf("INFO: Created docker volume is: %s", string(output))
		}
	}

	fmt.Println("INFO: Creating Docker Volume done.")
	return nil
}

func createClientContainer() error {
	fmt.Println("INFO: Creating client container start.")

	c := exec.Command("docker", "run", "-d",
		"--name="+gokiResourceName+"-client",
		"--hostname="+gokiResourceName+"-client",
		"--network="+gokiResourceName+"-net",
		"--mount=type=volume,src="+gokiResourceName+"-volume-client,dst=/cockroach/certs",
		"--label="+gokiResourceLabel,
		"--entrypoint=sleep",
		crdbContainerImage+":"+createCmdFlags.crdbVersion,
		"inf",
	)

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: docker run command that creating "+gokiResourceName+"-client failed.\n Error is: %v\n", string(output))
		return err
	} else {
		fmt.Printf("INFO: Created client container is: %s", string(output))
	}

	fmt.Println("INFO: Creating client container done.")
	return nil
}

func createCertFile() error {
	fmt.Println("INFO: Creating cert files start.")

	// Create CA cert dir.
	c := exec.Command("docker", "exec", gokiResourceName+"-client",
		"mkdir", "-p", "/cockroach/certs/node-certs/../.setup/my-safe-directory/")
	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: docker exec command that creating certs dir failed.\n Error is: %v\n", string(output))
		return err
	}

	// Create cockroach (node) certs dir.
	for i := 1; i <= createCmdFlags.node; i++ {
		c = exec.Command("docker", "exec", gokiResourceName+"-client",
			"mkdir", "-p", "/cockroach/certs/node-certs/"+gokiResourceName+"-"+strconv.Itoa(i))

		if output, err := c.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: docker exec command that creating certs dir failed.\n Error is: %v\n", string(output))
			return err
		}
	}

	// Create CA.
	c = exec.Command("docker", "exec", gokiResourceName+"-client",
		"./cockroach", "cert", "create-ca",
		"--certs-dir=/cockroach/certs/.setup/cert-tmp",
		"--ca-key=/cockroach/certs/.setup/my-safe-directory/ca.key",
		"--allow-ca-key-reuse",
		"--overwrite",
	)

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: docker exec command that creating CA cert files failed.\n Error is: %v\n", string(output))
		return err
	}

	// Create cockroach (node) certs.
	for i := 1; i <= createCmdFlags.node; i++ {
		c := exec.Command("docker", "exec", gokiResourceName+"-client",
			"./cockroach", "cert", "create-node", gokiResourceName+"-"+strconv.Itoa(i), "localhost",
			"--certs-dir=/cockroach/certs/.setup/cert-tmp",
			"--ca-key=/cockroach/certs/.setup/my-safe-directory/ca.key",
			"--overwrite",
		)

		if output, err := c.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: docker exec command that creating node cert files failed.\n Error is: %v\n", string(output))
			return err
		}

		c = exec.Command("docker", "exec", gokiResourceName+"-client",
			"cp",
			"/cockroach/certs/.setup/cert-tmp/ca.crt",
			"/cockroach/certs/.setup/cert-tmp/node.crt",
			"/cockroach/certs/.setup/cert-tmp/node.key",
			"/cockroach/certs/node-certs/"+gokiResourceName+"-"+strconv.Itoa(i)+"/",
		)

		if output, err := c.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: docker exec command that copy node cert files failed.\n Error is: %v\n", string(output))
			return err
		}
	}

	// Create client cert.
	c = exec.Command("docker", "exec", gokiResourceName+"-client",
		"./cockroach", "cert", "create-client", "root",
		"--certs-dir=/cockroach/certs/.setup/cert-tmp",
		"--ca-key=/cockroach/certs/.setup/my-safe-directory/ca.key",
		"--overwrite",
	)

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: docker exec command that create client cert files failed.\n Error is: %v\n", string(output))
		return err
	}

	c = exec.Command("docker", "exec", gokiResourceName+"-client",
		"cp",
		"/cockroach/certs/.setup/cert-tmp/ca.crt",
		"/cockroach/certs/.setup/cert-tmp/client.root.crt",
		"/cockroach/certs/.setup/cert-tmp/client.root.key",
		"/cockroach/certs/",
	)

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: docker exec command that copy client cert files failed.\n Error is: %v\n", string(output))
		return err
	}

	fmt.Println("INFO: Creating cert files done.")
	return nil
}

func createFirstGoki() error {
	// Create first node.
	fmt.Println("INFO: Creating First node start.")

	if createCmdFlags.locality {
		gokiRegionId = 1
		gokiZoneId = 1
	}

	c := exec.Command("docker", "run", "-d",
		"--name="+gokiResourceName+"-1",
		"--hostname="+gokiResourceName+"-1",
		"--network="+gokiResourceName+"-net",
		"-p", gokiSqlIp+":"+gokiSqlPort+":26257",
		"-p", gokiWebUiIp+":"+gokiWebUiPort+":8080",
		"--mount=type=volume,src="+gokiResourceName+"-volume-client,dst=/cockroach/certs",
		"--mount=type=volume,src="+gokiResourceName+"-volume-1,dst=/cockroach/cockroach-data",
		"--label="+gokiResourceLabel,
		crdbContainerImage+":"+createCmdFlags.crdbVersion,
		"start",
		"--certs-dir=certs/node-certs/"+gokiResourceName+"-1",
		"--join="+gokiResourceName+"-1,"+gokiResourceName+"-2,"+gokiResourceName+"-3",
		"--locality=region=region-"+strconv.Itoa(gokiRegionId)+",zone=zone-"+strconv.Itoa(gokiZoneId),
	)

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Start first node failed.\n Error is: %v\n", string(output))
		return err
	} else {
		fmt.Printf("INFO: Created container is: %s", string(output))
	}

	// Wait for 1 second just in case, for waiting first node start before init cluster.
	time.Sleep(time.Second * 1)
	fmt.Println("INFO: Creating First node done.")

	// Init CockroachDB Cluster.
	// For sorting internal IDs of each goki in ascending order, initialize cluster after first node run.
	// And, after that, run the second and later node.
	// If the "gokiResourceName-volume-1" already exsit when goki has started,
	// it is assumed to be the cluster (DB) data is already initialized.
	// We don't need to initialize cluster, if it is already initialized.
	if !gokiVolumeAlreadyExist {
		fmt.Println("INFO: Initializing Cluster start.")

		c = exec.Command("docker", "exec", gokiResourceName+"-client",
			"./cockroach", "init",
			"--certs-dir=certs/",
			"--host="+gokiResourceName+"-1:26257",
		)

		if output, err := c.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Start first node failed.\n Error is: %v\n", string(output))
			return err
		}

		fmt.Println("INFO: Initializing Cluster done.")
	} else {
		fmt.Println("INFO: Skip initializing cluster, because the Cluster is already initialized.")
	}

	// If the Cluster already initialized, we don't need to setup user.
	// Also, we don't need to check the node ID.
	if !gokiVolumeAlreadyExist {
		// Set root user's password.
		if err := setRootPassword(); err != nil {
			return err
		}

		// Create non-root user.
		if err := createNonRootUser(); err != nil {
			return err
		}

		// Check container name suffix and internal ID are same or not.
		if err := checkGokiNode(1); err != nil {
			return err
		}
	}

	fmt.Println("INFO: Creating First node done.")
	return nil
}

func setRootPassword() error {
	c := exec.Command("docker", "exec", gokiResourceName+"-client",
		"./cockroach", "sql",
		"--certs-dir=/cockroach/certs/",
		"--host="+gokiResourceName+"-1:26257",
		"-e", "ALTER USER root WITH PASSWORD '"+gokiRootUserPassword+"'",
	)

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: docker exec command that set root user's password failed.\n Error is: %v\n", string(output))
		return err
	}
	return nil
}

func createNonRootUser() error {
	c := exec.Command("docker", "exec", gokiResourceName+"-client",
		"./cockroach", "sql",
		"--certs-dir=/cockroach/certs/",
		"--host="+gokiResourceName+"-1:26257",
		"-e", "CREATE USER IF NOT EXISTS "+gokiNonRootUserName+" WITH PASSWORD '"+gokiNonRootUserPassword+"'",
	)

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: docker exec command that create non-root user failed.\n Error is: %v\n", string(output))
		return err
	}
	return nil
}

func checkGokiNode(g int) error {
	// Connect to CockroachDB via PostgreSQL driver.
	connStr := "postgresql://root:" + gokiRootUserPassword + "@localhost:26257/defaultdb?sslmode=require"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Connecting to CockroachDB via PostgreSQL driver failed.\n")
		return err
	}
	// Get internal ID and node name (address) from cluster meta data.
	rows, err := db.Query("SELECT node_id, address FROM crdb_internal.gossip_nodes WHERE node_id = $1", g)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Getting internal ID and node name (address) from cluster meta data failed.\n")
		return err
	}
	defer rows.Close()
	// Check the number of container name suffix and internal ID are same or not.
	for rows.Next() {
		var id int
		var address string
		if err := rows.Scan(&id, &address); err != nil {
			fmt.Fprintf(os.Stderr, "Scanning rows failed.\n")
			return err
		}
		// If it does not match, exit with error.
		if id != g && address != gokiResourceName+"-"+strconv.Itoa(g)+":26257" {
			fmt.Fprintf(os.Stderr, "ERROR: Node number and internal ID does not match.\n Node number is: %v\n Internal ID is: %v\n Node name is: %v\n", g, id, address)
			return errors.New("node number and internal ID does not match")
		}
	}
	// Container name suffix and internal ID are same.
	return nil
}

func createGokiCluster() error {
	fmt.Println("INFO: Creating Cluster start.")

	// Run the second and later node.
	for i := 2; i <= createCmdFlags.node; i++ {

		if createCmdFlags.locality {
			switch i {
			case 1, 2, 3:
				gokiRegionId = 1
			case 4, 5, 6:
				gokiRegionId = 2
			case 7, 8, 9:
				gokiRegionId = 3
			}

			if i%3 == 0 {
				gokiZoneId = 3
			} else {
				gokiZoneId = i % 3
			}
		}

		c := exec.Command("docker", "run", "-d",
			"--name="+gokiResourceName+"-"+strconv.Itoa(i),
			"--hostname="+gokiResourceName+"-"+strconv.Itoa(i),
			"--network="+gokiResourceName+"-net",
			"--mount=type=volume,src="+gokiResourceName+"-volume-client,dst=/cockroach/certs",
			"--mount=type=volume,src="+gokiResourceName+"-volume-"+strconv.Itoa(i)+",dst=/cockroach/cockroach-data",
			"--label="+gokiResourceLabel,
			crdbContainerImage+":"+createCmdFlags.crdbVersion,
			"start",
			"--certs-dir=certs/node-certs/"+gokiResourceName+"-"+strconv.Itoa(i),
			"--join="+gokiResourceName+"-1,"+gokiResourceName+"-2,"+gokiResourceName+"-3",
			"--locality=region=region-"+strconv.Itoa(gokiRegionId)+",zone=zone-"+strconv.Itoa(gokiZoneId),
		)

		if output, err := c.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Start second or later node failed.\n Error is: %v\n", string(output))
			return err
		} else {
			fmt.Printf("INFO: Created container is: %s", string(output))
		}

		// Wait for 1 second just in case, for waiting each node start before start next node.
		time.Sleep(time.Second * 1)

		// If the Cluster already initialized, we don't need to check the node ID.
		if !gokiVolumeAlreadyExist {
			// Check container name suffix and internal ID are same or not.
			if err := checkGokiNode(i); err != nil {
				return err
			}
		}
	}
	fmt.Println("INFO: Creating Cluster done.")
	return nil
}

func checkSqlConnectionAcceptance() error {
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second * 1)

		c := exec.Command("docker", "exec", gokiResourceName+"-client",
			"./cockroach", "sql",
			"--certs-dir=/cockroach/certs/",
			"--host="+gokiResourceName+"-1:26257",
			"-e", "SELECT 1",
		)

		if output, err := c.CombinedOutput(); err != nil {
			if i < 9 {
				fmt.Println("INFO: CockroachDB is NOT ready to accept connections.")
			} else {
				fmt.Fprintf(os.Stderr, "ERROR: CockroachDB was NOT ready to accept connections, even if try to connect to DB 10 times (even if waiting about 10 second).\n Error is: %v\n", string(output))
				fmt.Println("HINT: There is possibility that some error occurred. Please check the DB or Container log.")
				return err
			}
		} else {
			fmt.Println("INFO: CockroachDB is ready to accept connections.")
			break
		}
	}
	return nil
}

func confirmClusterStatus() error {
	fmt.Println("INFO: CockroachDB Cluster Status is the following.")

	c := exec.Command("docker", "exec", "-t", gokiResourceName+"-client",
		"./cockroach", "node", "status",
		"--certs-dir=/cockroach/certs/",
		"--host="+gokiResourceName+"-1:26257",
	)

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cockroach node status command failed.\n Error is: %v\n", string(output))
		return err
	} else {
		fmt.Println(string(output))
	}

	return nil
}

func loadIntroDB() error {
	c := exec.Command("docker", "exec", "-t", gokiResourceName+"-client",
		"./cockroach", "workload", "init", "intro",
		"postgresql://root@"+gokiResourceName+"-1:26257?sslcert=certs%2Fclient.root.crt&sslkey=certs%2Fclient.root.key&sslmode=verify-full&sslrootcert=certs%2Fca.crt",
	)

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cockroach workload init intro command failed.\n Error is: %v\n", string(output))
		return err
	}

	return nil
}

func showCockroach() error {
	c := exec.Command("docker", "exec", "-t", gokiResourceName+"-client",
		"./cockroach", "sql",
		"--certs-dir=/cockroach/certs/",
		"--host="+gokiResourceName+"-1:26257",
		"-e", "SELECT v as \"Hello, CockroachDB!\" FROM intro.mytable WHERE (l % 2) = 0",
	)

	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: docker exec command that show intro DB failed.\n Error is: %v\n", string(output))
		return err
	} else {
		fmt.Println(string(output))
	}

	return nil
}

func showAccessCommand() {
	fmt.Println("*** Creating CockroachDB Local Cluster done ***")
	fmt.Println("INFO: Let's access to the CockroachDB by using built-in SQL Shell, and Web UI!")

	fmt.Printf("\nAccess DB as a root user:\n")
	fmt.Printf("  goki sql\n")

	fmt.Printf("\nAccess DB as a non-root user:\n")
	fmt.Printf("  goki sql --non-root\n")
	fmt.Printf("    or\n")
	fmt.Printf("  goki sql -u <user name> -p <password>\n")

	fmt.Printf("\nAccess Web UI as a root user (User: root / Password: %v):\n", gokiRootUserPassword)
	fmt.Printf("  URL: https://%v:%v/\n", gokiWebUiIp, gokiWebUiPort)

	fmt.Printf("\nAccess Web UI as a non-root user (User: %v / Password: %v):\n", gokiNonRootUserName, gokiNonRootUserPassword)
	fmt.Printf("  URL: https://%v:%v/\n\n", gokiWebUiIp, gokiWebUiPort)
}

func init() {
	rootCmd.AddCommand(createCmd)
	// Flags of goki create.
	createCmd.Flags().IntVarP(&createCmdFlags.node, "node", "n", 3, "The number of cockroaches.")
	createCmd.Flags().StringVar(&createCmdFlags.crdbVersion, "crdb-version", crdbVersion, "Version of CockroachDB (Tag of container image).")
	createCmd.Flags().BoolVarP(&createCmdFlags.locality, "set-locality", "l", false, "Set --locality flag (region and zone value) to all nodes.")
}
