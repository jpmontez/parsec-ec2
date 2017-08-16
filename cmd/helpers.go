package cmd

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"encoding/json"
	"io"
	"io/ioutil"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"strings"
)

func getVpcId(svc *ec2.EC2) (string, error) {
	vpc, err := svc.DescribeVpcs(&ec2.DescribeVpcsInput{})

	if err != nil {
		return "", err
	}

	if len(vpc.Vpcs) < 1 {
		fmt.Println(`
You have at some point manually deleted the default VPC created by AWS for this region.
parsec-ec2 will not function for this region until a default VPC is recreated for it.
You can still try to launch Parsec EC2 instances in other regions that have retained
their default VPC.
`)
		os.Exit(0)
	}

	return *vpc.Vpcs[0].VpcId, nil
}

func getSubnetId(svc *ec2.EC2, availabilityZone string) (string, error) {
	values := []*string{&availabilityZone}

	filter := ec2.Filter{
		Name:   aws.String("availability-zone"),
		Values: values,
	}

	filters := []*ec2.Filter{&filter}

	describeSubnetsInput := ec2.DescribeSubnetsInput{
		Filters: filters,
	}

	result, err := svc.DescribeSubnets(&describeSubnetsInput)

	if err != nil {
		return "", err
	}

	if len(result.Subnets) == 0 {
		fmt.Printf("\nCould not get the subnet id for availability zone %'s'.\n", availabilityZone)
		os.Exit(1)
	}

	return *result.Subnets[0].SubnetId, nil
}

func getSpotPrice(svc *ec2.EC2, instanceType string) (ec2.SpotPrice, error) {
	instanceTypes := []*string{&instanceType}
	productDescriptions := []*string{&Windows}
	startTime := time.Now().AddDate(0, 0, -1)
	endTime := time.Now()

	result, err := svc.DescribeSpotPriceHistory(&ec2.DescribeSpotPriceHistoryInput{
		StartTime:           &startTime,
		EndTime:             &endTime,
		InstanceTypes:       instanceTypes,
		ProductDescriptions: productDescriptions,
	})

	if err != nil {
		return ec2.SpotPrice{}, err
	}

	if len(result.SpotPriceHistory) == 0 {
		fmt.Printf("\nCould not get the highest spot price for instance type '%s' in the \n"+
			"region '%s'. Either this instance type may not yet be available\n"+
			"in that region, or the instance type id given may contain a typo.\n", instanceType, awsRegion)
		os.Exit(0)
	}
	sort.Reverse(SpotPriceHistory(result.SpotPriceHistory))

	return *result.SpotPriceHistory[0], nil
}

type SpotPriceHistory []*ec2.SpotPrice

func (spotPriceHistory SpotPriceHistory) Len() int {
	return len(spotPriceHistory)
}

func (spotPriceHistory SpotPriceHistory) Less(i, j int) bool {
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

func (spotPriceHistory SpotPriceHistory) Swap(i, j int) {
	spotPriceHistory[i], spotPriceHistory[j] = spotPriceHistory[j], spotPriceHistory[i]
}

func copyFile(originalFilePath, destinationFileName, destinationFolder string) error {
	originalFile, err := os.Open(originalFilePath)
	if err != nil {
		return err
	}
	defer originalFile.Close()

	destinationFile, err := os.Create(fmt.Sprintf("%s/%s", destinationFolder, destinationFileName))
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, originalFile)
	if err != nil {
		return err
	}

	err = destinationFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

type TfVars struct {
	VpcId           string `json:"vpcId"`
	SubnetId        string `json:"subnetId"`
	AwsRegion       string `json:"awsRegion"`
	UserBid         string `json:"userBid"`
	Ec2InstanceType string `json:"instanceType"`
	ParsecServerKey string `json:"parsecServerKey"`
}

func hasParsecServerKey(parsecServerKey string) bool {
	return len(parsecServerKey) > 0
}

func writeSessionVars(vars TfVars) error {
	bytes, err := json.Marshal(vars)
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("%s/%s", appFolder, CurrentSession)

	return ioutil.WriteFile(filePath, bytes, 0644)
}

func constructTerraformCommand(vars TfVars, args []string) *exec.Cmd {
	command := exec.Command(Terraform, args...)

	command.Dir = appFolder

	command.Env = os.Environ()
	command.Env = append(command.Env, fmt.Sprintf("TF_VAR_vpc=%s", vars.VpcId))
	command.Env = append(command.Env, fmt.Sprintf("TF_VAR_subnet=%s", vars.SubnetId))
	command.Env = append(command.Env, fmt.Sprintf("TF_VAR_region=%s", vars.AwsRegion))
	command.Env = append(command.Env, fmt.Sprintf("TF_VAR_user_bid=%s", vars.UserBid))
	command.Env = append(command.Env, fmt.Sprintf("TF_VAR_instance_type=%s", vars.Ec2InstanceType))
	command.Env = append(command.Env, fmt.Sprintf("TF_VAR_parsec_server_key=%s", vars.ParsecServerKey))

	return command
}

func executeTerraformCommandAndSwallowOutput(command *exec.Cmd) error {
	err := command.Start()
	if err != nil {
		return err
	}

	return nil
}

func executeTerraformCommandAndReturnOutput(command *exec.Cmd) (string, error) {
	initOut, err := command.StdoutPipe()
	if err != nil {
		return "", err
	}

	initErr, err := command.StderrPipe()
	if err != nil {
		return "", err
	}

	err = command.Start()
	if err != nil {
		return "", err
	}

	stdOutput, err := ioutil.ReadAll(initOut)
	if err != nil {
		return "", err
	}

	_, err = ioutil.ReadAll(initErr)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(stdOutput)), nil
}

func executeTerraformCommandAndPrintOutput(command *exec.Cmd) error {
	initOut, err := command.StdoutPipe()
	if err != nil {
		return err
	}

	initErr, err := command.StderrPipe()
	if err != nil {
		return err
	}

	err = command.Start()
	if err != nil {
		return err
	}

	stdOutput, err := ioutil.ReadAll(initOut)
	if err != nil {
		return err
	}

	errOutput, err := ioutil.ReadAll(initErr)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", stdOutput)
	fmt.Printf("%s\n", errOutput)

	return nil
}

func isValidAwsRegion(validRegions []string, userRegion string) bool {
	for _, validRegion := range validRegions {
		if validRegion == userRegion {
			return true
		}
	}
	return false
}
