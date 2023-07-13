<h1>Creating Infra using Pulumi and Golang</h1>

 <h2>Before you begin</h2>
 
 Install and setup Pulumi using pulumi [docs](https://www.pulumi.com/docs/clouds/aws/get-started/begin/) 

 <h3> Steps: </h3>
 
```shell

1. git clone https://github.com/rk280392/pulumi-infra-as-code
2. cd pulumi-infra-as-code
3. export AWS_PROFILE=my-profile
# pulumi up creates or update infrastructure resources based on your code and configuration.
4. pulumi up
```
<h4>Get Cluster Context</h4>

```shell

1. cd pulumi-infra-as-code
2. AWS_REGION=$(pulumi stack | grep endpoint | awk '{print $2}' | awk -F '.' '{print $3}')
3. CLUSTER_NAME=$(pulumi stack | grep clusterName | awk '{print $2}')
# update KUBECONFIG with the the new cluster config
4. aws eks update-kubeconfig --region $AWS_REGION --name $CLUSTER_NAME
```

<h4>Destroy Cluster</h4>

```shell

1. cd pulumi-infra-as-code
# pulumi destroy deletes and removes infrastructure resources managed by Pulumi.
2. pulumi destroy

```
<h2> Contact Information </h2>

[Google](k90229@gmail.com)

[Linkedin](https://www.linkedin.com/in/rajesh-kumar-624082ab/) 
