Containerlab labs are typically deployed in the isolated environments, such as company's internal network, cloud instance or even a laptop. The nodes deployed in a lab can happily talk to each other and, if needed, can reach Internet in the outbound direction.

But sometimes it is really needed to let your lab nodes be reachable over Internet securely and privately in the incoming direction. There are many use cases that warrant such _publishing_, some of the most notable are:

* create a lab in your environment and share it with a customer/colleague on-demand in no time
* make an interactive demo/training where certain nodes' are shared with an audience for hand-on experience
* share a private lab with someone to collaborate
* expose management interfaces (gNMI, NETCONF, SNMP) to test integration with collectors deployed outside of your lab environment

Containerlab made all of these use cases possible by integrating with [mysocket.io](https://mysocket.io) service. Mysocket.io provides personal tunnels for https/https/tls/tcp ports over global anycast[^1] network spanning US, Europe and Asia.

To make a certain port of a certain node available via mysocket.io tunnel a single line in the topology definition file is all that's needed:

```yaml
name: demo
topology:
  nodes:
    r1:
      kind: srl
      publish:
        - tcp/22     # tcp port 22 will be published
        - tcp/57400  # tcp port 57400 will be published
        - http/10200 # http service running over 10200 will be published
```

<!-- <video width="100%" controls>
  <source src="https://gitlab.com/rdodin/pics/-/wikis/uploads/709405ded4ccf7387725b4fab1ab87f6/containerlab-mysocketio.mp4" type="video/mp4">
</video> -->

## Registration
Tunnels set up by mysocket.io are associated with a user who set them, thus users are required to register within the service. Luckily, the [registration](https://mysocket.readthedocs.io/en/latest/mysocketctl/mysocket.html#creating-an-account) is trivial, all you need to provide is an email and a public SSH key that will be used to set up tunnels.

For convenience, containerlab comes with a script to create mysocket.io account in one go:

```bash
# create mysocket.io account
# docker needs to be installed
# usage: mysocket-user.sh <email> <password> <abs path to public ssh key>
/etc/containerlab/tools/mysocket-user myemail@gmail.com mypassword /root/.ssh/mykey.pub
```

A confirmation email will arrive shortly to finish account setup procedure.

## Acquiring a token
To authenticate with mysocket.io service a user needs to acquire/refresh the token. Containerlab users can leverage another convenience script that eases this step:

```bash
# get/refresh mysocketio token
# the script will save the token under the $(pwd)/mysocket_token file
# usage: mysocket-token.sh <email> <password>
/etc/containerlab/tools/mysocket-token.sh myemail@gmail.com mypassword
```

The script will get the token and save it in the current directory under `mysocket_token` name.

!!!info
    The token is valid for 5 hours, once the token expires, the already established tunnels will continue to work, but to establish new tunnels a new token must be provided.

## Specify what to share
To indicate which ports to publish the users need to add `publish` section under node/kind or default level of the [topology definition file](topo-def-file.md). In the example below, we decide to publish SSH and gNMI services of `r1` node:

```yaml
name: demo
topology:
  nodes:
    r1:
      kind: srl
      publish:
        - tcp/22     # tcp port 22 will be exposed
        - tcp/57400  # tcp port 57400 will be exposed
```

The `publish` section holds a list of `<type>/<port-number>` strings, where `type` must be one of the supported mysocket.io socket type[^2] - http/https/tls/tcp. Every type/port combination will be exposed via its own private tunnel.

!!!note
    For a single account the following maximum number of tunnels is set:  
      * tcp based tunnels - 5  
      * http based tunnels - 10  
    If >5 tcp tunnels is required users should launch a VM in a lab, expose it's SSH service and use this VM as a jumpbox for other TCP services.

## Add mysocketio node
Containerlab integrates with mysocket.io service by leveraging it's client application packaged in a container format. In order for the sockets indicated in the `publish` block to be exposed, a user needs to add a node of `mysocketio` kind to the topology. Augmenting the topology we used above, the full topology file will look like:

```yaml
name: demo
topology:
  nodes:
    r1:
      kind: srl
      publish:
        - tcp/22     # tcp port 22 will be exposed
        - tcp/57400  # tcp port 57400 will be exposed

    # adding mysocketio linux container
    mysocketio:
      kind: mysocketio
      image: ghcr.io/hellt/mysocketctl:0.1.0
      binds:
        - ~/.ssh/privkey:/root/.ssh/id_rsa         # bind mount your private key
        - mysocketio_token:/root/.mysocketio_token # bind mount API token
```

The `mysocketio` node is a simple linux container with mysocketctl client installed. Containerlab uses this node to create the sockets and start tunnels as defined in the `publish` block.

Pay specific attention to `binds` defined for mysocketio node. With this section we provide the two crucial artifacts:
* path to the private key, that matches the public key used during the registration
* path to the API token that we acquired before launching the lab

And that is all that is needed to expose the sockets in an automated way.

## Explore published ports
When a user launches a lab with published ports it will be presented with a summary table after the lab deployment process finishes:

```
+---+-----------------------+--------------+---------------------------------+------------+-------+---------+----------------+----------------------+
| # |         Name          | Container ID |              Image              |    Kind    | Group |  State  |  IPv4 Address  |     IPv6 Address     |
+---+-----------------------+--------------+---------------------------------+------------+-------+---------+----------------+----------------------+
| 1 | clab-sock-r1          | 9cefd6cdb239 | srlinux:20.6.3-145              | srl        |       | running | 172.20.20.2/24 | 2001:172:20:20::2/80 |
| 2 | clab-sock-mysocketctl | 8f5385beb97e | ghcr.io/hellt/mysocketctl:0.1.0 | mysocketio |       | running | 172.20.20.3/24 | 2001:172:20:20::3/80 |
+---+-----------------------+--------------+---------------------------------+------------+-------+---------+----------------+----------------------+
Published ports:
┌──────────────────────────────────────┬──────────────────────────────────────┬─────────┬──────┬────────────┬────────────────────────┐
│ SOCKET ID                            │ DNS NAME                             │ PORT(S) │ TYPE │ CLOUD AUTH │ NAME                   │
├──────────────────────────────────────┼──────────────────────────────────────┼─────────┼──────┼────────────┼────────────────────────┤
│ 2b2808ae-b05d-4103-9b95-ab473a82658c │ nameless-bird-8969.edge.mysocket.io  │ 16086   │ tcp  │ false      │ clab-tcp-22-r1         │
│ 8455571c-deea-4b09-bc1d-7a56f41e8c52 │ restless-night-8051.edge.mysocket.io │ 11107   │ tcp  │ false      │ clab-tcp-57400-r1      │
└──────────────────────────────────────┴──────────────────────────────────────┴─────────┴──────┴────────────┴────────────────────────┘
```
The **Published ports** table lists the ports and their corresponding DNS names. Looking at the NAME column users can quickly discover which tunnel corresponds to which node and its published port. The socket name follows the `clab-<type>-<port>-<node-name>` pattern.

To access the published port, users need to combine the DNS name and the Port to derive the full address. For the exposed SSH port, for example, the ssh client can use the following command to access remote SSH service:

```
ssh user@nameless-bird-8969.edge.mysocket.io -p 16086
```

## Troubleshooting
To check the health status of the established tunnels execute the following command to check the logs created on mysocketio container:

```
docker exec -it <mysocketio-node-name> /bin/sh -c "cat socket*"
```

This command will display all the logs for the published ports. If something is not right, you will see the erros in the log.

[^1]: https://mysocket.readthedocs.io/en/latest/about/about.html#build-on-a-global-anycast-network
[^2]: https://mysocket.readthedocs.io/en/latest/about/about.html#features