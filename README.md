<h1>Creating Infra using Pulumi and Golang</h1>

1. Install and setup Pulumi using https://www.pulumi.com/docs/clouds/aws/get-started/begin/
2. git clone https://github.com/rk280392/pulumi-infra-as-code
3. cd pulumi-infra-as-code
4. export AWS_PROFILE=my-profile
5. pulumi up

<h2>Get Cluster Context</h2>

1. cd pulumi-infra-as-code
2. AWS_REGION=$(pulumi stack | grep endpoint | awk '{print $2}' | awk -F '.' '{print $3}')
3. CLUSTER_NAME=$(pulumi stack | grep clusterName | awk '{print $2}')
4. aws eks update-kubeconfig --region $AWS_REGION --name $CLUSTER_NAME


<h2>Destroy Cluster</h2>

1. cd pulumi-infra-as-code
2. pulumi destroy

<h2> Contact Information </h2>

[Google](k90229@gmail.com)

[Linkedin](https://www.linkedin.com/in/rajesh-kumar-624082ab/) 
