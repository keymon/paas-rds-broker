package helpers

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
)

type RDSClient struct {
	region string
	rdssvc *rds.RDS
}

func NewRDSClient(region string) (*RDSClient, error) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		fmt.Println("Failed to create AWS session,", err)
		return nil, err
	}

	rdssvc := rds.New(sess)
	return &RDSClient{
		region: region,
		rdssvc: rdssvc,
	}, nil
}

func (r *RDSClient) Ping() (bool, error) {
	params := &rds.DescribeDBEngineVersionsInput{}

	_, err := r.rdssvc.DescribeDBEngineVersions(params)

	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *RDSClient) GetDBFinalSnapshots(ID string) (*rds.DescribeDBSnapshotsOutput, error) {
	params := &rds.DescribeDBSnapshotsInput{
		DBSnapshotIdentifier: aws.String(ID + "-final-snapshot"),
	}

	resp, err := r.rdssvc.DescribeDBSnapshots(params)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *RDSClient) deleteDBFinalSnapshot(ID string) (*rds.DeleteDBSnapshotOutput, error) {
	params := &rds.DeleteDBSnapshotInput{
		DBSnapshotIdentifier: aws.String(ID + "-final-snapshot"),
	}

	resp, err := r.rdssvc.DeleteDBSnapshot(params)

	if err != nil {
		return nil, err
	}
	return resp, nil
}
