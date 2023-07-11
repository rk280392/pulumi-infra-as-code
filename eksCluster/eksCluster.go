package eksCluster

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateEKSCluster(ctx *pulumi.Context, prefix, internetCIDR, k8sVersion string, roleArn pulumi.StringOutput, sgID pulumi.IDOutput, pvtSubnetIDs, pubSubnetIDs []pulumi.IDOutput) (*eks.Cluster, error) {

	subnetIdsStringArray := pulumi.StringArray{}
	subnetIDs := append(pvtSubnetIDs, pubSubnetIDs...)
	for _, op := range subnetIDs {
		subnetIdsStringArray = append(subnetIdsStringArray, op.ToStringOutput())
	}

	cluster, err := eks.NewCluster(ctx, prefix+"-cluster", &eks.ClusterArgs{
		RoleArn: roleArn,
		VpcConfig: &eks.ClusterVpcConfigArgs{
			PublicAccessCidrs: pulumi.StringArray{
				pulumi.String(internetCIDR),
			},
			SecurityGroupIds: pulumi.StringArray{
				sgID.ToStringOutput(),
			},
			SubnetIds: subnetIdsStringArray,
		},
		Version: pulumi.String(k8sVersion),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(prefix + "-cluster"),
		},
	})
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func CreateEKSNodeGroup(ctx *pulumi.Context, prefix, capacityType string, nodeRoleArn, clusterName pulumi.StringOutput, instanceTypes []string, pvtSubnetIDs []pulumi.IDOutput, desiredSize, maxSize, minSize int) error {
	pvtSubnetIdsStringArray := pulumi.StringArray{}
	for _, op := range pvtSubnetIDs {
		pvtSubnetIdsStringArray = append(pvtSubnetIdsStringArray, op.ToStringOutput())
	}

	instanceTypesStringArray := pulumi.StringArray{}
	for _, op := range instanceTypes {
		instanceTypesStringArray = append(instanceTypesStringArray, pulumi.String(op))
	}

	_, err := eks.NewNodeGroup(ctx, prefix+"-ng-1", &eks.NodeGroupArgs{
		ClusterName:   clusterName,
		InstanceTypes: instanceTypesStringArray,
		CapacityType:  pulumi.String(capacityType),
		SubnetIds:     pvtSubnetIdsStringArray,
		NodeRoleArn:   nodeRoleArn,
		ScalingConfig: &eks.NodeGroupScalingConfigArgs{
			DesiredSize: pulumi.Int(desiredSize),
			MinSize:     pulumi.Int(minSize),
			MaxSize:     pulumi.Int(maxSize),
		},
	})
	if err != nil {
		return err
	}
	return nil
}
