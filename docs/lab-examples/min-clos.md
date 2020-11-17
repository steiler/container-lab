|                               |                                                                        |
| ----------------------------- | ---------------------------------------------------------------------- |
| **Description**               | A minimal CLOS topology with two leafs and a spine                     |
| **Components**                | [Nokia SR Linux][srl]                                                  |
| **Resource requirements**[^1] | :fontawesome-solid-microchip: 1.5 <br/>:fontawesome-solid-memory: 3 GB |
| **Topology file**             | [clos01.yml][topofile]                                                 |
| **Name**                      | clos01                                                                 |

## Description
This labs provides a lightweight folded CLOS fabric topology using a minimal set of nodes: two leaves and a single spine.

<center><div class="mxgraph" style="max-width:100%;border:1px solid transparent;" data-mxgraph="{&quot;page&quot;:5,&quot;zoom&quot;:1.5,&quot;highlight&quot;:&quot;#0000ff&quot;,&quot;nav&quot;:true,&quot;check-visible-state&quot;:true,&quot;resize&quot;:true,&quot;url&quot;:&quot;https://raw.githubusercontent.com/srl-wim/containerlab-diagrams/main/containerlab.drawio&quot;}"></div></center>

The topology is additionally equipped with the Linux containers connected to leaves to facilitate use cases which require access side emulation.

## Use cases
With this lightweight CLOS topology a user can exhibit the following scenarios:

* perform configuration tasks applied to the 3-stage CLOS fabric
* demonstrate fabric behavior leveraging the user-emulating linux containers attached to the leaves

[srl]: https://www.nokia.com/networks/products/service-router-linux-NOS/
[topofile]: https://github.com/srl-wim/container-lab/tree/master/lab-examples/clos01/clos01.yml

[^1]: Resource requirements are provisional. Consult with SR Linux Software Installation guide for additional information.

<script type="text/javascript" src="https://cdn.jsdelivr.net/gh/hellt/drawio-js@main/embed2.js?&fetch=https%3A%2F%2Fraw.githubusercontent.com%2Fsrl-wim%2Fcontainerlab-diagrams%2Fmain%2Fcontainerlab.drawio" async></script>