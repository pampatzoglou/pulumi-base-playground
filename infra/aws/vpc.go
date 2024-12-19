package aws

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type VpcConfig struct {
	Vpc            *ec2.Vpc
	PrivateSubnets []*ec2.Subnet
	PublicSubnets  []*ec2.Subnet
}

func CreateVpc(ctx *pulumi.Context, name string, region string) (*VpcConfig, error) {
	// Create VPC
	vpc, err := ec2.NewVpc(ctx, name, &ec2.VpcArgs{
		CidrBlock:          pulumi.String("10.0.0.0/16"),
		EnableDnsHostnames: pulumi.Bool(true),
		EnableDnsSupport:   pulumi.Bool(true),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(name),
		},
	})
	if err != nil {
		return nil, err
	}

	// Create Internet Gateway
	igw, err := ec2.NewInternetGateway(ctx, name+"-igw", &ec2.InternetGatewayArgs{
		VpcId: vpc.ID(),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(name + "-igw"),
		},
	})
	if err != nil {
		return nil, err
	}

	// Create Public and Private Subnets across 3 AZs
	azs := []string{region + "a", region + "b", region + "c"}
	privateSubnets := make([]*ec2.Subnet, len(azs))
	publicSubnets := make([]*ec2.Subnet, len(azs))

	// Create NAT Gateway for each AZ
	eips := make([]*ec2.Eip, len(azs))
	natGateways := make([]*ec2.NatGateway, len(azs))

	for i, az := range azs {
		// Public Subnet
		publicSubnet, err := ec2.NewSubnet(ctx, name+"-public-"+az, &ec2.SubnetArgs{
			VpcId:               vpc.ID(),
			CidrBlock:          pulumi.String(generateSubnetCidr(i, true)),
			AvailabilityZone:   pulumi.String(az),
			MapPublicIpOnLaunch: pulumi.Bool(true),
			Tags: pulumi.StringMap{
				"Name": pulumi.String(name + "-public-" + az),
				"kubernetes.io/role/elb": pulumi.String("1"),
			},
		})
		if err != nil {
			return nil, err
		}
		publicSubnets[i] = publicSubnet

		// EIP for NAT Gateway
		eip, err := ec2.NewEip(ctx, name+"-eip-"+az, &ec2.EipArgs{
			Vpc: pulumi.Bool(true),
			Tags: pulumi.StringMap{
				"Name": pulumi.String(name + "-eip-" + az),
			},
		})
		if err != nil {
			return nil, err
		}
		eips[i] = eip

		// NAT Gateway
		natGateway, err := ec2.NewNatGateway(ctx, name+"-nat-"+az, &ec2.NatGatewayArgs{
			SubnetId:     publicSubnet.ID(),
			AllocationId: eip.ID(),
			Tags: pulumi.StringMap{
				"Name": pulumi.String(name + "-nat-" + az),
			},
		})
		if err != nil {
			return nil, err
		}
		natGateways[i] = natGateway

		// Private Subnet
		privateSubnet, err := ec2.NewSubnet(ctx, name+"-private-"+az, &ec2.SubnetArgs{
			VpcId:             vpc.ID(),
			CidrBlock:         pulumi.String(generateSubnetCidr(i, false)),
			AvailabilityZone: pulumi.String(az),
			Tags: pulumi.StringMap{
				"Name": pulumi.String(name + "-private-" + az),
				"kubernetes.io/role/internal-elb": pulumi.String("1"),
			},
		})
		if err != nil {
			return nil, err
		}
		privateSubnets[i] = privateSubnet
	}

	// Create Route Tables
	publicRT, err := ec2.NewRouteTable(ctx, name+"-public-rt", &ec2.RouteTableArgs{
		VpcId: vpc.ID(),
		Routes: ec2.RouteTableRouteArray{
			&ec2.RouteTableRouteArgs{
				CidrBlock: pulumi.String("0.0.0.0/0"),
				GatewayId: igw.ID(),
			},
		},
		Tags: pulumi.StringMap{
			"Name": pulumi.String(name + "-public-rt"),
		},
	})
	if err != nil {
		return nil, err
	}

	// Associate public subnets with public route table
	for i, subnet := range publicSubnets {
		_, err = ec2.NewRouteTableAssociation(ctx, name+"-public-rta-"+azs[i], &ec2.RouteTableAssociationArgs{
			SubnetId:     subnet.ID(),
			RouteTableId: publicRT.ID(),
		})
		if err != nil {
			return nil, err
		}
	}

	// Create private route tables and associate with private subnets
	for i, subnet := range privateSubnets {
		privateRT, err := ec2.NewRouteTable(ctx, name+"-private-rt-"+azs[i], &ec2.RouteTableArgs{
			VpcId: vpc.ID(),
			Routes: ec2.RouteTableRouteArray{
				&ec2.RouteTableRouteArgs{
					CidrBlock:    pulumi.String("0.0.0.0/0"),
					NatGatewayId: natGateways[i].ID(),
				},
			},
			Tags: pulumi.StringMap{
				"Name": pulumi.String(name + "-private-rt-" + azs[i]),
			},
		})
		if err != nil {
			return nil, err
		}

		_, err = ec2.NewRouteTableAssociation(ctx, name+"-private-rta-"+azs[i], &ec2.RouteTableAssociationArgs{
			SubnetId:     subnet.ID(),
			RouteTableId: privateRT.ID(),
		})
		if err != nil {
			return nil, err
		}
	}

	return &VpcConfig{
		Vpc:            vpc,
		PrivateSubnets: privateSubnets,
		PublicSubnets:  publicSubnets,
	}, nil
}

func generateSubnetCidr(index int, isPublic bool) string {
	if isPublic {
		return "10.0." + string(rune(index*16)) + ".0/20"
	}
	return "10.0." + string(rune(128+index*16)) + ".0/20"
}
