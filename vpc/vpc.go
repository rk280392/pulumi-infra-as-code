package vpc

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateVPC(ctx *pulumi.Context, prefix, vpcCIDR string) (*ec2.Vpc, error) {
	vpc, err := ec2.NewVpc(ctx, prefix+"-vpc", &ec2.VpcArgs{
		CidrBlock: pulumi.String(vpcCIDR),
		Tags: pulumi.StringMap{
			"Name": pulumi.String("eks-vpc"),
		},
	})
	if err != nil {
		return nil, err
	}
	return vpc, nil
}

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

func (sg *SubnetGen) CreatePvtSubnet() ([]pulumi.IDOutput, error) {

	// Create private subnets. The plan is to create a subnet in each AZ. depending on the length of az and subnet slices, we might create multiple subnets in one az.
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

func CreateInternetGateway(ctx *pulumi.Context, prefix string, vpcID pulumi.IDOutput) (*ec2.InternetGateway, error) {
	igw, err := ec2.NewInternetGateway(ctx, prefix+"-igw", &ec2.InternetGatewayArgs{
		VpcId: vpcID,
	})
	if err != nil {
		return nil, err
	}
	return igw, nil
}

func CreateNatGateway(ctx *pulumi.Context, prefix string, subnetID, eip pulumi.IDOutput) (*ec2.NatGateway, error) {
	ngw, err := ec2.NewNatGateway(ctx, prefix+"-ngw", &ec2.NatGatewayArgs{
		AllocationId: eip,
		SubnetId:     subnetID,
	})
	if err != nil {
		return nil, err
	}
	return ngw, nil
}

func CreatePvtRouteTable(ctx *pulumi.Context, prefix, internetCIDR string, ngwID, vpcID pulumi.StringInput) (*ec2.RouteTable, error) {
	pvtRouteTable, err := ec2.NewRouteTable(ctx, prefix+"-pvt-rt", &ec2.RouteTableArgs{
		VpcId: vpcID,
		Routes: ec2.RouteTableRouteArray{
			&ec2.RouteTableRouteArgs{
				CidrBlock: pulumi.String(internetCIDR),
				GatewayId: ngwID,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return pvtRouteTable, nil
}
func CreatePubRouteTable(ctx *pulumi.Context, prefix, internetCIDR string, vpcID, igwID pulumi.StringInput) (*ec2.RouteTable, error) {
	publicRouteTable, err := ec2.NewRouteTable(ctx, prefix+"-pub-rt", &ec2.RouteTableArgs{
		VpcId: vpcID,
		Routes: ec2.RouteTableRouteArray{
			&ec2.RouteTableRouteArgs{
				CidrBlock: pulumi.String(internetCIDR),
				GatewayId: igwID,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return publicRouteTable, nil
}

func CreatePvtRouteTableAssoc(ctx *pulumi.Context, prefix string, pvtSubnetIDs []pulumi.IDOutput, rtID pulumi.IDOutput) error {

	for i, id := range pvtSubnetIDs {
		_, err := ec2.NewRouteTableAssociation(ctx, fmt.Sprintf(prefix+"-pvt-rta-%d", i), &ec2.RouteTableAssociationArgs{
			SubnetId:     id,
			RouteTableId: rtID,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func CreatePubRouteTableAssoc(ctx *pulumi.Context, prefix string, pubSubnetIDs []pulumi.IDOutput, rtID pulumi.IDOutput) error {

	for i, id := range pubSubnetIDs {
		_, err := ec2.NewRouteTableAssociation(ctx, fmt.Sprintf(prefix+"-pub-rta-%d", i), &ec2.RouteTableAssociationArgs{
			SubnetId:     id,
			RouteTableId: rtID,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
