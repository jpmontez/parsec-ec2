# Variables

variable "parsec_server_key" {
  type = "string"
}

variable "region" {
  type = "string"
}

variable "vpc" {
  type = "string"
}

variable "subnet" {
  type = "string"
}

variable "user_bid" {
  type = "string"
}

variable "instance_type" {
  type = "string"
}

# Template

provider "aws" {
  region = "${var.region}"
}

data "aws_ami" "parsec" {
  most_recent = true
  filter {
    name = "name"
    values = ["parsec-g3-ws2016-9"]
  }
}

resource "aws_security_group" "parsec" {
  vpc_id = "${var.vpc}"
  name = "parsec"
  description = "Allow inbound Parsec traffic and all outbound."

  ingress {
      from_port = 8000
      to_port = 8004
      protocol = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
      from_port = 5900
      to_port = 5900
      protocol = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
      from_port = 5900
      to_port = 5900
      protocol = "udp"
      cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
      from_port = 8000
      to_port = 8004
      protocol = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
      from_port = 8000
      to_port = 8004
      protocol = "udp"
      cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
      from_port = 0
      to_port = 0
      protocol = "-1"
      cidr_blocks = ["0.0.0.0/0"]
  }
}

data "template_file" "user_data" {
    template = "${file("user_data.tmpl")}"

    vars {
        server_key = "${var.parsec_server_key}"
    }
}

resource "aws_spot_instance_request" "parsec" {
    spot_price = "${var.user_bid}"
    ami = "${data.aws_ami.parsec.id}"
    subnet_id = "${var.subnet}"
    instance_type = "${var.instance_type}"

    tags {
        Name = "ParsecServer"
    }

    root_block_device {
      volume_size = 50
    }

    ebs_block_device {
      volume_size = 100
      volume_type = "gp2"
      device_name = "xvdg"
    }

    user_data = "${data.template_file.user_data.rendered}"

    vpc_security_group_ids = ["${aws_security_group.parsec.id}"]
    associate_public_ip_address = true
}

output "region" {
  value = "${var.region}"
}

output "subnet_id" {
  value = "${var.subnet}"
}

output "vpc_id" {
  value = "${var.vpc}"
}

output "spot_instance_id" {
  value = "${aws_spot_instance_request.parsec.spot_instance_id}"
}
