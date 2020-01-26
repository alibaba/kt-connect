Local Develope With Connect And Exchange
=========

For developers who use Kubernetes as the application runtime environment, we can use the namespace to quickly create multiple sets of isolation environments in the same cluster. In the same namespace, the services use the internal DNS domain name of the service to interact with each other. access. Based on Kubernetes' powerful isolation and service orchestration capabilities, the ability to define multiple deployments of YAML can be implemented.

However, in general, the container network used by Kubernetes is not directly connected to the developer's office network. Therefore, how to effectively use Kubernetes for joint testing of services has become a hurdle that cannot be avoided in daily development work. In this article, let’s talk about how to accelerate the development efficiency based on Kubernetes.

## Use Pipeline

In order to enable developers to deploy modified code to the cluster test environment more quickly, we will generally introduce a continuous delivery pipeline, which will solve the problem of code compilation, image packaging upload and deployment through automation. As follows:

![](../../media/guide/local-dev-01.png)


To a certain extent, this approach can avoid developers doing a lot of repetitive work. However, although the entire process is automated, developers also have to wait for the pipeline to run after each code change. For developers, waiting for the pipeline to run after each code change may have become the worst part of the entire development task.

## Breaking network restrictions

Ideally, the developer can start the service directly locally, and the service can be called seamlessly with each other in the remote kubernetes cluster. Need to solve two problems:

- I rely on other services: the code running locally can access other applications deployed in the cluster directly through the podIP, clusterIP or even the DNS address in the Kubernetes cluster, as shown below;
- Other services depend on me: other applications running in the Kubernetes cluster can access my running local code without any changes, as shown below:

![](../../media/guide/local-dev-02.png)

To achieve the two local joint adjustment methods just mentioned, we mainly need to solve the following three problems.：

- Direct connection between the local network and the Kubernetes cluster network
- Locally implement DNS resolution of internal services in Kubernetes
- If the traffic to other Pods in the cluster is transferred to the local;

## Connect And Exchange

The role of `Connect` is to break the network isolation between the local and the cluster, ensuring that local applications can directly access other services deployed in the Kubernetes cluster. At the same time, Connect provides the DNS service of the cluster cluster, so that developers can directly access the resources in the cluster such as PodIP, ClusterIP and DNS domain name directly from the cluster.

`Exchange` is responsible for getting traffic forwarding from the cluster to the local. Exchange will completely replace an application instance in the cluster, take over all received traffic and forward it to the local service port, thus enabling remote to local joint testing.