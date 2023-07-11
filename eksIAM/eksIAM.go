package eksIAM

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateEKSRole(ctx *pulumi.Context, prefix, eksRole string) (*iam.Role, error) {
	role, err := iam.NewRole(ctx, prefix+"-role", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(eksRole),
	})
	if err != nil {
		return nil, err
	}
	return role, nil
}

func CreateEKSPolicies(ctx *pulumi.Context, prefix string, eksPolicies []string, roleName pulumi.Input) error {
	for i, eksPolicy := range eksPolicies {
		_, err := iam.NewRolePolicyAttachment(ctx, fmt.Sprintf(prefix+"-policy-%d", i), &iam.RolePolicyAttachmentArgs{
			PolicyArn: pulumi.String(eksPolicy),
			Role:      roleName,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateNodeRole(ctx *pulumi.Context, prefix, nodeGroupRole string) (*iam.Role, error) {
	nodeRole, err := iam.NewRole(ctx, prefix+"-ng-role", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(nodeGroupRole),
	})
	if err != nil {
		return nil, err
	}
	return nodeRole, nil
}

func CreateNodePolicies(ctx *pulumi.Context, prefix string, nodeRoleName pulumi.StringInput, nodeGroupPolicies []string) error {
	for i, nodeGroupPolicy := range nodeGroupPolicies {
		_, err := iam.NewRolePolicyAttachment(ctx, fmt.Sprintf(prefix+"-ngp-%d", i), &iam.RolePolicyAttachmentArgs{
			Role:      nodeRoleName,
			PolicyArn: pulumi.String(nodeGroupPolicy),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
