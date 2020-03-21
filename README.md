# memberlist-healthcheck-example

This repository contains a sample code for the [hashicorp/memberlist](https://github.com/hashicorp/memberlist)
# Usage
## Initiate a cluster on the first node
```
go run main.go init --bind-ip=x.x.x.x --http-port=8888
```
Where `x.x.x.x` is the IP that you want your local node to bind. And `8888` is the port of a simple web server to view health check results.

You should see something like this:

```
2020/03/20 18:21:22 new cluster created. key: 8X1/jqq2W2XpLvKOcv5vjPKx087X8xXmYcyJG8Ry/qQ=
2020/03/20 18:21:22 webserver is up. URL: http://192.168.21.101:8888/

```
## Join other nodes to the cluster
```
go run main.go join --bind-ip=x.x.x.x --http-port=8888 --cluster-key={CLUSTER_KEY} --known-ip=y.y.y.y
```
Both `--bind-ip` and `--http-port` flags are similar to the **init** command. `{CLUSTER_KEY}` is the key you receive on the first node. `--known-ip` is your gateway to the cluster; It can be any live node.

You should see something like this:

```
2020/03/20 16:55:11 [DEBUG] memberlist: Initiating push/pull sync with:  192.168.21.101:7946
2020/03/20 16:55:11 Joined the cluster URL
2020/03/20 16:55:11 webserver is up. URL: http://127.0.0.1:8889/ 
2020/03/20 16:55:16 [DEBUG] memberlist: Stream connection from=127.0.0.1:53954
2020/03/20 16:55:30 [DEBUG] memberlist: Initiating push/pull sync with: amir 192.168.21.101:7946
```
## View health check results
Now you can view health check using the port you provided in the command from any alive node in the cluster. For example `http://192.168.21.101:8888/`

You should see a json response like this:
```
[{"ip":"127.0.0.1:80","status":"UP"},{"ip":"192.168.21.101:80","status":"UP"}]
```