package scaler

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
)

type AutoScalingGroup struct {
	scale           func(scaleFactor int64, dryRun bool) error
	currentCapacity func() (int64, error)
}

func getASG(asgName string, awsRegion string, verbose bool) *AutoScalingGroup {

	var currentCap int64
	svc := autoscaling.New(session.New(), &aws.Config{Region: aws.String(awsRegion)})

	currentCapacity := func() (int64, error) {
		if currentCap != 0 {
			return currentCap, nil
		}
		params := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{
				aws.String(asgName), // Required
			},
			MaxRecords: aws.Int64(1),
		}
		resp, err := svc.DescribeAutoScalingGroups(params)

		if err != nil {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
			return -1, err
		}

		// Pretty-print the response data.
		if verbose {
			fmt.Println(resp)
		}

		group := resp.AutoScalingGroups[0]
		currentCap = aws.Int64Value(group.DesiredCapacity)
		return currentCap, nil
	}

	scale := func(scaleFactor int64, dryRun bool) error {
		current, err := currentCapacity()

		if err != nil {
			log.Fatalln("Error getting current group capacity")
		}

		params := &autoscaling.SetDesiredCapacityInput{
			AutoScalingGroupName: aws.String(asgName),              // Required
			DesiredCapacity:      aws.Int64(scaleFactor + current), // Required
			HonorCooldown:        aws.Bool(true),
		}

		if !dryRun {
			log.Println("Scaling with params:", params)
			resp, err := svc.SetDesiredCapacity(params)

			if err != nil {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				fmt.Println(err.Error())
				return err
			}
			// Pretty-print the response data.
			fmt.Println("Scale response: ", resp)
		} else {
			log.Printf("Would have applied: %+v", params)
		}
		return nil
	}

	return &AutoScalingGroup{
		scale,
		currentCapacity,
	}
}
