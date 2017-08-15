# Parsec-EC2
Easily start and stop Parsec-ready EC2 spot instances to make cloud gaming even cheaper.

## Background
Building on the excellent work done by [Larry Gadea](https://lg.io/), [Daniel Thomas](https://github.com/DanielThomas/ec2gaming),
[Josh McGhee](https://github.com/joshpmcghee/parsec-terraform), [Benjamin Malley](https://github.com/BenjaminMalley/ec2gaming),
and the [Parsec team](https://parsec.tv/), I started working on this project to allow for two very specific pieces of 
functionality I was looking for that I had not yet seen implemented anywhere else:

* To be able to easily switch between instance types without manually editing files
* To be able to arbitrarily specify a spot bid price per session relative to the current lowest spot price in a given availability zone

This is very much a work in progress and my first attempt at writing a non-trivial application in Go. Improvements and pull
requests are very much encouraged and welcome.

## Requirements
* [Terraform](https://github.com/hashicorp/terraform)
* [aws-cli](https://github.com/aws/aws-cli)
* [Go](https://github.com/golang/go)

This has been developed with MacOS in mind, but should also work on Linux.

## Installation
The latest version of `parsec-ec2` can be installed using `go get`.

```
go get -u github.com/LGUG2Z/parsec-ec2
```

Make sure `$GOPATH` is set correctly that and that `$GOPATH/bin` is in your `$PATH`.

The `parsec-ec2` executable will be installed under the `$GOPATH/bin` directory.

Once installed, add `export PARSEC_EC2_SERVER_KEY=your_server_key` to your `.bashrc` or `.zshrc` or `.${shell}rc` file.

## Usage
### init
After installation, all users should run `parsec-ec2 init`.

The init command will create the directory `$HOME/.parsec-ec2` and the required Terraform template and provisioning
userdata files.

This command should only need to be run once, but if a fresh copy of the files is needed, the `$HOME/.parsec-ec2`
folder can be manually removed and the `init` command run again.

### price
The `price` command looks for the current cheapest spot price for the requested instance type in the requested region
and returns the price along with the availability zone in the requested region where the cheapest price was found.

The `--region` and `--instance-type` flags are required.

Example:
```
$ parsec-ec2 price --region eu-west-1 --instance-type g2.2xlarge

>> 'eu-west-1a' is the least expensive availability zone in the region 
>> 'eu-west-1' for 'g2.2xlarge' instances with a spot price of $0.105800/hour.
```

### start
The `start` command sends a spot request for the requested EC2 instance type in the specified region.
If `PARSEC_EC2_SERVER_KEY` has not been exported in the shell rc file, it must be passed to the command using the `--server-key` flag.

The amount to bid above the current lowest spot price for the instance is specified using the `--bid` flag, so if the
current lowest spot price is $0.20, running the command with `--bid 0.10` will send a spot request with a bid price
of $0.30.

If the `--plan` flag is used, the spot request will not be sent and instead the `terraform plan` command will be run
which will output to the terminal the details of any AWS resources that will be created by running the `start` command.

Examples:
```
# With PARSEC_EC2_SERVER_KEY already set as an env variable
parsec-ec2 start \
--aws-region eu-west-1 \
--instance-type g3.4xlarge \
--bid 0.10
```
```
# With the server key passed using the --server-key flag
parsec-ec2 start \
--aws-region eu-west-2 \
--instance-type g2.2xlarge \
--bid 0.10 \ 
---server-key xxxxx
```
```
# With the --plan flag
parsec-ec2 start \
--aws-region eu-central-1 \
--instance-type g2.2xlarge \
--bid 0.10 \
--plan
```

### stop
The `stop` command stops a Parsec EC2 instance created using the start command. Under the hood this command runs 
`terraform destroy`, with removes all AWS resources that are identified for creation in the terraform template.

This command depends on session information that is created by the start command and stored in `$HOME/.parsec-ec2/currentSession.json`,
so if this has been manually modified or removed after running the start command, the stop command will not execute. In
this situation it is still possible to manually run `terraform destroy` in the `$HOME/.parsec-ec2` directory.

Example:

```
parsec-ec2 stop
```
