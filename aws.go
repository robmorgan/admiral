package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

func awsConfig() (*aws.Config, error) {
	return &aws.Config{}, nil
}

type config struct {
	AwsAccessKeyID     string `json:"aws_access_key_id"`     // "key",
	AwsSecretAccessKey string `json:"aws_secret_access_key"` // "secret",
	AwsDefaultRegion   string `json:"aws_default_region"`    // "eu-west-1",
}

type provider struct {
	accessKeyID     string
	secretAccessKey string
}

func (p *provider) IsExpired() bool {
	return false
}

func (p *provider) Retrieve() (credentials.Value, error) {
	return credentials.Value{AccessKeyID: p.accessKeyID, SecretAccessKey: p.secretAccessKey}, nil
}
