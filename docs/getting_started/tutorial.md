# GLBC

The GLBCs main API is the Kubernetes Ingress object. GLBC watches for ingress objects and mutates them adding in the GLBC managed host and TLS certificate.

For more information on the architecture of GLBC and how the various component work, refer to the [overview documentation](https://github.com/Kuadrant/kcp-glbc/blob/bb8e43639691568b594720244a0c94a23470a587/docs/getting_started/overview.md).

Use this tutorial to perform the following actions:

* Install the `kcp-glbc` instance and verify installation.
* Create a new ingress resource
* Interact with the workload cluster
* Troubleshoot installation 


##  Procedure

### Prerequisites
- Clone this repository.
- Install [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl).
- Install Go 1.17 or higher as that is the version used in KCP-GLBC as per the [go.mod](https://github.com/Kuadrant/kcp-glbc/blob/main/go.mod) file.
- Have an AWS account, a DNS Zone, and a subdomain of the domain being used. This is optional as kind clusters can be used initially.

### Installation

Clone the repo and run:

```bash
make local-setup
```
Note: If errors are being encountered during the local-setup, please refer to the "Troubleshooting Installation" section at the end of this document.

This script will: 
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


Once the local-setup has completed running successfully, it will say that KCP is now running, and to run KCP-GLBC. For this, you will have two options:
1. [Local-deployment](https://github.com/Kuadrant/kcp-glbc/blob/main/docs/local_deployment.md): this option is good for testing purposes by using a local KCP instance and Kind clusters.

2. [Deploy latest in KCP](https://github.com/Kuadrant/kcp-glbc/blob/main/docs/deployment.md) with monitoring enabled: this will deploy GLBC to your target KCP instance. This will enable you to view observability in Prometheus and Grafana.


### Verify Installation
1. After running the local-setup successfully and ran one of the two options to have GLBC running, attempt to deploy the provided sample service in the terminal. Then, verify that the resources were created:
```bash
kubectl get deployment,service,ingress
```
2. [Verify Workload cluster and GLBC deployment](https://github.com/Kuadrant/kcp-glbc/blob/main/docs/deployment.md#:~:text=Verify%20the%20workload,1%20%20%20%20%20%20%20%20%20%20%20%201%20%20%20%20%20%20%20%20%20%20%2037m) are in a "ready" state.


### Troubleshooting Installation
While attempting to run make local-setup it’s possible you will encounter some of the following errors:
<br><br>

**make: *** No rule to make target 'local-setup':**
After cloning the repo, make sure to run the “make local-setup” command in the directory where the repo was cloned.<br><br>


**bash: line 1: go: command not found:**
We must install the correct go version used for this project. The version number can be found in the go.mod file of the repo. In this case, it is go 1.17.
If running on Fedora, here is a [guide to install go on Fedora 36](https://nextgentips.com/2022/05/21/how-to-install-go-1-18-on-fedora-36/). Before running the command to install go, make sure to type in the correct go version that is needed.<br><br>


**kubectl: command not found:**
Here is a quick and easy way of [installing kubectl on Fedora](https://snapcraft.io/install/kubectl/fedora).<br><br>


**Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?:**
Run the following command to start Docker daemon:
```bash
sudo systemctl start docker
```
<br><br>
**Kind cluster failed to become ready - Check logs for errors:**
Attempt the following to confirm if *kcp-cluster-1* and *kcp-cluster-2* are in a READY state:
```bash
KUBECONFIG=config/deploy/local/kcp.kubeconfig ./bin/kubectl-kcp workspace use root:default:kcp-glbc-user-compute
Current workspace is "root:default:kcp-glbc-user-compute".
```
```bash
kubectl get workloadclusters -o wide
NAME            LOCATION        READY   SYNCED API RESOURCES
kcp-cluster-1   kcp-cluster-1   True    
kcp-cluster-2   kcp-cluster-2   False 
```
If a cluster is not in READY state, the following procedure might solve the issue: [Configure Linux for Many Watch Folders](https://www.ibm.com/docs/en/ahte/4.0?topic=wf-configuring-linux-many-watch-folders) (we want to bump up each of the limits).


## Main Use Case - Providing ingress in a multi-cluster ingress scenario

This section will show how GLBC is used to provide ingress in a multi-cluster ingress scenario.

For this example we will be running the Sample Service which will create an ingress named *"ingress-nondomain"*. The "default" namespace is where we are putting all the sample stuff(resources) at the moment.

We can run `kubectl edit ns default -o yaml` to view and edit the "*default*" namespace.

There is a label named something like: "*state.internal.workload.kcp.dev/kcp-cluster-1: Sync*". GLBC is telling KCP where to sync all of the work resources in the namespace to. Meaning, since the namespace has *kcp-cluster-1* set on it, the ingress will have *kcp-cluster-1* set on it. We can edit that to *kcp-cluster-2*.

![Screenshot from 2022-07-26 11-46-27](https://user-images.githubusercontent.com/73656840/180992544-c21516fa-85a0-4c6a-9abc-efeb1a7c3433.png)

Then we can run the following command to view the ingress "*ingress-nondomain*": `kubectl get ingress ingress-nondomain -o yaml`. 

We can then observe that the label in the ingress has changed from *kcp-cluster-1* to *kcp-cluster-2*. KCP has propagated that label down from the namespace to everything in it. Everything in the namespace gets the same label. Because of that label, KCP is syncing it to *kcp-cluster-2*.

![Screenshot from 2022-07-26 11-47-30](https://user-images.githubusercontent.com/73656840/180992725-c6a4f985-da9f-4b68-bda7-ed3e61f43499.png)

Moreover, In the annotations we also have a status there for *kcp-cluster-2* and it has an IP address in it meaning that it has indeed synced to *kcp-cluster-2*. We will also find another annotation named something like "*kuadrant.dev/glbc-delete-at-kcp-cluster-1: 1658757564*", which is code coming from the GLBC which is saying "Don't remove this work from *kcp-cluster-1* until the DNS has propagated."

For that reason we can also observe another annotation named "*finalizers.workload.kcp.dev/kcp-cluster-1: kuadrant.dev/glbc-migration*" which remains there because GLBC is saying to KCP "Don't get rid of this yet, as we're waiting for it to come up to another cluster before you remove it from *kcp-cluster-1*" Once it has completely migrated, the finalizer for *kcp-cluster-1* will no longer be there.

![Screenshot from 2022-07-26 11-48-57](https://user-images.githubusercontent.com/73656840/180993006-78f47abc-d006-4045-95b7-33428cf65d6b.png)


### From the Client-Side
To continue on from the demo above, the following shows how the domain is viewed from the client's perspective. We will first start seeing a 404 error as the workload gets removed from *kcp-cluster-1* and migrates to *kcp-cluster-2*, the DNS record will update and in a few seconds we will start getting a response from *kcp-cluster-2*.

[comment]: <> (I will be adding more to this part...)






