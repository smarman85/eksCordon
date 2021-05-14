# eksCordon
A helper script for managing resources in EKS.

![eksCordon Status](https://github.com/smarman85/eksCordon/workflows/eksCordon/badge.svg)


## Usage:
eksCordon uses the Kube config associated with the person running the command. You will need to change contexts to the cluster you
wish to interact with.

```bash
$ eksCordon
Helper script to cordon and drain a troublesome availability zone

Usage:
  eksCordon [command]

Available Commands:
  cordonAZ    Cordon an AZ (only one)
  help        Help about any command
  listAZs     List AZs

Flags:
  -h, --help   help for eksCordon

Use "eksCordon [command] --help" for more information about a command.
```

### Listing Availability Zones:
```bash
$ eksCordon listAZs
AZ: us-east-1a  Number of nodes: 16
AZ: us-east-1b  Number of nodes: 17
AZ: us-east-1c  Number of nodes: 17
AZ: us-east-1d  Number of nodes: 13
```

### Cordon and Drain nodes in an availability zone
By default, this will:
  * scale cluster autoscaler to 0 replicas
  * find all the nodes in the specified zone
  * evict all pods hosted in the specified Availability Zone (kills pod immediately)

```bash
$ eksCordon cordonAZ -z us-east-1d
```

### Testing:
```
go test -v ./cmd/...
zsh: correct './cmd/...' to './cmd/..' [nyae]? n
=== RUN   TestDisplayAvailabilityZones
=== RUN   TestDisplayAvailabilityZones/Prints_AZs_and_number_of_nodes_in_an_az
--- PASS: TestDisplayAvailabilityZones (0.00s)
    --- PASS: TestDisplayAvailabilityZones/Prints_AZs_and_number_of_nodes_in_an_az (0.00s)
=== RUN   TestGetFailureDomains
=== RUN   TestGetFailureDomains/#00
--- PASS: TestGetFailureDomains (0.00s)
    --- PASS: TestGetFailureDomains/#00 (0.00s)
=== RUN   TestGetNodesInAZ
=== RUN   TestGetNodesInAZ/#00
--- PASS: TestGetNodesInAZ (0.00s)
    --- PASS: TestGetNodesInAZ/#00 (0.00s)
=== RUN   TestCordonNodes
--- PASS: TestCordonNodes (0.00s)
=== RUN   TestPodsOnNode
--- PASS: TestPodsOnNode (0.00s)
=== RUN   TestEvictPods
Container:  test-pod Namespace:  test-space
--- PASS: TestEvictPods (0.00s)
PASS
ok      eksCordon/cmd     0.623s
```


### TODOs:
- [ ] Wire up testing scripts
- [x] Parallel node cordoning and pod eviction
- [x] Write tests
- [ ] Pin packages
- [ ] ensure daemonsets aren't removed from the node

