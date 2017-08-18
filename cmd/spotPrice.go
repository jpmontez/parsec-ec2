package cmd

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type spotPriceHistory []*ec2.SpotPrice

func (spotPriceHistory spotPriceHistory) Len() int {
	return len(spotPriceHistory)
}

func (spotPriceHistory spotPriceHistory) Less(i, j int) bool {
	price1, err := strconv.ParseFloat(*spotPriceHistory[i].SpotPrice, 32)
	if err != nil {
		panic(err)
	}
	price2, err := strconv.ParseFloat(*spotPriceHistory[j].SpotPrice, 32)
	if err != nil {
		panic(err)
	}
	return price1 < price2
}

func (spotPriceHistory spotPriceHistory) Swap(i, j int) {
	spotPriceHistory[i], spotPriceHistory[j] = spotPriceHistory[j], spotPriceHistory[i]
}

func getSpotPrice(svc *ec2.EC2, instanceType string) (ec2.SpotPrice, error) {
	instanceTypes := []*string{&instanceType}
	productDescriptions := []*string{aws.String(windows)}
	startTime := time.Now().AddDate(0, 0, -1)
	endTime := time.Now()

	describeSpotPriceHistoryInput := ec2.DescribeSpotPriceHistoryInput{
		StartTime:           &startTime,
		EndTime:             &endTime,
		InstanceTypes:       instanceTypes,
		ProductDescriptions: productDescriptions,
	}

	result, err := svc.DescribeSpotPriceHistory(&describeSpotPriceHistoryInput)

	if err != nil {
		return ec2.SpotPrice{}, err
	}

	if len(result.SpotPriceHistory) == 0 {
		fmt.Printf("\n%s instances are not yet available in the requested region.\n", instanceType)
		os.Exit(0)
	}
	sort.Reverse(spotPriceHistory(result.SpotPriceHistory))

	return *result.SpotPriceHistory[0], nil
}
