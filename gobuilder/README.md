# GoBuilder

`gobuilder` is a lambda function that accepts Go source code as input and
compiles it into Lambda-ready binary artifacts.

The purpose of the `gobuilder` system is to allow Terraform to create AWS
Lambda functions from Go source code by sending that source code to the
`gobuilder` lambda function, which will compile the source code, zip up the
resulting binary artifact (into a Lambda-compatible zip file), and deposit the
zipped artifact into S3 for reference by an `aws_lambda_function` resource.
