package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Network struct {
	VpcID           pulumi.IDOutput
	PublicSubnetIDs []pulumi.IDOutput
	PrivateSubnetID pulumi.IDOutput
}

func createNetwork(ctx *pulumi.Context) (*Network, error) {
	vpc, err := ec2.NewVpc(ctx, "example_vpc", &ec2.VpcArgs{
		CidrBlock: pulumi.String("10.100.0.0/16"),
		Tags: pulumi.StringMap{
			"Name": pulumi.String("example_vpc"),
		},
	})
	if err != nil {
		return nil, err
	}

	var pubSubIDs []pulumi.IDOutput
	pubSub1a, err := ec2.NewSubnet(ctx, "pubSub1a", &ec2.SubnetArgs{
		VpcId:            vpc.ID(),
		CidrBlock:        pulumi.String("10.100.0.0/24"),
		AvailabilityZone: pulumi.String("ap-northeast-1a"),
		Tags: pulumi.StringMap{
			"Name": pulumi.String("example_public_1a"),
		},
	})
	if err != nil {
		return nil, err
	}

	pubSubIDs = append(pubSubIDs, pubSub1a.ID())

	pubSub1c, err := ec2.NewSubnet(ctx, "pubSub1c", &ec2.SubnetArgs{
		VpcId:            vpc.ID(),
		CidrBlock:        pulumi.String("10.100.1.0/24"),
		AvailabilityZone: pulumi.String("ap-northeast-1c"),
		Tags: pulumi.StringMap{
			"Name": pulumi.String("example_public_1c"),
		},
	})
	if err != nil {
		return nil, err
	}

	pubSubIDs = append(pubSubIDs, pubSub1c.ID())

	priSub1a, err := ec2.NewSubnet(ctx, "priSub1a", &ec2.SubnetArgs{
		VpcId:            vpc.ID(),
		CidrBlock:        pulumi.String("10.100.100.0/24"),
		AvailabilityZone: pulumi.String("ap-northeast-1a"),
		Tags: pulumi.StringMap{
			"Name": pulumi.String("example_private_1a"),
		},
	})
	if err != nil {
		return nil, err
	}

	eip, err := ec2.NewEip(ctx, "eip", &ec2.EipArgs{
		Vpc: pulumi.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	igw, err := ec2.NewInternetGateway(ctx, "igw", &ec2.InternetGatewayArgs{
		VpcId: vpc.ID(),
		Tags: pulumi.StringMap{
			"Name": pulumi.String("example_igw"),
		},
	})
	if err != nil {
		return nil, err
	}

	natgw, err := ec2.NewNatGateway(ctx, "natgw", &ec2.NatGatewayArgs{
		AllocationId: eip.ID(),
		SubnetId:     pubSub1a.ID(),
	}, pulumi.DependsOn([]pulumi.Resource{eip}))
	if err != nil {
		return nil, err
	}

	pubRoutaTable, err := ec2.NewRouteTable(ctx, "pubRouteTable", &ec2.RouteTableArgs{
		VpcId: vpc.ID(),
		Tags: pulumi.StringMap{
			"Name": pulumi.String("example_route_table_public"),
		},
	})
	if err != nil {
		return nil, err
	}

	_, err = ec2.NewRoute(ctx, "pubRoute", &ec2.RouteArgs{
		RouteTableId:         pubRoutaTable.ID(),
		DestinationCidrBlock: pulumi.String("0.0.0.0/0"),
		GatewayId:            igw.ID(),
	})
	if err != nil {
		return nil, err
	}

	_, err = ec2.NewRouteTableAssociation(ctx, "pubRoute1a", &ec2.RouteTableAssociationArgs{
		SubnetId:     pubSub1a.ID(),
		RouteTableId: pubRoutaTable.ID(),
	})
	if err != nil {
		return nil, err
	}

	_, err = ec2.NewRouteTableAssociation(ctx, "pubRoute1c", &ec2.RouteTableAssociationArgs{
		SubnetId:     pubSub1c.ID(),
		RouteTableId: pubRoutaTable.ID(),
	})
	if err != nil {
		return nil, err
	}

	priRouteTable, err := ec2.NewRouteTable(ctx, "priRouteTable", &ec2.RouteTableArgs{
		VpcId: vpc.ID(),
		Tags: pulumi.StringMap{
			"Name": pulumi.String("example_route_table_private"),
		},
	})
	if err != nil {
		return nil, err
	}

	_, err = ec2.NewRoute(ctx, "priRoute", &ec2.RouteArgs{
		RouteTableId:         priRouteTable.ID(),
		DestinationCidrBlock: pulumi.String("0.0.0.0/0"),
		NatGatewayId:         natgw.ID(),
	})
	if err != nil {
		return nil, err
	}

	_, err = ec2.NewRouteTableAssociation(ctx, "priRoute1a", &ec2.RouteTableAssociationArgs{
		SubnetId:     priSub1a.ID(),
		RouteTableId: priRouteTable.ID(),
	})
	if err != nil {
		return nil, err
	}

	network := new(Network)
	network.VpcID = vpc.ID()
	network.PublicSubnetIDs = pubSubIDs
	network.PrivateSubnetID = priSub1a.ID()

	return network, nil
}
