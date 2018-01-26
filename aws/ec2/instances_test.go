package ec2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevievelesperance/leftovers/aws/ec2"
	"github.com/genevievelesperance/leftovers/aws/ec2/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Instances", func() {
	var (
		client *fakes.InstancesClient
		logger *fakes.Logger

		instances ec2.Instances
	)

	BeforeEach(func() {
		client = &fakes.InstancesClient{}
		logger = &fakes.Logger{}
		logger.PromptCall.Returns.Proceed = true

		instances = ec2.NewInstances(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			client.DescribeInstancesCall.Returns.Output = &awsec2.DescribeInstancesOutput{
				Reservations: []*awsec2.Reservation{{
					Instances: []*awsec2.Instance{{
						State: &awsec2.InstanceState{Name: aws.String("available")},
						Tags: []*awsec2.Tag{{
							Key:   aws.String("Name"),
							Value: aws.String("banana-instance"),
						}},
						InstanceId: aws.String("the-instance-id"),
						KeyName:    aws.String(""),
					}},
				}},
			}
		})

		It("returns a list of ec2 instances to delete", func() {
			items, err := instances.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DescribeInstancesCall.CallCount).To(Equal(1))

			Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to terminate instance the-instance-id (Name:banana-instance)?"))

			Expect(items).To(HaveLen(1))
			Expect(items).To(HaveKeyWithValue("the-instance-id (Name:banana-instance)", "the-instance-id"))
		})

		Context("when the instance name does not contain the filter", func() {
			It("does not try to delete it", func() {
				items, err := instances.List("kiwi")
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DescribeInstancesCall.CallCount).To(Equal(1))

				Expect(logger.PromptCall.CallCount).To(Equal(0))

				Expect(items).To(HaveLen(0))
			})
		})

		Context("when there is no tag name", func() {
			BeforeEach(func() {
				client.DescribeInstancesCall.Returns.Output = &awsec2.DescribeInstancesOutput{
					Reservations: []*awsec2.Reservation{{
						Instances: []*awsec2.Instance{{
							State:      &awsec2.InstanceState{Name: aws.String("available")},
							InstanceId: aws.String("the-instance-id"),
							KeyName:    aws.String(""),
						}},
					}},
				}
			})

			It("uses just the instance id in the prompt", func() {
				items, err := instances.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to terminate instance the-instance-id?"))

				Expect(items).To(HaveLen(1))
				Expect(items).To(HaveKeyWithValue("the-instance-id", "the-instance-id"))
			})
		})

		Context("when there is a key name", func() {
			BeforeEach(func() {
				client.DescribeInstancesCall.Returns.Output = &awsec2.DescribeInstancesOutput{
					Reservations: []*awsec2.Reservation{{
						Instances: []*awsec2.Instance{{
							State:      &awsec2.InstanceState{Name: aws.String("available")},
							InstanceId: aws.String("the-instance-id"),
							KeyName:    aws.String("the-key-pair"),
						}},
					}},
				}
			})

			It("uses just the instance id in the prompt", func() {
				items, err := instances.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to terminate instance the-instance-id (KeyPairName:the-key-pair)?"))

				Expect(items).To(HaveLen(1))
				Expect(items).To(HaveKeyWithValue("the-instance-id (KeyPairName:the-key-pair)", "the-instance-id"))
			})
		})

		Context("when the instance state is terminated", func() {
			BeforeEach(func() {
				client.DescribeInstancesCall.Returns.Output = &awsec2.DescribeInstancesOutput{
					Reservations: []*awsec2.Reservation{{
						Instances: []*awsec2.Instance{{
							State:      &awsec2.InstanceState{Name: aws.String("terminated")},
							InstanceId: aws.String("the-instance-id"),
						}},
					}},
				}
			})

			It("does not return it in the list", func() {
				items, err := instances.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DescribeInstancesCall.CallCount).To(Equal(1))
				Expect(client.TerminateInstancesCall.CallCount).To(Equal(0))
				Expect(logger.PromptCall.CallCount).To(Equal(0))
				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the client fails to list instances", func() {
			BeforeEach(func() {
				client.DescribeInstancesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := instances.List(filter)
				Expect(err).To(MatchError("Describing instances: some error"))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptCall.Returns.Proceed = false
			})

			It("does not return it to the list", func() {
				items, err := instances.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to terminate instance the-instance-id (Name:banana-instance)?"))
				Expect(items).To(HaveLen(0))
			})
		})
	})

	Describe("Delete", func() {
		var items map[string]string

		BeforeEach(func() {
			items = map[string]string{"the-instance-id (Name:banana-instance)": "the-instance-id"}
		})

		It("terminates ec2 instances", func() {
			err := instances.Delete(items)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.TerminateInstancesCall.CallCount).To(Equal(1))
			Expect(client.TerminateInstancesCall.Receives.Input.InstanceIds).To(HaveLen(1))
			Expect(client.TerminateInstancesCall.Receives.Input.InstanceIds[0]).To(Equal(aws.String("the-instance-id")))

			Expect(logger.PrintfCall.Messages).To(Equal([]string{"SUCCESS terminating instance the-instance-id (Name:banana-instance)\n"}))
		})

		Context("when the client fails to terminate the instance", func() {
			BeforeEach(func() {
				client.TerminateInstancesCall.Returns.Error = errors.New("some error")
			})

			It("logs the error", func() {
				err := instances.Delete(items)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PrintfCall.Messages).To(Equal([]string{"ERROR terminating instance the-instance-id (Name:banana-instance): some error\n"}))
			})
		})
	})
})
