package aws

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"

	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/integr8ly/cluster-service/pkg/clusterservice"
	"github.com/sirupsen/logrus"
)

func TestRouteTableManager_DeleteResourcesForCluster(t *testing.T) {
	fakeLogger, err := fakeLogger(func(l *logrus.Entry) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		taggingClient func() *taggingClientMock
		logger        *logrus.Entry
		Ec2Api        ec2iface.EC2API
	}
	type args struct {
		clusterId string
		tags      map[string]string
		dryRun    bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*clusterservice.ReportItem
		wantErr string
	}{
		{
			name: "do not fail when getting route does not exists",
			fields: fields{
				Ec2Api: buildMockEc2Client(func(ec2Client *mockEc2Client) {
					ec2Client.deleteRouteTableFn = func(input *ec2.DeleteRouteTableInput) (*ec2.DeleteRouteTableOutput, error) {
						return nil, awserr.New("InvalidRouteTableID.NotFound", "", errors.New("InvalidRouteTableID.NotFound"))
					}
				}),
				taggingClient: func() *taggingClientMock {
					client, err := fakeTaggingClient(func(c *taggingClientMock) error {
						c.GetResourcesFunc = func(in1 *resourcegroupstaggingapi.GetResourcesInput) (*resourcegroupstaggingapi.GetResourcesOutput, error) {
							return &resourcegroupstaggingapi.GetResourcesOutput{
								ResourceTagMappingList: []*resourcegroupstaggingapi.ResourceTagMapping{
									fakeResourceTagMapping(func(mapping *resourcegroupstaggingapi.ResourceTagMapping) {
										mapping.ResourceARN = aws.String(fakeEc2ClientInstanceArn)
									}),
								},
							}, nil
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return client
				},
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeClusterId,
				tags:      map[string]string{},
				dryRun:    false,
			},
			want: []*clusterservice.ReportItem{
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.ID = fakeEc2ClientInstanceArn
					item.Name = fakeResourceIdentifier
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusComplete
				}),
			},
		},
		{
			name: "fail when getting resources via tags returns an error",
			fields: fields{
				Ec2Api: buildMockEc2Client(func(ec2Client *mockEc2Client) {
					ec2Client.deleteRouteTableFn = func(input *ec2.DeleteRouteTableInput) (*ec2.DeleteRouteTableOutput, error) {
						return &ec2.DeleteRouteTableOutput{}, nil
					}
				}),
				taggingClient: func() *taggingClientMock {
					client, err := fakeTaggingClient(func(c *taggingClientMock) error {
						c.GetResourcesFunc = func(in1 *resourcegroupstaggingapi.GetResourcesInput) (output *resourcegroupstaggingapi.GetResourcesOutput, e error) {
							return nil, errors.New("")
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return client
				},
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeClusterId,
				tags:      map[string]string{},
				dryRun:    false,
			},
			wantErr: "failed to filter route tables: ",
		},
		{
			name: "fail when vpc deletion returns an error",
			fields: fields{
				Ec2Api: buildMockEc2Client(func(ec2Client *mockEc2Client) {
					ec2Client.deleteRouteTableFn = func(input *ec2.DeleteRouteTableInput) (*ec2.DeleteRouteTableOutput, error) {
						return nil, errors.New("some error deleting route")
					}
				}),
				taggingClient: func() *taggingClientMock {
					client, err := fakeTaggingClient(func(c *taggingClientMock) error {
						c.GetResourcesFunc = func(in1 *resourcegroupstaggingapi.GetResourcesInput) (*resourcegroupstaggingapi.GetResourcesOutput, error) {
							return &resourcegroupstaggingapi.GetResourcesOutput{
								ResourceTagMappingList: []*resourcegroupstaggingapi.ResourceTagMapping{
									fakeResourceTagMapping(func(mapping *resourcegroupstaggingapi.ResourceTagMapping) {
										mapping.ResourceARN = aws.String(fakeEc2ClientInstanceArn)
									}),
								},
							}, nil
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return client
				},
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeClusterId,
				tags:      map[string]string{},
				dryRun:    false,
			},
			wantErr: "failed to delete route table: some error deleting route",
		},
		{
			name: "succeeds with status completed if dry run is false and no errors on delete aka successful deletion",
			fields: fields{
				Ec2Api: buildMockEc2Client(func(ec2Client *mockEc2Client) {
					ec2Client.deleteRouteTableFn = func(input *ec2.DeleteRouteTableInput) (*ec2.DeleteRouteTableOutput, error) {
						return &ec2.DeleteRouteTableOutput{}, nil
					}
				}),
				taggingClient: func() *taggingClientMock {
					client, err := fakeTaggingClient(func(c *taggingClientMock) error {
						c.GetResourcesFunc = func(in1 *resourcegroupstaggingapi.GetResourcesInput) (*resourcegroupstaggingapi.GetResourcesOutput, error) {
							return &resourcegroupstaggingapi.GetResourcesOutput{
								ResourceTagMappingList: []*resourcegroupstaggingapi.ResourceTagMapping{
									fakeResourceTagMapping(func(mapping *resourcegroupstaggingapi.ResourceTagMapping) {
										mapping.ResourceARN = aws.String(fakeEc2ClientInstanceArn)
									}),
								},
							}, nil
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return client
				},
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeClusterId,
				tags:      map[string]string{},
				dryRun:    false,
			},
			want: []*clusterservice.ReportItem{
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.ID = fakeEc2ClientInstanceArn
					item.Name = fakeResourceIdentifier
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusComplete
				}),
			},
		},
		{
			name: "succeeds with status dry run if dry run is true",
			fields: fields{
				Ec2Api: buildMockEc2Client(func(ec2Client *mockEc2Client) {
					ec2Client.deleteRouteTableFn = func(input *ec2.DeleteRouteTableInput) (*ec2.DeleteRouteTableOutput, error) {
						return &ec2.DeleteRouteTableOutput{}, nil
					}
				}),
				taggingClient: func() *taggingClientMock {
					client, err := fakeTaggingClient(func(c *taggingClientMock) error {
						c.GetResourcesFunc = func(in1 *resourcegroupstaggingapi.GetResourcesInput) (*resourcegroupstaggingapi.GetResourcesOutput, error) {
							return &resourcegroupstaggingapi.GetResourcesOutput{
								ResourceTagMappingList: []*resourcegroupstaggingapi.ResourceTagMapping{
									fakeResourceTagMapping(func(mapping *resourcegroupstaggingapi.ResourceTagMapping) {
										mapping.ResourceARN = aws.String(fakeEc2ClientInstanceArn)
									}),
								},
							}, nil
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return client
				},
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeClusterId,
				tags:      map[string]string{},
				dryRun:    true,
			},
			want: []*clusterservice.ReportItem{
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.ID = fakeEc2ClientInstanceArn
					item.Name = fakeResourceIdentifier
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusDryRun
				}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RouteTableManager{
				ec2Client:     tt.fields.Ec2Api,
				taggingClient: tt.fields.taggingClient(),
				logger:        tt.fields.logger,
			}
			got, err := r.DeleteResourcesForCluster(tt.args.clusterId, tt.args.tags, tt.args.dryRun)
			if tt.wantErr != "" && err.Error() != tt.wantErr {
				t.Errorf("DeleteResourcesForCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteResourcesForCluster() got = %v, want %v", buildReportItemsString(got), buildReportItemsString(tt.want))
			}
		})
	}
}
