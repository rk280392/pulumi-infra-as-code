package main

import (
	"fmt"
	"infra-eks/eip"
	"infra-eks/internetGateway"
	"infra-eks/natGateway"
	"infra-eks/subnet"
	"infra-eks/vpc"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/eks"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {

	prefix := "my-eks"
	vpcCIDR := "172.16.0.0/16"
	pubCIDRs := []string{"172.16.1.0/24", "172.16.2.0/24", "172.16.3.0/24"}
	pvtCIDRs := []string{"172.16.5.0/24", "172.16.6.0/24", "172.16.7.0/24"}
	azs := []string{"ap-south-1a", "ap-south-1b"}

	pulumi.Run(func(ctx *pulumi.Context) error {
		vpc, err := vpc.CreateVPC(ctx, prefix, vpcCIDR)
		if err != nil {
			return err
		}

		SubnetGen := subnet.NewSubnetGenerator(ctx, pubCIDRs, pvtCIDRs, azs, vpc.ID(), prefix)
		pvtSubnetIDs, err := SubnetGen.CreatePvtSubnet()
		if err != nil {
			return err
		}

		pubSubnetIDs, err := SubnetGen.CreatePubSubnet()
		if err != nil {
			return err
		}

		eip, err := eip.CreateEIPs(ctx, prefix)
		if err != nil {
			return err
		}

		ngw, err := natGateway.CreateNatGateway(ctx, prefix, pubSubnetIDs[0], eip.ID())
		if err != nil {
			return err
		}

		igw, err := internetGateway.CreateInternetGateway(ctx, prefix, vpc.ID())
		if err != nil {
			return err
		}

		pvtRouteTable, err := ec2.NewRouteTable(ctx, prefix+"-pvt-rt", &ec2.RouteTableArgs{
			VpcId: vpc.ID(),
			Routes: ec2.RouteTableRouteArray{
				&ec2.RouteTableRouteArgs{
					CidrBlock: pulumi.String("0.0.0.0/0"),
					GatewayId: ngw.ID(),
				},
			},
		})
		if err != nil {
			return err
		}

		publicRouteTable, err := ec2.NewRouteTable(ctx, prefix+"pub-rt", &ec2.RouteTableArgs{
			VpcId: vpc.ID(),
			Routes: ec2.RouteTableRouteArray{
				&ec2.RouteTableRouteArgs{
					CidrBlock: pulumi.String("0.0.0.0/0"),
					GatewayId: igw.ID(),
				},
			},
		})

		if err != nil {
			return err
		}

		_, err = ec2.NewRouteTableAssociation(ctx, prefix+"-pvt1-rta", &ec2.RouteTableAssociationArgs{
			SubnetId:     pvtSubnetIDs[0],
			RouteTableId: pvtRouteTable.ID(),
		})
		if err != nil {
			return err
		}

		_, err = ec2.NewRouteTableAssociation(ctx, prefix+"-pvt2-rta", &ec2.RouteTableAssociationArgs{
			SubnetId:     pvtSubnetIDs[1],
			RouteTableId: pvtRouteTable.ID(),
		})
		if err != nil {
			return err
		}

		_, err = ec2.NewRouteTableAssociation(ctx, prefix+"-pub-rta", &ec2.RouteTableAssociationArgs{
			SubnetId:     pubSubnetIDs[0],
			RouteTableId: publicRouteTable.ID(),
		})
		if err != nil {
			return err
		}

		_, err = ec2.NewRouteTableAssociation(ctx, prefix+"-pubrta", &ec2.RouteTableAssociationArgs{
			SubnetId:     pvtSubnetIDs[1],
			RouteTableId: pvtRouteTable.ID(),
		})
		if err != nil {
			return err
		}

		role, err := iam.NewRole(ctx, "eks-role", &iam.RoleArgs{
			AssumeRolePolicy: pulumi.String(`{
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
	    }`),
		})
		if err != nil {
			return err
		}
		eksPolicies := []string{
			"arn:aws:iam::aws:policy/AmazonEKSServicePolicy",
			"arn:aws:iam::aws:policy/AmazonEKSClusterPolicy",
		}
		for i, eksPolicy := range eksPolicies {
			_, err := iam.NewRolePolicyAttachment(ctx, fmt.Sprintf(prefix+"rpa-%d", i), &iam.RolePolicyAttachmentArgs{
				PolicyArn: pulumi.String(eksPolicy),
				Role:      role.Name,
			})
			if err != nil {
				return err
			}
		}

		sgs, err := ec2.NewSecurityGroup(ctx, prefix+"-sg", &ec2.SecurityGroupArgs{
			VpcId: vpc.ID(),
			Egress: ec2.SecurityGroupEgressArray{
				ec2.SecurityGroupEgressArgs{
					Protocol:   pulumi.String("-1"),
					FromPort:   pulumi.Int(0),
					ToPort:     pulumi.Int(0),
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				},
			},
			Ingress: ec2.SecurityGroupIngressArray{
				ec2.SecurityGroupIngressArgs{
					Protocol:   pulumi.String("tcp"),
					FromPort:   pulumi.Int(80),
					ToPort:     pulumi.Int(80),
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				},
			},
		})

		if err != nil {
			return err
		}

		var subnetIds pulumi.StringArray

		cluster, err := eks.NewCluster(ctx, "eks-test-cluster", &eks.ClusterArgs{
			RoleArn: role.Arn,
			VpcConfig: &eks.ClusterVpcConfigArgs{
				PublicAccessCidrs: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
				SecurityGroupIds: pulumi.StringArray{
					sgs.ID().ToStringOutput(),
				},
				SubnetIds: pulumi.StringArray(
					append(
						subnetIds,
						pvtSubnetIDs[0],
						pvtSubnetIDs[1],
						pubSubnetIDs[0],
						pubSubnetIDs[1],
					)),
			},
			Version: pulumi.String("1.27"),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("my-eks-cluster"),
			},
		})
		if err != nil {
			return err
		}

		nodeGroupRole, err := iam.NewRole(ctx, prefix+"nodegroup-role", &iam.RoleArgs{
			AssumeRolePolicy: pulumi.String(`{
		    "Version": "2012-10-17",
		    "Statement": [{
			"Sid": "",
			"Effect": "Allow",
			"Principal": {
			    "Service": "ec2.amazonaws.com"
			},
			"Action": "sts:AssumeRole"
		    }]
		}`),
		})
		if err != nil {
			return err
		}
		nodeGroupPolicies := []string{
			"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
			"arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy",
			"arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
		}
		for i, nodeGroupPolicy := range nodeGroupPolicies {
			_, err := iam.NewRolePolicyAttachment(ctx, fmt.Sprintf(prefix+"ngr-%d", i), &iam.RolePolicyAttachmentArgs{
				Role:      nodeGroupRole.Name,
				PolicyArn: pulumi.String(nodeGroupPolicy),
			})
			if err != nil {
				return err
			}
		}

		var privateSubnets pulumi.StringArray

		_, err = eks.NewNodeGroup(ctx, prefix+"-ng-1", &eks.NodeGroupArgs{
			ClusterName: cluster.Name,
			InstanceTypes: pulumi.StringArray{
				pulumi.String("t3.medium"),
			},
			CapacityType: pulumi.String("SPOT"),
			SubnetIds: pulumi.StringArray(
				append(
					privateSubnets,
					pvtSubnetIDs[0],
					pvtSubnetIDs[1],
				)),
			NodeRoleArn: pulumi.StringInput(nodeGroupRole.Arn),
			ScalingConfig: &eks.NodeGroupScalingConfigArgs{
				DesiredSize: pulumi.Int(3),
				MinSize:     pulumi.Int(2),
				MaxSize:     pulumi.Int(5),
			},
		})
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
