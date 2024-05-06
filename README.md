# Goki

Goki creates a CockroachDB Local Cluster in your local environment utilize Docker.

## Requirement

You must install Go and Docker on your local environment.

## How to install

### Download binary

You can get binary from [releases](https://github.com/kota2and3kan/goki/releases).

### Build from source code

```shell
git clone https://github.com/kota2and3kan/goki.git
cd goki
go install
```

## How to use

### Create a cluster

You can create a CockroachDB local cluster as follows. By default, it creates 3 node cluster.

```shell
goki create
```

If you want to create 5 node cluster, you can specify the number of node using `--node (-n)` flag.

```shell
goki create -n 5
```

You can specify the version of CockroachDB usint `--crdb-version` flag.

```shell
goki create -n 5 --crdb-version v22.1.11
```

### Set region and zone information to each node

You can set region and zone information for each node by specifying the `--set-locality (-l)` flag. Mainly, this is for the testing of Table Localities.

```shell
goki create --set-locality
```

If you set the `--set-locality (-l)` flag, goki adds region and zone information as follows:

```sql
root@goki-1:26257/defaultdb> SELECT node_id, address, locality FROM crdb_internal.gossip_nodes;
  node_id |   address    |          locality
----------+--------------+------------------------------
        1 | goki-1:26257 | region=region-1,zone=zone-1
        2 | goki-2:26257 | region=region-1,zone=zone-2
        3 | goki-3:26257 | region=region-1,zone=zone-3
        4 | goki-4:26257 | region=region-2,zone=zone-1
        5 | goki-5:26257 | region=region-2,zone=zone-2
        6 | goki-6:26257 | region=region-2,zone=zone-3
        7 | goki-7:26257 | region=region-3,zone=zone-1
        8 | goki-8:26257 | region=region-3,zone=zone-2
        9 | goki-9:26257 | region=region-3,zone=zone-3
(9 rows)
```

```console
+---[Region 1]-------------------------------------------+  +---[Region 2]-------------------------------------------+  +---[Region 3]-------------------------------------------+
|                                                        |  |                                                        |  |                                                        |
|  +---[Zone 1]---+  +---[Zone 2]---+  +---[Zone 3]---+  |  |  +---[Zone 1]---+  +---[Zone 2]---+  +---[Zone 3]---+  |  |  +---[Zone 1]---+  +---[Zone 2]---+  +---[Zone 3]---+  |
|  |              |  |              |  |              |  |  |  |              |  |              |  |              |  |  |  |              |  |              |  |              |  |
|  |  +--------+  |  |  +--------+  |  |  +--------+  |  |  |  |  +--------+  |  |  +--------+  |  |  +--------+  |  |  |  |  +--------+  |  |  +--------+  |  |  +--------+  |  |
|  |  | goki-1 |  |  |  | goki-2 |  |  |  | goki-3 |  |  |  |  |  | goki-4 |  |  |  | goki-5 |  |  |  | goki-6 |  |  |  |  |  | goki-7 |  |  |  | goki-8 |  |  |  | goki-9 |  |  |
|  |  +--------+  |  |  +--------+  |  |  +--------+  |  |  |  |  +--------+  |  |  +--------+  |  |  +--------+  |  |  |  |  +--------+  |  |  +--------+  |  |  +--------+  |  |
|  |              |  |              |  |              |  |  |  |              |  |              |  |              |  |  |  |              |  |              |  |              |  |
|  +--------------+  +--------------+  +--------------+  |  |  +--------------+  +--------------+  +--------------+  |  |  +--------------+  +--------------+  +--------------+  |
|                                                        |  |                                                        |  |                                                        |
+--------------------------------------------------------+  +--------------------------------------------------------+  +--------------------------------------------------------+
```

### Connect to the cluster using built-in SQL shell

After creating your CockroachDB local cluster, you can access to it using built-in SQL shell as follows. By default, it access to the first node `goki-1` as a `root` user.

```shell
goki sql
```

If you want to access other node, you can specify the node number using `--goki (-g)` flag.

```shell
goki sql -g 3
```

You can use default non-root user (User name is `goki`) using `--non-root` flag.

```shell
goki sql --non-root
```

If you create your own user, you can specify the user using `--user (-u)` flag and `--password (-p)` flag.

```shell
goki sql -u foo -p foopass
```

`goki sql` command uses the CockroachDB's built-in SQL shell. To exit buitl-in SQL shell, you can use `\q`, `quit`, `exit`, or `Ctrl-d`.

```shell
root@goki-1:26257/defaultdb> \q
```

### Delete the cluster

You can delete the CockroachDB local cluster as follows. By default, it deletes docker containers and docker network only. The docker volumes that include CockroachDB's data are not deleted.

```shell
goki delete
```

In this case, you can restart your CockroachDB local cluster using the existing data in the docker volume by `goki create` command.

If you want to all component includes docker volume, you can specify the `--volume (-v)` flag as follows.

```shell
goki delete -v
```

## License

Please refer to the [LICENSE](https://github.com/kota2and3kan/goki/blob/main/LICENSE) for the details on the license of the files in this repository.
