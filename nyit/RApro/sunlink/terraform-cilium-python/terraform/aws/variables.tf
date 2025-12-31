variable "region" { default = "us-west-2" }
variable "cluster_name" { default = "cilium-eks" }
variable "k8s_version" { default = "1.26" }
variable "node_count" { default = 2 }
variable "node_instance_type" { default = "t3.medium" }
variable "vpc_id" { default = "" }
variable "public_subnets" { type = list(string) default = [] }
