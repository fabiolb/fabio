variable "public_key_path" {
  default = "~/.ssh/terraform.pub"
}

variable "key_name" {
  default = "terraform"
}

variable "aws_region" {
  description = "AWS region to launch servers."
  default = "eu-west-1"
}

# Ubuntu Precise 14.04 LTS (x64)
variable "aws_amis" {
  default = {
    eu-west-1 = "ami-f95ef58a"
  }
}
