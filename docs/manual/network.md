<script type="text/javascript" src="https://cdn.jsdelivr.net/gh/hellt/drawio-js@main/embed2.js?&fetch=https%3A%2F%2Fraw.githubusercontent.com%2Fsrl-wim%2Fcontainerlab-diagrams%2Fmain%2Fcontainerlab.drawio" async></script>
One of the most important tasks in the process of building container based labs is to create a virtual wiring between the containers and the host. That is one of the problems that containerlab was designed to solve.

In this document we will discuss the networking concepts that containerlab employs to provide the following connectivity scenarios:

1. Make containers available from the lab host
2. Interconnect containers to create network topologies of users choice

## Management network
As governed by the well-established container [networking principles](https://docs.docker.com/network/) containers are able to get network connectivity using various drivers/methods. The most common networking driver that is enabled by default for docker-managed containers is the [bridge driver](https://docs.docker.com/network/bridge/).

The bridge driver connects containers to a linux bridge interface named `docker0` on most linux operating systems. The containers are then able to communicate with each other and the host via this virtual switch (bridge interface).

In containerlab we follow a similar approach: containers launched by containerlab will be attached with their interface to a containerlab-managed docker network. It's best to be explained by an example which we will base on a [two nodes](../lab-examples/two-srls.md) lab from our catalog:

```yaml
name: srl02

topology:
  kinds:
    srl:
      type: ixr6
      image: srlinux
      license: license.key
  nodes:
    srl1:
      kind: srl
    srl2:
      kind: srl

  links:
    - endpoints: ["srl1:e1-1", "srl2:e1-1"]
```

As seen from the topology definition file, the lab consists of the two SR Linux nodes which are interconnected via a single point-to-point link.

<div class="mxgraph" style="max-width:100%;border:1px solid transparent;margin:0 auto; display:block;" data-mxgraph="{&quot;page&quot;:3,&quot;zoom&quot;:1.5,&quot;highlight&quot;:&quot;#0000ff&quot;,&quot;nav&quot;:true,&quot;check-visible-state&quot;:true,&quot;resize&quot;:true,&quot;url&quot;:&quot;https://raw.githubusercontent.com/srl-wim/containerlab-diagrams/main/containerlab.drawio&quot;}"></div>

The diagram above shows that these two nodes are not only interconnected between themselves, but also connected to a bridge interface on the lab host. This is driven by the containerlab default management network settings.

### default settings
When no information about the management network is provided within the topo definition file, containerlab will do the following

1. create, if not already created, a docker network named `clab`
2. configure the IPv4/6 addressing pertaining to this docker network

!!!info
    We often refer to `clab` docker network simply as _management network_ since its the network to which management interfaces of the containerized NOS'es are connected.

The addressing information that containerlab will use on this network:

* IPv4: subnet 172.20.20.0/24, gateway 172.20.20.1
* IPv6: subnet 2001:172:20:20::/80, gateway 2001:172:20:20::1

With these defaults in place, the two containers from this lab will get connected to that management network and will be able to communicate using the IP addresses allocated by docker daemon. The addresses that docker carves out for each container are presented to a user once the lab deployment finishes or can be queried any time after:

```bash
# addressing information is available once the lab deployment completes
❯ containerlab deploy -t srl02.yml
# deployment log omitted for brevity
+---+-----------------+--------------+---------+------+-------+---------+----------------+----------------------+
| # |      Name       | Container ID |  Image  | Kind | Group |  State  |  IPv4 Address  |     IPv6 Address     |
+---+-----------------+--------------+---------+------+-------+---------+----------------+----------------------+
| 1 | clab-srl02-srl1 | ca24bf3d23f7 | srlinux | srl  |       | running | 172.20.20.3/24 | 2001:172:20:20::3/80 |
| 2 | clab-srl02-srl2 | ee585eac9e65 | srlinux | srl  |       | running | 172.20.20.2/24 | 2001:172:20:20::2/80 |
+---+-----------------+--------------+---------+------+-------+---------+----------------+----------------------+

# addresses can also be fetched afterwards with `inspect` command
❯ containerlab inspect -a
+---+----------+-----------------+--------------+---------+------+-------+---------+----------------+----------------------+
| # | Lab Name |      Name       | Container ID |  Image  | Kind | Group |  State  |  IPv4 Address  |     IPv6 Address     |
+---+----------+-----------------+--------------+---------+------+-------+---------+----------------+----------------------+
| 1 | srl02    | clab-srl02-srl1 | ca24bf3d23f7 | srlinux | srl  |       | running | 172.20.20.3/24 | 2001:172:20:20::3/80 |
| 2 | srl02    | clab-srl02-srl2 | ee585eac9e65 | srlinux | srl  |       | running | 172.20.20.2/24 | 2001:172:20:20::2/80 |
+---+----------+-----------------+--------------+---------+------+-------+---------+----------------+----------------------+
```

The output above shows that srl1 container has been assigned `172.20.20.3/24 / 2001:172:20:20::3/80` IPv4/6 address. We can ensure this by querying the srl1 management interfaces address info:

```bash
❯ docker exec clab-srl02-srl1 ip address show dummy-mgmt0
6: dummy-mgmt0: <BROADCAST,NOARP> mtu 1500 qdisc noop state DOWN group default qlen 1000
    link/ether 2a:66:2b:09:2e:4d brd ff:ff:ff:ff:ff:ff
    inet 172.20.20.3/24 brd 172.20.20.255 scope global dummy-mgmt0
       valid_lft forever preferred_lft forever
    inet6 2001:172:20:20::3/80 scope global
       valid_lft forever preferred_lft forever
```

Now it's possible to reach the assigned IP address from the lab host as well as from other containers connected to this management network.
```bash
# ping srl1 management interface from srl2
❯ docker exec -it clab-srl02-srl2 sr_cli "ping 172.20.20.3 network-instance mgmt"
x -> $176x48
Using network instance mgmt
PING 172.20.20.3 (172.20.20.3) 56(84) bytes of data.
64 bytes from 172.20.20.3: icmp_seq=1 ttl=64 time=2.43 ms
```

!!!note
    If you run multiple labs without changing the default management settings, the containers of those labs will end up connecting to the same management network with their management interface.

### configuring management network
Most of the time there is no need to change the defaults for management network configuration, but sometimes it is needed. For example, it might be that the default network ranges are overlapping with existing addressing scheme on the lab host.

For such cases the users need to add the `mgmt` container at the top level of their topology definition file:

```yaml
name: srl02

mgmt:
  network: custom_mgmt                # management network name
  ipv4_subnet: 172.100.100.0/24       # ipv4 range
  ipv6_subnet: 2001:172:100:100::/80  # ipv6 range (optional)

topology:
# the rest of the file is omitted for brevity
```

With this settings in place container will get their IP addresses from the specified ranges accordingly.

### connection details
When containerlab needs to create the management network it asks the docker daemon to do this. Docker will fullfil the request and will create a network with the underlying linux bridge interface backing it. The bridge interface name is generated by the docker daemon, but it is easy to find it:

```bash
# list existing docker networks
# notice the presence of the `clab` network with a `bridge` driver
❯ docker network ls
NETWORK ID          NAME                DRIVER              SCOPE
5d60b6ec8420        bridge              bridge              local
d2169a14e334        clab                bridge              local
58ec5037122a        host                host                local
4c1491a09a1a        none                null                local

# the underlying linux bridge interface name follows the `br-<first_12_chars_of_docker_network_id> pattern
# to find the network ID use:
❯ docker network inspect clab -f {{.ID}} | head -c 12
d2169a14e334

# now the name is known and its easy to show bridge state
❯ brctl show br-d2169a14e334
bridge name	        bridge id		    STP enabled	  interfaces
br-d2169a14e334		8000.0242fe382b74	no		      vetha57b950
							                          vethe9da10a
```

As explained in the beginning of this article, containers will connect to this docker network. This connection is carried out by the `veth` devices created and attached with one end to bridge interface in the lab host and the other end in the container namespace. This is illustrated by the bridge output above and the diagram at the beginning the of the article.

## Point-to-point links
Management network is used to provide management access to the NOS containers, it does not carry control or dataplane traffic. In containerlab we create additional point-to-point links between the containers to provide the datapath between the lab nodes.

<div class="mxgraph" style="max-width:100%;border:1px solid transparent;margin:0 auto; display:block;" data-mxgraph="{&quot;page&quot;:11,&quot;zoom&quot;:1.5,&quot;highlight&quot;:&quot;#0000ff&quot;,&quot;nav&quot;:true,&quot;check-visible-state&quot;:true,&quot;resize&quot;:true,&quot;url&quot;:&quot;https://raw.githubusercontent.com/srl-wim/containerlab-diagrams/main/containerlab.drawio&quot;}"></div>

The above diagram shows how links are created in the topology definition file. In this example, the datapath consists of the two virtual point-to-point wires between SR Linux and cEOS containers. These links are created on-demand by containerlab itself.

The p2p links are provided by the `veth` device pairs where each end of the `veth` pair is attached to a respective container.