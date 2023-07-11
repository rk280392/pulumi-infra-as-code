package main

import (
	"infra-eks/ec2Instance"
	"infra-eks/eksCluster"
	"infra-eks/eksIAM"
	"infra-eks/vpc"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {

	k8sVersion := "1.27"
	prefix := "my-eks"
	vpcCIDR := "172.16.0.0/16"
	pubCIDRs := []string{"172.16.1.0/24", "172.16.2.0/24", "172.16.3.0/24"}
	pvtCIDRs := []string{"172.16.5.0/24", "172.16.6.0/24", "172.16.7.0/24"}
	azs := []string{"ap-south-1a", "ap-south-1b"}
	internetCIDR := "0.0.0.0/0"
	capacityType := "SPOT"
	instanceTypes := []string{"t3.medium"}
	desiredSize := 3
	minSize := 2
	maxSize := 5
	eksRole := `{
		"Version": "2012-10-17",
		"Statement": [
		{
			"Action": "sts:AssumeRole",
			"Principal": {
			"Service": "eks.amazonaws.com"
			},
			"Effect": "Allow",
			"Sid": ""
		}]
		}`
	eksPolicies := []string{
		"arn:aws:iam::aws:policy/AmazonEKSServicePolicy",
		"arn:aws:iam::aws:policy/AmazonEKSClusterPolicy",
	}
	nodeGroupRole := `{
		"Version": "2012-10-17",
		"Statement": [{
		"Sid": "",
		"Effect": "Allow",
		"Principal": {
			"Service": "ec2.amazonaws.com"
		},
		"Action": "sts:AssumeRole"
		}]
	}`
	nodeGroupPolicies := []string{
		"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
		"arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy",
		"arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
	}

	pulumi.Run(func(ctx *pulumi.Context) error {
		vpcInstance, err := vpc.CreateVPC(ctx, prefix, vpcCIDR)
		if err != nil {
			return err
		}

		SubnetGen := vpc.NewSubnetGenerator(ctx, pubCIDRs, pvtCIDRs, azs, vpcInstance.ID(), prefix)
		pvtSubnetIDs, err := SubnetGen.CreatePvtSubnet()
		if err != nil {
			return err
		}

		pubSubnetIDs, err := SubnetGen.CreatePubSubnet()
		if err != nil {
			return err
		}

		eip, err := ec2Instance.CreateEIPs(ctx, prefix)
		if err != nil {
			return err
		}

		ngw, err := vpc.CreateNatGateway(ctx, prefix, pubSubnetIDs[0], eip.ID())
		if err != nil {
			return err
		}

		igw, err := vpc.CreateInternetGateway(ctx, prefix, vpcInstance.ID())
		if err != nil {
			return err
		}

		pvtRouteTable, err := vpc.CreatePvtRouteTable(ctx, prefix, internetCIDR, ngw.ID(), vpcInstance.ID())
		if err != nil {
			return err
		}

		publicRouteTable, err := vpc.CreatePubRouteTable(ctx, prefix, internetCIDR, vpcInstance.ID(), igw.ID())
		if err != nil {
			return err
		}

		err = vpc.CreatePvtRouteTableAssoc(ctx, prefix, pvtSubnetIDs, pvtRouteTable.ID())
		if err != nil {
			return err
		}

		err = vpc.CreatePubRouteTableAssoc(ctx, prefix, pubSubnetIDs, publicRouteTable.ID())
		if err != nil {
			return err
		}

		roleInstance, err := eksIAM.CreateEKSRole(ctx, prefix, eksRole)
		if err != nil {
			return err
		}

		err = eksIAM.CreateEKSPolicies(ctx, prefix, eksPolicies, roleInstance.Name)
		if err != nil {
			return err
		}

		sgsInstance, err := ec2Instance.CreateSecurityGroup(ctx, prefix, internetCIDR, vpcInstance.ID())
		if err != nil {
			return err
		}

		cluster, err := eksCluster.CreateEKSCluster(ctx, prefix, internetCIDR, k8sVersion, roleInstance.Arn, sgsInstance.ID(), pvtSubnetIDs, pubSubnetIDs)
		if err != nil {
			return err
		}

		nodeRole, err := eksIAM.CreateNodeRole(ctx, prefix, nodeGroupRole)
		if err != nil {
			return err
		}

		err = eksIAM.CreateNodePolicies(ctx, prefix, nodeRole.Name, nodeGroupPolicies)
		if err != nil {
			return err
		}

		err = eksCluster.CreateEKSNodeGroup(ctx, prefix, capacityType, nodeRole.Arn, cluster.Name, instanceTypes, pvtSubnetIDs, desiredSize, maxSize, minSize)
		if err != nil {
			return err
		}
		ctx.Export("endpoint", cluster.Endpoint)
		ctx.Export("clusterName", cluster.Name)
		ctx.Export("kubeconfig-certificate-authority-data", cluster.CertificateAuthority.ApplyT(func(certificateAuthority eks.ClusterCertificateAuthority) (string, error) {
			return *certificateAuthority.Data, nil
		}).(pulumi.StringOutput))
		return nil
	})
}
