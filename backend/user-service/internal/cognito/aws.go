package cognito

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

type AWSConfig struct {
	Region string
}

func NewAWSClient(ctx context.Context, cfg AWSConfig) (*cognitoidentityprovider.Client, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	if err != nil {
		return nil, err
	}
	return cognitoidentityprovider.NewFromConfig(awsCfg), nil
}

/*

Dev Cognito UserPool Details:
User pool name
code-path - dev
User pool ID
us-east-1_TppiqQvor
ARN
arn:aws:cognito-idp:us-east-1:577206043902:userpool/us-east-1_TppiqQvor
*/
