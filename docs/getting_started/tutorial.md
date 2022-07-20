# GLBC

The GLBCs main API is the Kubernetes Ingress object. GLBC watches for ingress objects and mutates them adding in the GLBC managed host and TLS certificate.

For more information on the architecture of GLBC and how the various component work, refer to the overview documentation

Use this tutorial to perform the following actions:

* Install the `kcp-glbc` instance and verify installation.
* Create a new ingress resource
* Interact with the workload cluster
* Troubleshoot installation 


## Installation

Install to 
* build all the binaries
* Deploy three Kubernetes `1.22` clusters locally using kind.
* Deploy and configure the ingress controllers in each cluster.
* Start the KCP server.
* Create KCP workspaces for glbc and user resources:
    * kcp-glbc
    * kcp-glbc-compute
    * kcp-glbc-user
    * kcp-glbc-user-compute
* Add workload clusters to the `*-compute` workspaces
    * kcp-glbc-compute: 1x kind cluster
    * kcp-glbc-user-compute: 2x kind clusters
* Deploy glbc dependencies (cert-manager) into kcp-glbc workspace.



### Prerequisites

###  Procedure

For more information refer to ...

### Verify Installation


### Troubleshooting Installation



## Providing ingress in a multi-cluster ingress scenario

A guide on how to use the GLBC to provide ingress in a multi-cluster ingress scenario (kind clusters are fine initially)


## Common Features 

### Testing

- Common features such as testing to see workloads moving after a workspace is unavailable






