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

	"github.com/spf13/cobra"
)

// Flag value of sql command.
var sqlCmdFlags struct {
	gokiId   int  // Number of node (container).
	nonRoot  bool // Version of CockroachDB that specified to tag of container image.
	userName string
	password string
}

// sqlCmd represents the sql command
var sqlCmd = &cobra.Command{
	Use:   "sql",
	Short: "Access the CockroachDB Local Cluster by built-in SQL shell",
	Long: `The "goki sql" command access to the created CockroachDB Local Cluster.
* By default, it accesses to Node 1 with root user.
    goki sql
* You can specify the Node ID with -g (--goki) flag.
    goki sql -g 3
* You can use default non-root user (User name : goki) with --non-root flag.
    goki sql --non-root
* You can use non-root user that you created by CREATE USER with -u (--user) and -p (--password) flag.
    goki sql -u foo -p foopass
* The "goki sql" command use CockroachDB's built-in SQL shell.
    * To exit the built-in SQL shell, use \q, quit, exit, or ctrl-d.
        * See official document for more details of CockroachDB's built-in SQL shell.
          https://www.cockroachlabs.com/docs/stable/cockroach-sql.html
`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if !sqlCmdFlags.nonRoot && sqlCmdFlags.userName == "" { // Access to DB as a root user.
			if err := accessGokiAsRoot(sqlCmdFlags.gokiId); err != nil {
				return err
			}
		} else if sqlCmdFlags.nonRoot && sqlCmdFlags.userName == "" { // Access to DB as a default non-root user.
			if err := accessGokiAsNonRoot(sqlCmdFlags.gokiId, gokiNonRootUserName, gokiNonRootUserPassword); err != nil {
				return err
			}
		} else if sqlCmdFlags.userName != "" { // Access to DB as a specified non-root user.
			if err := accessGokiAsNonRoot(sqlCmdFlags.gokiId, sqlCmdFlags.userName, sqlCmdFlags.password); err != nil {
				return err
			}
		}

		return nil
	},
}

func accessGokiAsRoot(id int) error {
	// Access to DB as a root user. Since I want to test certificate authentication method,
	// I don't use password authentication for root.
	c := exec.Command("docker", "exec", "-it", gokiResourceName+"-client",
		"./cockroach", "sql",
		"--certs-dir=/cockroach/certs/",
		"--host="+gokiResourceName+"-"+strconv.Itoa(id)+":26257",
	)

	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: \"cockroach sql\" command via to \"docker exec\" command failed.\n Error is: %v\n", err)
		return err
	}

	return nil
}

func accessGokiAsNonRoot(id int, user string, password string) error {
	// Access to DB as a non-root user.
	c := exec.Command("docker", "exec", "-it", gokiResourceName+"-client",
		"./cockroach", "sql",
		"--url",
		"postgresql://"+user+":"+password+"@"+gokiResourceName+"-"+strconv.Itoa(id)+":26257/defaultdb?sslmode=require",
	)

	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: \"cockroach sql\" command via to \"docker exec\" command failed.\n Error is: %v\n", err)
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(sqlCmd)
	// Flags of goki sql.
	sqlCmd.Flags().IntVarP(&sqlCmdFlags.gokiId, "goki", "g", 1, "The node ID of container that goki sql command will access.")
	sqlCmd.Flags().BoolVar(&sqlCmdFlags.nonRoot, "non-root", false, "Access as a default non-root user. Default user is \"goki\".")
	sqlCmd.Flags().StringVarP(&sqlCmdFlags.userName, "user", "u", "", "The user name to access DB.")
	sqlCmd.Flags().StringVarP(&sqlCmdFlags.password, "password", "p", "", "The password to access DB.")
}
