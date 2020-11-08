|                               |                                                                        |
| ----------------------------- | ---------------------------------------------------------------------- |
| **Description**               | a single Nokia SR Linux node                                           |
| **Components**                | [Nokia SR Linux][srl]                                                  |
| **Resource requirements**[^1] | :fontawesome-solid-microchip: 0.5 <br/>:fontawesome-solid-memory: 1 GB |
| **Topology file**             | [srl01.yml][topofile]                                                  |
| **Prefix**                    | srl01                                                                  |

## Description
A lab consists of a single SR Linux container equipped with a single interface - its management interface. No other network/data interfaces are created.

<center><div class="mxgraph" style="max-width:100%;border:1px solid transparent;" data-mxgraph="{&quot;page&quot;:2,&quot;zoom&quot;:1.5,&quot;highlight&quot;:&quot;#0000ff&quot;,&quot;nav&quot;:true,&quot;check-visible-state&quot;:true,&quot;resize&quot;:true,&quot;url&quot;:&quot;https://raw.githubusercontent.com/srl-wim/containerlab-diagrams/main/containerlab.drawio&quot;}"></div></center>

The SR Linux's `mgmt` interface is connected to the `containerlab` docker network that is created as part of the lab deployment process. The `mgmt` interface of SRL will get IPv4/6 address information via DHCP service provided by docker daemon.

## Use cases
This lightweight lab enables the users to perform the following exercises:

* get familiar with SR Linux architecture
* explore SR Linux extensible CLI
* navigate the SR Linux YANG tree
* play with gNMI[^2] and JSON-RPC programmable interfaces
* write/debug/manage custom apps built for SR Linux NDK

[srl]: https://www.nokia.com/networks/products/service-router-linux-NOS/
[topofile]: https://github.com/srl-wim/container-lab/tree/master/lab-examples/srl01/srl01.yml

[^1]: Resource requirements are provisional. Consult with SR Linux Software Installation guide for additional information.
[^2]: Check out [gnmic](https://gnmic.kmrd.dev) gNMI client to interact with SR Linux gNMI server.

<script type="text/javascript" src="https://cdn.jsdelivr.net/gh/hellt/drawio-js@main/embed2.js?&fetch=https%3A%2F%2Fraw.githubusercontent.com%2Fsrl-wim%2Fcontainerlab-diagrams%2Fmain%2Fcontainerlab.drawio" async></script>