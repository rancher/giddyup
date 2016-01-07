# Giddyup

##  Purpose

Giddyup is a tool to that helps get services started in a Rancher compose stack. It aims to simplify entrypoint and command scripting to start your Docker services. This is a first pass at addressing common tasks when starting up applications in Docker containers on Rancher.

Current capabilities:
 * Get connection strings from DNS or Rancher Metadata.
 * Determine if your container is the leader in the service.
 * Proxy traffic to the leader
 * Wait for service to have the desired scale.
 * Get the scale of the service


## Examples

### Determine if running container is the leader

```
#!/bin/bash
...
giddyup leader check
if [ "$?" -eq "0" ];then
   echo "I'm the leader"
fi
...
```

### Wait for service scale
This is useful if you need to generate configurations based on all nodes in a service. Otherwise, each container will only get itself and the previous containers metadata.

```
#!/bin/bash
...
# wait upto 2 minutes for all containers to come up
./giddyup service wait scale --timeout 120
...
# Bring up the service
```

### Get a Zookeeper connection string

```
#!/bin/bash
...
connection_string=$(./giddyup ip stringify --suffix :2181 zookeeper/zookeeper)
# Results in something like:
# 10.42.231.55:2181,10.42.145.91:2181,10.42.55.78:2181
...
```

## Usage

### ip stringify
```
NAME:
   giddyup ip stringify - Prints a joined list of IPs

USAGE:
   giddyup ip stringify [command options] [arguments...]

OPTIONS:
   --delimiter ","	Delimiter to use between entries
   --prefix 		Prepend Entries with this value
   --suffix 		Add this value to the end of each entry.
   --source "metadata"	Source to lookup IPs. [metadata, dns]

```

### leader

```
NAME:
   giddyup leader - Determines if this container has lowest start index

USAGE:
   giddyup leader command [command options] [arguments...]

COMMANDS:
   check	Check if we are leader and exit.
   elect	Simple leader election with Rancher
   forward	Listen and forward all port traffic to leader.
   get		Get the leader of service
   help, h	Shows a list of commands or help for one command

OPTIONS:
   --help, -h	show help
```
`elect` and `forward` are to be used as entrypoints.

If `elect` is used, only the container determined by Rancher (lowest create_index) will run the service. A command must be given to the `elect` command to execute if the container becomes the leader. Otherwise all traffic is forwarded to the leader, and upon election the container will exit.

`forward` should be used in its own container. It works in situations where you want your service running all the time, for replication or something, but want all of the traffic to go to a specific (leader) host. 

a forward example:
`giddyup leader forward --src-port 3307 --dst-port 3306`

This will listen on port 3307 and forward to port 3306 on the leader. This allows you to put a service behind a load balancer, and still have traffic go to one place. 

### Service

```
NAME:
   giddyup service - Service actions

USAGE:
   giddyup service command [command options] [arguments...]

COMMANDS:
   wait		Wait for service states
   scale	Get the set scale of the service
   containers	lists containers in the calling container's service one per line
   help, h	Shows a list of commands or help for one command

OPTIONS:
   --help, -h	show help
```

scale will give you the set scale of the service, and giddyup service scale --current will give you the current number of containers running in your service.

### Simple Health Check
```
NAME:
   ./bin/giddyup health - simple healthcheck

USAGE:
   ./bin/giddyup health [command options] [arguments...]

OPTIONS:
   --listen-port, -p "1620"	set port to listen on
```

This check just listens on the port specified (default: 1620) and responds to requests at `http://<ip>:<port>/ping` and responds with 200 OK. Its meant to be run in a sidekick as the entrypoint. It should share the network namespace as your application.
   