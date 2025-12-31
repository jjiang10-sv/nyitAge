#!/usr/bin/env python3
from cdktf import App
from stacks.aws_stack import AwsStack
from stacks.azure_stack import AzureStack

app = App()
AwsStack(app, "aws-eks-cilium-typed")
AzureStack(app, "azure-aks-cilium-typed")
app.synth()
