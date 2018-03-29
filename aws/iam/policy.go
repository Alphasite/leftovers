package iam

import (
	"fmt"

	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type Policy struct {
	client     policiesClient
	logger     logger
	name       *string
	arn        *string
	identifier string
	rtype      string
}

func NewPolicy(client policiesClient, logger logger, name, arn *string) Policy {
	return Policy{
		client:     client,
		logger:     logger,
		name:       name,
		arn:        arn,
		identifier: *name,
		rtype:      "IAM Policy",
	}
}

func (p Policy) Delete() error {
	versions, err := p.client.ListPolicyVersions(&awsiam.ListPolicyVersionsInput{PolicyArn: p.arn})
	if err != nil {
		return fmt.Errorf("List policy versions for %s %s: %s", p.rtype, p.identifier, err)
	}

	for _, v := range versions.Versions {
		if !*v.IsDefaultVersion {
			_, err := p.client.DeletePolicyVersion(&awsiam.DeletePolicyVersionInput{
				PolicyArn: p.arn,
				VersionId: v.VersionId,
			})
			if err != nil {
				p.logger.Printf("[WARNING] Delete policy version %s of %s %s: %s\n", *v.VersionId, p.rtype, p.identifier, err)
			}
		}
	}

	_, err = p.client.DeletePolicy(&awsiam.DeletePolicyInput{PolicyArn: p.arn})
	if err != nil {
		return fmt.Errorf("Delete %s %s: %s", p.rtype, p.identifier, err)
	}

	return nil
}

func (p Policy) Name() string {
	return p.identifier
}

func (p Policy) Type() string {
	return p.rtype
}
