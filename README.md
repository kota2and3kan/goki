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
