package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/eks"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {

	prefix := "my-eks"
	pulumi.Run(func(ctx *pulumi.Context) error {
		vpc, err := ec2.NewVpc(ctx, prefix+"-vpc", &ec2.VpcArgs{
			CidrBlock:       pulumi.String("172.16.0.0/16"),
			InstanceTenancy: pulumi.String("default"),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("eks-vpc"),
			},
		})
		if err != nil {
			return err
		}
		privSubnet1, err := ec2.NewSubnet(ctx, prefix+"-priv-subnet-1", &ec2.SubnetArgs{
			VpcId:            vpc.ID(),
			CidrBlock:        pulumi.String("172.16.1.0/24"),
			AvailabilityZone: pulumi.String("ap-south-1a"),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("eks-subnet"),
			},
		})
		if err != nil {
			return err
		}

		privSubnet2, err := ec2.NewSubnet(ctx, prefix+"-priv-subnet-2", &ec2.SubnetArgs{
			VpcId:            vpc.ID(),
			CidrBlock:        pulumi.String("172.16.2.0/24"),
			AvailabilityZone: pulumi.String("ap-south-1b"),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("eks-subnet"),
			},
		})
		if err != nil {
			return err
		}

		pubSubnet1, err := ec2.NewSubnet(ctx, prefix+"-pub-subnet-1", &ec2.SubnetArgs{
			VpcId:            vpc.ID(),
			CidrBlock:        pulumi.String("172.16.5.0/24"),
			AvailabilityZone: pulumi.String("ap-south-1a"),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("eks-subnet"),
			},
		})
		if err != nil {
			return err
		}

		pubSubnet2, err := ec2.NewSubnet(ctx, prefix+"-pub-subnet-2", &ec2.SubnetArgs{
			VpcId:            vpc.ID(),
			CidrBlock:        pulumi.String("172.16.6.0/24"),
			AvailabilityZone: pulumi.String("ap-south-1b"),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("eks-subnet"),
			},
		})
		if err != nil {
			return err
		}

		eip, err := ec2.NewEip(ctx, prefix+"-eip", &ec2.EipArgs{
			Vpc: pulumi.Bool(true),
		})

		if err != nil {
			return err
		}

		ngw, err := ec2.NewNatGateway(ctx, prefix+"-ngw", &ec2.NatGatewayArgs{
			AllocationId: eip.ID(),
			SubnetId:     pubSubnet1.ID(),
		})

		igw, err := ec2.NewInternetGateway(ctx, prefix+"-igw", &ec2.InternetGatewayArgs{
			VpcId: vpc.ID(),
		})
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
			SubnetId:     privSubnet1.ID(),
			RouteTableId: pvtRouteTable.ID(),
		})
		if err != nil {
			return err
		}

		_, err = ec2.NewRouteTableAssociation(ctx, prefix+"-pvt2-rta", &ec2.RouteTableAssociationArgs{
			SubnetId:     privSubnet2.ID(),
			RouteTableId: pvtRouteTable.ID(),
		})
		if err != nil {
			return err
		}

		_, err = ec2.NewRouteTableAssociation(ctx, prefix+"-pub-rta", &ec2.RouteTableAssociationArgs{
			SubnetId:     pubSubnet1.ID(),
			RouteTableId: publicRouteTable.ID(),
		})
		if err != nil {
			return err
		}

		_, err = ec2.NewRouteTableAssociation(ctx, prefix+"-pubrta", &ec2.RouteTableAssociationArgs{
			SubnetId:     privSubnet1.ID(),
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

		_, err = eks.NewCluster(ctx, "eks-test-cluster", &eks.ClusterArgs{
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
						privSubnet1.ID(),
						privSubnet2.ID(),
						pubSubnet1.ID(),
						pubSubnet2.ID(),
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
		return nil
	})
}
