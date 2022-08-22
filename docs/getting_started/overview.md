# What is `kcp`?

`kcp` is a prototype of a multi-tentant Kubernetes control plane for workloads on many clusters. `kcp` can be used to manage Kubernetes-like applications across one or more clusters and integrate them with cloud services. 

It provides a generic CRD apiserver that is divided into multiple logical clusters (in which each of the logical clusters are fully isolated) that enable multitenancy of cluster-scoped resources such as CRDs and Namespaces. 

See the [`kcp` docs](https://github.com/Kuadrant/kcp) for further explanation. To learn more about the terminology, refer to the [docs](https://github.com/kcp-dev/kcp/blob/main/docs/terminology.md).


# What is GLBC?

The KCP Global Load Balancer Controller (GLBC) solves multi-cluster ingress use cases while leveraging KCP to provide transparent multi-cluster deployments.

Currently, the GLBC is deployed in a Kubernetes cluster, referred as the GLBC control cluster, outside the KCP control plane. The GLBC dependencies, such as cert-manager, and eventually external-dns, are deployed alongside the GLBC in that control cluster.

These components coordinate via a shared state, that's persisted in the control cluster data plane.

The following benefits are envisioned:

The main use case it solves currently is providing you with a single host that can be used to access your workload and bring traffic to the correct physical clusters. The GLBC manages the DNS for this host and provides you with a valid TLS certificate. If your workload moves/is moved or expands contracts across clusters, GLBC will ensure that the DNS for this host is correct and traffic will continue to reach your workload.

Leverage the data durability guarantees, provided by hosted KCP environments;
Compute commoditization, and workload movement.

For more information on GLBC, refer to 

# Architecture

Refer to the architecture 

# Terms to Know

