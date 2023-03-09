---
Title: "Build Go Lambdas from Terraform"
Date: 2022-11-01
Tags: [aws, golang, meta, homelab]
---

AWS Lambda Functions are a really handy way to run infrastructure scripts, but
coordinating the build and deployment of these functions is more complicated.
For small Python lambdas that only use the Python standard library + boto3 (the
AWS APIs), no building is required, but to use a compiled language or even
Python with some other third-party dependencies, a build step is required.
Unfortunately, Terraform doesn't have any good solution for building these
lambdas.

In this post, I'll describe the solution I struck upon for building Go lambdas,
although the general approach should be broadly applicable to any kind of
lambda.

<!-- more -->

# Requirements

Before embarking on the solution, let's talk through some guiding principles
and requirements.

## Short iteration loops

Short iteration loops make developer velocity faster. As it pertains to
building and deploying lambdas (or anything else, really), I think it's
really important that developers can iterate locally without needing to commit,
push, and wait for CI resources to spin up. This also means that it's important
that we don't do expensive build/deploy steps if the lambda source code or
configuration haven't changed.

## Reproducibility

Since we're assuming local iteration, we want to make sure that the local
environment builds the same artifact that the build pipeline would do, and that
it deploys it the same way that the deployment pipeline would do. In the
context of building a Go lambda, this mostly means making sure that the build
and deploy scripts behave the same locally as they do in CI/CD *and* that the
version of the Go toolchain is the same (Go projects have a sort of lockfile
mechanism so the same versions of the dependencies are always pulled, but we
still need to make sure the toolchain version is fixed so we compile those
dependencies the same way).

There are a couple of approaches to "reproducibility"--you can either build
locally, controlling the build environment via lockfile, Docker container, or
reproducible build tool (Nix, Bazel, etc), or you can build remotely in a CI
environment or some other builder service.

I decided that the local options aren't very good because (1) they all involve
developers installing new tooling on their machines and I like to minimize dev
dependencies (2) Docker is a pretty heavy-weight dependency (both in terms of
resource consumption / performance, and maintenance--pruning unused images,
volumes, etc) (3) there is no good lockfile mechanism or similar (that I'm
aware of) for managing the Go build tool across developer environments and CI,
and (4) reproducible build tools are too complex just to control the version of
the Go toolchain.

Similarly, we already ruled out building in CI, so that leaves "calling a
remote 'build service'".

# Solution

To keep things simple, I prefer to just have the build happen as part of the
`terraform apply` for a couple of reasons:

1. It's easier for developers to reason about rather than opening up a shell
   script
2. Terraform is already good at only doing things when the inputs have changed,
   which we can use to avoid unnecessary rebuilds

The "build service" design I opted for is itself a lambda that has the Go
toolchain installed, takes some zipped Go source code, compiles the source into
a lambda-ready artifact, stores the artifact in s3, and returns the s3 path to
the caller. Since the AWS provider for Terraform already includes an
`aws_lambda_invocation` resource that only invokes the lambda when parameters
have changed, making the call is conceptually straightforward, although there
are some hoops to jump through to build the zip archive.

Also, I'm calling this service `gobuilder` because I'm not especially creative.

## `gobuilder` Terraform

<details>

```hcl
locals {
  bucket_prefix = "gobuilder"
  lambda_name   = "gobuilder"

  default_tags = { Application = "gobuilder" }
}

module "bucket" {
  source           = "../aws/s3/bucket"
  name             = var.bucket_name
  retention_policy = 30 # 30 days
  tags             = local.default_tags
}

module "lambda" {
  source = "../aws/lambda/function"

  name         = "gobuilder"
  tags         = local.default_tags
  image_uri    = "988080168334.dkr.ecr.us-east-2.amazonaws.com/gobuildlambda:latest"
  memory_size  = 2056
  package_type = "Image"
  timeout      = 120

  environment = {
    BUCKET        = module.bucket.bucket.id
    BUCKET_PREFIX = local.bucket_prefix
  }

  inline_policies = {
    s3-put-object = {
      Version = "2012-10-17"
      Statement = [
        {
          Action   = ["s3:PutObject"]
          Effect   = "Allow"
          Resource = "arn:aws:s3:::${var.bucket_name}/${local.bucket_prefix}/*"
        },
      ]
    }
  }
}
```

</details>

It's basically just a Lambda function and an S3 bucket for depositing the
artifacts. The S3 bucket has a retention policy of 30 days, so old artifacts
get collected. The Lambda function gets created with a policy that allows it
to put objects into the bucket. Note also that the lambda is created with
`package_type = "Image"`, which means we're using a container image rather than
a zipfile as the code for the lambda to execute (this is the only reasonable
way to package the Go toolchain into a Lambda function). Further, the memory
and timeout values were tweaked.

## `gobuilder` client Terraform

Clients of `gobuilder` will invoke it by zipping the source code, passing
the zipped source code to an `aws_lambda_invocation` resource, and then
creating an `aws_lambda_function` from the returned S3 path.

### Zipping the source code

This is surprisingly complicated because Terraform doesn't allow you to zip
files directly into memory--you have to zip them to some temporary file with
the `archive_file` data source and then read that file into memory with the
`local_file` data source (you can't use the `file()` function because it
doesn't support generated files).

<details>

```hcl
locals {
  archive_path       = "/tmp/archive.zip"
  compilation_result = jsondecode(aws_lambda_invocation.build.result)
}

data "archive_file" "lambda_source" {
  type = "zip"

  source {
    content  = file("${path.module}/go.mod")
    filename = "go.mod"
  }

  source {
    content  = file("${path.module}/go.sum")
    filename = "go.sum"
  }

  source {
    content  = file("${path.module}/main.go")
    filename = "main.go"
  }

  output_path = local.archive_path
}

# Read the local file deposited by data.archive_file.lambda_source. This is
# necessary because archive_file doesn't have an attribute containing the
# archive contents, and gobuilder expects the archive contents in the request
# payload.
data "local_file" "archive" {
  filename   = local.archive_path
  depends_on = [data.archive_file.lambda_source]
}

resource "aws_lambda_invocation" "build" {
  function_name = "gobuilder"
  input = jsonencode({
    name         = local.lambda_name
    architecture = "amd64"
    archive      = data.local_file.archive.content_base64
  })
}
```

</details>

### Invoking `gobuilder` and creating the lambda function resource

Here we're just passing the base64-encoded zip file data to the `gobuilder`
lambda via the `aws_lambda_invocation` resource (which will only actually call
the lambda if the source code changes) and then pulling the s3 bucket/key info
from the lambda's response and passing it into the s3 bucket/key fields on an
`aws_lambda_function` resource.

<details>

```hcl
resource "aws_lambda_invocation" "build" {
  function_name = "gobuilder"
  input = jsonencode({
    name         = local.lambda_name
    architecture = "amd64"
    archive      = data.local_file.archive.content_base64
  })
}

resource "aws_lambda_function" "target" {
  function_name = local.lambda_name
  handler       = "main"
  runtime       = "go1.x"
  s3_bucket     = local.compilation_result["bucket"]
  s3_key        = local.compilation_result["key"]
  timeout       = 10
  tags          = { Application = local.lambda_name }
  role          = aws_iam_role.target_role.arn
}
```

</details>
