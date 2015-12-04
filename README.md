# Giddyup

##  Purpose

Giddyup is a tool to that helps get services started in a Rancher compose stack. It aims to simplify entrypoint and command scripting to start your Docker services. This is a first pass at addressing common tasks when starting up applications in Docker containers.

Current capabilities:
 * Get connection strings from DNS or Rancher Metadata.
 * Determine if your container is the leader in the service.
 * Wait for service to have the desired scale.


## Examples

### Determine if running container is the leader

```
#!/bin/bash
...
giddyup leader
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
   ./giddyup leader - Determines if this container has lowest start index

USAGE:
   ./giddyup leader [arguments...]
```

### Service Wait

```
NAME:
   giddyup service wait - Wait for service states

USAGE:
   giddyup service wait command [command options] [arguments...]

COMMANDS:
   scale	Wait for number of service containers to reach set scale
   help, h	Shows a list of commands or help for one command

OPTIONS:
   --help, -h	show help
```