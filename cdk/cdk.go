package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3deployment"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

const (
	domainName = "isthereaseattlehomegametoday.com"
)

type CdkStackProps struct {
	awscdk.StackProps
}

func NewCdkStack(scope constructs.Construct, id string, props *CdkStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	bucket := awss3.NewBucket(stack, jsii.String("hosting-bucket"), &awss3.BucketProps{
		AccessControl: awss3.BucketAccessControl_PRIVATE,
	})

	originAccessIdentity := awscloudfront.NewOriginAccessIdentity(stack, jsii.String("OriginAccessID"), &awscloudfront.OriginAccessIdentityProps{})
	bucket.GrantRead(originAccessIdentity, nil)

	awss3deployment.NewBucketDeployment(stack, jsii.String("WebsiteDeployment"), &awss3deployment.BucketDeploymentProps{
		DestinationBucket: bucket,
		Sources: &[]awss3deployment.ISource{
			awss3deployment.Source_Asset(jsii.String("../static"), nil),
		},
	})

	hostedZone := awsroute53.HostedZone_FromLookup(stack, jsii.String("HostedZone"), &awsroute53.HostedZoneProviderProps{
		DomainName: jsii.String(domainName),
	})
	cert := awscertificatemanager.NewCertificate(stack, jsii.String("TLSCertificate"), &awscertificatemanager.CertificateProps{
		DomainName: jsii.String(domainName),
		Validation: awscertificatemanager.CertificateValidation_FromDns(hostedZone),
	})

	distribution := awscloudfront.NewDistribution(stack, jsii.String("CloudfrontWebsiteDistribution"), &awscloudfront.DistributionProps{
		DefaultRootObject: jsii.String("index.html"),
		DefaultBehavior: &awscloudfront.BehaviorOptions{
			Origin:               awscloudfrontorigins.NewS3Origin(bucket, &awscloudfrontorigins.S3OriginProps{OriginAccessIdentity: originAccessIdentity}),
			Compress:             jsii.Bool(true),
			AllowedMethods:       awscloudfront.AllowedMethods_ALLOW_GET_HEAD(),
			ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
		},
		DomainNames:            jsii.Strings(domainName),
		Certificate:            cert,
		MinimumProtocolVersion: awscloudfront.SecurityPolicyProtocol_TLS_V1_2_2021,
	})

	awsroute53.NewARecord(stack, jsii.String("DomainARecord"), &awsroute53.ARecordProps{
		Zone:       hostedZone,
		RecordName: jsii.String(domainName),
		Target:     awsroute53.RecordTarget_FromAlias(awsroute53targets.NewCloudFrontTarget(distribution)),
	})

	updateFunction := awslambda.NewFunction(stack, jsii.String("UpdateFunction"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("bootstrap"),
		Code:    awslambda.Code_FromAsset(jsii.String("../bin"), &awss3assets.AssetOptions{}),
		Environment: &map[string]*string{
			"UPLOAD_S3_BUCKET_NAME":     bucket.BucketName(),
			"UPLOAD_CF_DISTRIBUTION_ID": distribution.DistributionId(),
		},
	})

	bucket.GrantWrite(updateFunction, nil, nil)
	distribution.GrantCreateInvalidation(updateFunction)

	eventRule := awsevents.NewRule(stack, jsii.String("UpdateFunctionCron"), &awsevents.RuleProps{
		Schedule: awsevents.Schedule_Cron(&awsevents.CronOptions{
			Hour:   jsii.String("9"),
			Minute: jsii.String("0"),
		}),
	})
	eventRule.AddTarget(awseventstargets.NewLambdaFunction(updateFunction, &awseventstargets.LambdaFunctionProps{}))

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewCdkStack(app, "SeattleSportsTodayStack", &CdkStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("SPORTS_AWS_ACCOUNT_ID")),
		Region:  jsii.String("us-east-1"),
	}

}
