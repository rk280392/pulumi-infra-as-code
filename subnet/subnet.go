package subnet

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type SubnetGen struct {
	pubCIDRs []string
	pvtCIDRs []string
	azs      []string
	vpcId    pulumi.StringInput
	prefix   string
	ctx      *pulumi.Context
}

func NewSubnetGenerator(ctx *pulumi.Context, pubCIDRs, pvtCIDRs, azs []string, vpcId pulumi.StringInput, prefix string) *SubnetGen {
	return &SubnetGen{
		pubCIDRs: pubCIDRs,
		pvtCIDRs: pvtCIDRs,
		azs:      azs,
		vpcId:    vpcId,
		prefix:   prefix,
		ctx:      ctx,
	}
}

//var pubSubnetID []*ec2.Subnet
//pubSubnetsMap := make(map[string]string)

func (sg *SubnetGen) CreatePvtSubnet() ([]pulumi.IDOutput, error) {
	var pvtSubnetID []pulumi.IDOutput
	pvtSubnetsMap := make(map[string]string)

	for i, key := range sg.pvtCIDRs {
		value := sg.azs[i%len(sg.azs)]
		pvtSubnetsMap[key] = value
	}

	i := 0
	for cidr, az := range pvtSubnetsMap {
		subnet, err := ec2.NewSubnet(sg.ctx, fmt.Sprintf(sg.prefix+"-pvt-sub-%d", i), &ec2.SubnetArgs{
			VpcId:            sg.vpcId,
			CidrBlock:        pulumi.String(cidr),
			AvailabilityZone: pulumi.String(az),
		})
		if err != nil {
			return nil, err
		}
		pvtSubnetID = append(pvtSubnetID, subnet.ID())
		i++
	}
	return pvtSubnetID, nil
}

func (sg *SubnetGen) CreatePubSubnet() ([]pulumi.IDOutput, error) {
	var pubSubnetID []pulumi.IDOutput
	pubSubnetsMap := make(map[string]string)

	for i, key := range sg.pubCIDRs {
		value := sg.azs[i%len(sg.azs)]
		pubSubnetsMap[key] = value
	}

	i := 0
	for cidr, az := range pubSubnetsMap {
		subnet, err := ec2.NewSubnet(sg.ctx, fmt.Sprintf(sg.prefix+"-pub-sub-%d", i), &ec2.SubnetArgs{
			VpcId:            sg.vpcId,
			CidrBlock:        pulumi.String(cidr),
			AvailabilityZone: pulumi.String(az),
		})
		if err != nil {
			return nil, err
		}
		pubSubnetID = append(pubSubnetID, subnet.ID())
		i++
	}
	return pubSubnetID, nil
}
