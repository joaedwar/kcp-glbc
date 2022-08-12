# GLBC

The Global Load Balancer Controller (GLBC) leverages [KCP](https://github.com/Kuadrant/kcp) to provide DNS-based global load balancing and transparent multi-cluster ingress. The main API for the GLBC is the Kubernetes Ingress object. GLBC watches Ingress objects and transforms them adding in the GLBC managed host and TLS certificate.

For more information on the architecture of GLBC and how the various component work, refer to the [overview documentation](https://github.com/Kuadrant/kcp-glbc/blob/bb8e43639691568b594720244a0c94a23470a587/docs/getting_started/overview.md).

Use this tutorial to perform the following actions:

* Install the KCP-GLBC instance and verify installation.
* Follow the demo and have GLBC running and working with an AWS domain. You can then deploy the sample service to view how GLBC allows access to services  in a multi-cluster ingress scenario.

---

## Prerequisites
- Install [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl).
- Install Go `1.17` or higher. This is the version used in KCP-GLBC as indicated in the [`go.mod`](https://github.com/Kuadrant/kcp-glbc/blob/main/go.mod) file.
- Install the [yq](https://snapcraft.io/install/yq/fedora) command-line YAML processor.
- Have an AWS account, a DNS Zone, and a subdomain of the domain being used. You will need this in order to instruct GLBC to make use of your AWS credentials and configuration.


## Installation

Clone the repo and run the following command:

```bash
make local-setup
```
> NOTE: If errors are encountered during the local-setup, refer to the Troubleshooting Installation document.

This script performs the following actions: 
* Builds all the binaries
* Deploys three Kubernetes `1.22` clusters locally using kind
* Deploys and configures the ingress controllers in each cluster
* Downloads KCP 0.6.0
* Starts the KCP server
* Creates KCP workspaces for GLBC and user resources:
    * kcp-glbc
    * kcp-glbc-compute
    * kcp-glbc-user
    * kcp-glbc-user-compute
* Add workload clusters to the `*-compute` workspaces
    * kcp-glbc-compute: 1x kind cluster
    * kcp-glbc-user-compute: 2x kind clusters
* Deploy GLBC dependencies (cert-manager) into the kcp-glbc workspace.

-----

After `local-setup` has successfully completed, it will say that KCP is now running. However, at this point, GLBC is not yet running. You will be presented in the terminal with two options to deploy GLBC:

1. [Local-deployment](https://github.com/Kuadrant/kcp-glbc/blob/main/docs/local_deployment.md): this option is good for testing purposes by using a local KCP instance and kind clusters.

2. [Deploy latest in KCP](https://github.com/Kuadrant/kcp-glbc/blob/main/docs/deployment.md) with monitoring enabled: this will deploy GLBC to your target KCP instance. This will enable you to view observability in Prometheus and Grafana.

For the demo, before deploying GLBC, we will want to provide it with your AWS credentials and configuration.


### Provide GLBC with AWS credentials and configuration
The easiest way to do this is to perform the following steps:

1. Open the KCP-GLBC project in your IDE.
1. Navigate to the `./config/deploy/local/aws-credentials.env` environment file.
1. Enter your `AWS access key ID` and `AWS Secret Access Key` as indicated in the example below:

   ![Screenshot from 2022-07-28 12-33-50](https://user-images.githubusercontent.com/73656840/181609265-8577f9c0-1d32-4e1f-8cf2-7542a340393b.png)
   
1. Navigate to `./config/deploy/local/controller-config.env` and change the following fields to something similar to this:

   ![Screenshot from 2022-07-28 12-43-56](https://user-images.githubusercontent.com/73656840/181609374-b0d2c81f-0d46-4816-b53e-05514fa382c2.png)

      The fields that might need to be edited include:
       - Replace `<AWS_DNS_PUBLIC_ZONE_ID>` with your own hosted zone ID.
       - Replace `<GLBC_DNS_PROVIDER>` with `aws`.
       - Replace `<GLBC_DOMAIN>` with your specified subdomain

### Run GLBC
After all the above is set up correctly, for the demo, now we can run the first 3 commands under Option 1 to have GLBC running. The commands are similar to the following (run them in a new tab):

```bash
Run Option 1 (Local):
       cd to/the/repo
       export KUBECONFIG=config/deploy/local/kcp.kubeconfig
       ./bin/kubectl-kcp workspace use root:default:kcp-glbc
```
We need to export KUBECONFIG to ensure that information about our clusters are passed to child processes, then we will be able to change our workspace. We should get an output saying: `Current workspace is "root:default:kcp-glbc"`


Then, in the same tab in the terminal, run the following command to make use of your "controller-config.env" and "aws-credentials.env". This way, we will be able to curl the domain in the tutorial and visualize how the workload from cluster-1 migrates to cluster-2.
```bash
(export $(cat ./config/deploy/local/controller-config.env | xargs) && export $(cat ./config/deploy/local/aws-credentials.env | xargs) && ./bin/kcp-glbc --kubeconfig .kcp/admin.kubeconfig)
```
<br>

### Deploy the sample service

Now we will attempt to deploy the sample service. You can do this by running the following command in a new tab in the terminal:
```bash
./samples/location-api/sample.sh
```
After the sample service has been deployed, we are presented with the following output of what was done, and some useful commands:

![Screenshot from 2022-08-02 12-22-17](https://user-images.githubusercontent.com/73656840/182363020-6aa61b73-c2a7-4570-ada7-aae97ad9db00.png)


The sample script will remain paused until we press the enter key to migrate the workload from one cluster to the other. However, we will not perform this action just yet.

<br>

## Verify sample service deployment
1. In a new terminal, verify that the ingress was created after deploying the sample service:
```bash
export KUBECONFIG=.kcp/admin.kubeconfig                                         
./bin/kubectl-kcp workspace use root:default:kcp-glbc-user
kubectl get ingress
```

2. Verify that the DNS record was created:
```bash
export KUBECONFIG=.kcp/admin.kubeconfig                                         
./bin/kubectl-kcp workspace use root:default:kcp-glbc-user
kubectl get dnsrecords ingress-nondomain -o yaml
```
We might not get an output just yet until the DNS record exists in route53 (This may take a couple of minutes).

<br>

You could also view in your AWS domain if the DNS record was created.

![Screenshot from 2022-08-02 12-26-19](https://user-images.githubusercontent.com/73656840/182363808-558f8a40-4ed6-4e08-9c02-1d74b6209b46.png)




<br>

## Main Use Case - Demo on providing ingress in a multi-cluster ingress scenario

This section will show how GLBC is used to provide ingress in a multi-cluster ingress scenario.

<b>Note: This version of the tutorial works with KCP 0.6.0.</b>

For this tutorial, after following along the "Installation" section of this document, we should already have KCP and GLBC running, and also have had deployed the sample service which would have created a placement resource, an ingress named *"ingress-nondomain"* and a DNS record. To note: the "default" namespace is where we are putting all the sample resources at the moment.

<br>

### Viewing the "default" namespace

We will run the following commands in a new tab:

```bash
export KUBECONFIG=.kcp/admin.kubeconfig                                         
./bin/kubectl-kcp workspace use root:default:kcp-glbc-user
kubectl get ns default -o yaml
```
As we can see, there is a label named: "*state.internal.workload.kcp.dev/kcp-cluster-1: Sync*":

![Screenshot from 2022-08-02 12-32-06](https://user-images.githubusercontent.com/73656840/182365628-22f04bb5-0818-46a3-8a12-3abc2e8451f3.png)

GLBC is telling KCP where to sync all of the work resources in the namespace to. Meaning, since the namespace has *kcp-cluster-1* set on it, the ingress will also have *kcp-cluster-1* set on it. 

<br>

### Watching the ingress and the DNS record

We can run the watch command in a new tab to start watching the ingress:

```bash
export KUBECONFIG=.kcp/admin.kubeconfig                                         
./bin/kubectl-kcp workspace use root:default:kcp-glbc-user
watch -n1 -d 'kubectl get ingress ingress-nondomain -o yaml | yq eval ".metadata" - | grep -v "kubernetes.io"'
```

As we can see in the first annotation, the load balancer for *kcp-cluster-1* will have an IP address (once the DNS record is created):

![Screenshot from 2022-08-02 12-40-48](https://user-images.githubusercontent.com/73656840/182366116-aa2f32ce-a603-49bb-b974-e9356c71c6fc.png)

<br>

We can also run the following command in another tab to start watching the DNS record in real-time:

```bash
export KUBECONFIG=.kcp/admin.kubeconfig                                         
./bin/kubectl-kcp workspace use root:default:kcp-glbc-user
watch -n1 'kubectl get dnsrecords ingress-nondomain -o yaml | yq eval ".spec" -'
```


<br>

### Curl the running domain

Now that the DNS record was created successfully, in a new tab in the terminal, we can curl the domain to view it running. To do this, we will run the following watch command that is outputted in our terminal, it will look similar to this:

```bash
watch -n1 "curl -k https://cbkgg75kjgmah1mbpvsg.cz.hcpapps.net"
```

This will curl the domain which will give an output similar to the following:

![Screenshot from 2022-08-02 12-44-15](https://user-images.githubusercontent.com/73656840/182368772-8a08a197-66d9-4d9c-9747-74ddaad0e4d7.png)

This would mean that *kcp-cluster-1* is up and running correctly.

<br>

### Migrating workload from *kcp-cluster-1* to *kcp-cluster-2*

As we continue with the following steps, we will want to be observing the tab where we are watching our domain. This way, we will notice that during the workload migration, there is no interruptions and no down time.

To proceed with the workload migration, we will go to the tab where we deployed the sample service, and press the enter key to "trigger migration from kcp-cluster-1 to kcp-cluster-2". This deletes "placement-1" and creates "placement-2" which is pointing at *kcp-cluster-2*. This will also change the label in the "default" namespace mentioned before: "*state.internal.workload.kcp.dev/kcp-cluster-1: Sync*", and change it from *kcp-cluster-1* to *kcp-cluster-2*.

![Screenshot from 2022-08-02 12-48-05](https://user-images.githubusercontent.com/73656840/182367670-dc6c243d-aea7-44e9-bebf-99685391d931.png)


In the tab where we are watching the ingress, we can observe that the label in the ingress has changed from *kcp-cluster-1* to *kcp-cluster-2*. KCP has propagated that label down from the namespace to everything in it. Everything in the namespace gets the same label. Because of that label, KCP is syncing it to *kcp-cluster-2*.

![Screenshot from 2022-08-02 12-51-37](https://user-images.githubusercontent.com/73656840/182367915-5de8acef-4c77-4c09-a1b9-049c2605ce12.png)

<br><br>

Moreover, In the annotations we also have a status there for *kcp-cluster-2* and it has an IP address in it meaning that it has indeed synced to *kcp-cluster-2*. We will also find another annotation named "*deletion.internal.workload.kcp.dev/kcp-cluster-1*", which is code coming from the GLBC which is saying "Don't remove this work from *kcp-cluster-1* until the DNS has propagated."

For that reason we can also observe another annotation named "*finalizers.workload.kcp.dev/kcp-cluster-1: kuadrant.dev/glbc-migration*" which remains there because GLBC is saying to KCP "Don't get rid of this yet, as we're waiting for it to come up to another cluster before you remove it from *kcp-cluster-1*" Once it has completely migrated, the finalizer for *kcp-cluster-1* will no longer be there and the workload will be deleted from *kcp-cluster-1*.

![Screenshot from 2022-08-02 12-49-21](https://user-images.githubusercontent.com/73656840/182368360-0bb65282-1751-44ea-a9da-7cfbe508e084.png)

<br><br>


We will notice that in the tab where we are curling the domain, we will always be getting a response back because the graceful migration is active. Meaning, even after the workload has been migrated, and the DNS record is updated, we will keep receiving a response without any interruption even after *kcp-cluster-1* has been deleted. This can be observed in our curl:

![Screenshot from 2022-08-02 12-55-24](https://user-images.githubusercontent.com/73656840/182368597-1ec0ade2-9849-4414-849f-ac342680d11b.png)

This shows that the workload has successfully migrated from cluster-1 to cluster-2 without any interruption.

<br>


### Clean-up

After finishing with this tutorial, we can go back to our tab where we deployed the sample service, and press the enter key to reset and undo everything that was done from running the sample.

![Screenshot from 2022-08-02 13-04-27](https://user-images.githubusercontent.com/73656840/182370379-4e5af83b-6ad9-4b2d-9b11-8be18edff290.png)




