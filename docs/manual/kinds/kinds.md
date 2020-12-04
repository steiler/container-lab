# Kinds

Containerlab launches, wires up and manages container based labs. The steps required to launch a `debian` or `centos` container image aren't at all different. On the other hand, Nokia SR Linux launching procedure is nothing like the one for Arista cEOS.

Things like required syscalls, mounted directories, entrypoints and commands to execute are all different for the containerized NOS'es. To let containerlab to understand which launching sequence to use the notion of a `kind` was introduced. Essentially we remove the need to understand certain setup peculiarities of different NOS'es by abstracting them with `kinds`.

Given the following [topology definition file](topo-def-file.md), containerlab is able to know how to launch `node1` as SR Linux container and `node2` as a cEOS one because they are associated with the kinds:

```yaml
name: srlceos01

topology:
  nodes:
    node1:
      kind: srl              # node1 is of srl kind
      type: ixrd2
      image: srlinux
      license: license.key
    node2:
      kind: ceos             # node2 is of ceos kind
      image: ceos            

  links:
    - endpoints: ["srl:e1-1", "ceos:eth1"]
```

Containerlab supports a fixed number of kinds. Within each predefined kind we store the necessary information that is used to launch the container successfully. The following kinds are supported or in the roadmap of containerlab:


| Name                | Kind            | Status    |
| ------------------- | --------------- | --------- |
| **Nokia SR Linux**  | [`srl`](srl.md) | supported |
| **Arista cEOS**     | `ceos`          | supported |
| **Linux container** | `linux`         | supported |
| **Linux bridge**    | `bridge`        | supported |
| **SONiC**           | `sonic`         | planned   |
| **Juniper cRPD**    | `crpd`          | planned   |

Refer to a specific kind documentation article to see the details about it.