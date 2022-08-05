package main

import (
	"encoding/base64"
	"io/ioutil"
	"os"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createApplication(ctx *pulumi.Context) error {
	network, err := createNetwork(ctx)
	if err != nil {
		return err
	}

	sgForALB, err := ec2.NewSecurityGroup(ctx, "sgForALB", &ec2.SecurityGroupArgs{
		Name:  pulumi.String("example_sg_for_ALB"),
		VpcId: network.VpcID,
		Ingress: ec2.SecurityGroupIngressArray{
			&ec2.SecurityGroupIngressArgs{
				FromPort: pulumi.Int(80),
				ToPort:   pulumi.Int(80),
				Protocol: pulumi.String("tcp"),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			&ec2.SecurityGroupEgressArgs{
				FromPort: pulumi.Int(0),
				ToPort:   pulumi.Int(0),
				Protocol: pulumi.String("-1"),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
			},
		},
	})
	if err != nil {
		return err
	}

	alb, err := lb.NewLoadBalancer(ctx, "ALB", &lb.LoadBalancerArgs{
		Name:             pulumi.String("example"),
		LoadBalancerType: pulumi.String("application"),
		SecurityGroups: pulumi.StringArray{
			sgForALB.ID(),
		},
		Subnets: pulumi.StringArray{
			network.PublicSubnetIDs[0],
			network.PublicSubnetIDs[1],
		},
		Tags: pulumi.StringMap{
			"Name": pulumi.String("example"),
		},
	})
	if err != nil {
		return err
	}

	httpTG, err := lb.NewTargetGroup(ctx, "httpTG", &lb.TargetGroupArgs{
		Name:     pulumi.String("HTTPTG"),
		Port:     pulumi.Int(80),
		Protocol: pulumi.String("HTTP"),
		VpcId:    network.VpcID,
		HealthCheck: lb.TargetGroupHealthCheckArgs{
			Path:     pulumi.String("/"),
			Matcher:  pulumi.String("403"),
			Port:     pulumi.String("80"),
			Protocol: pulumi.String("HTTP"),
		},
	})
	if err != nil {
		return err
	}

	_, err = lb.NewListener(ctx, "listener", &lb.ListenerArgs{
		LoadBalancerArn: alb.Arn,
		Port:            pulumi.Int(80),
		Protocol:        pulumi.String("HTTP"),
		DefaultActions: lb.ListenerDefaultActionArray{
			&lb.ListenerDefaultActionArgs{
				Type:           pulumi.String("forward"),
				TargetGroupArn: httpTG.Arn,
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{alb}))
	if err != nil {
		return err
	}

	sgForInstance, err := ec2.NewSecurityGroup(ctx, "sgForInstance", &ec2.SecurityGroupArgs{
		Name:  pulumi.String("example_sg_for_instance"),
		VpcId: network.VpcID,
		Ingress: ec2.SecurityGroupIngressArray{
			&ec2.SecurityGroupIngressArgs{
				FromPort: pulumi.Int(80),
				ToPort:   pulumi.Int(80),
				Protocol: pulumi.String("tcp"),
				SecurityGroups: pulumi.StringArray{
					sgForALB.ID(),
				},
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			&ec2.SecurityGroupEgressArgs{
				FromPort: pulumi.Int(0),
				ToPort:   pulumi.Int(0),
				Protocol: pulumi.String("-1"),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
			},
		},
	})
	if err != nil {
		return err
	}

	f, err := os.Open("install_apache.sh")
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	enc := base64.StdEncoding.EncodeToString(b)

	ins, err := ec2.NewInstance(ctx, "instance", &ec2.InstanceArgs{
		Ami:          pulumi.String("ami-0992fc94ca0f1415a"),
		InstanceType: pulumi.String("t2.micro"),
		SubnetId:     network.PrivateSubnetID,
		VpcSecurityGroupIds: pulumi.StringArray{
			sgForInstance.ID(),
		},
		UserDataBase64: pulumi.String(enc),
	})
	if err != nil {
		return err
	}

	_, err = lb.NewTargetGroupAttachment(ctx, "TGattach", &lb.TargetGroupAttachmentArgs{
		TargetGroupArn: httpTG.Arn,
		TargetId:       ins.ID(),
		Port:           pulumi.Int(80),
	})
	if err != nil {
		return err
	}

	return nil
}
