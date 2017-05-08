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
	currentCapacity int64
}

func getASG(asgName string, awsRegion string, verbose bool) (*AutoScalingGroup, error) {

	svc := autoscaling.New(session.New(), &aws.Config{Region: aws.String(awsRegion)})

	// -------- Fetch the ASG from AWS --------
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
		return nil, err
	}

	// Pretty-print the response data.
	if verbose {
		fmt.Println(resp)
	}
	asg := resp.AutoScalingGroups[0]
	// -------- We now have the ASG --------

	scale := func(scaleFactor int64, dryRun bool) error {
		current := aws.Int64Value(asg.DesiredCapacity)
		desired := scaleFactor + current

		min := aws.Int64Value(asg.MinSize)
		max := aws.Int64Value(asg.MaxSize)

		log.Printf("Desired: %d, Min: %d, Max: %d", desired, min, max)
		if desired < min {
			log.Printf("Desired %d is less than group min size %d, setting to min", desired, min)
			desired = min
		}
		if desired > max {
			log.Printf("Desired %d is more than group max size %d, setting to max", desired, max)
			desired = max
		}
		if desired == current {
			log.Printf("Current size %d is same as desired: %d Nothing to do here", current, desired)
			return nil
		}

		params := &autoscaling.SetDesiredCapacityInput{
			AutoScalingGroupName: aws.String(asgName), // Required
			DesiredCapacity:      aws.Int64(desired),  // Required
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
		aws.Int64Value(asg.DesiredCapacity),
	}, nil
}
