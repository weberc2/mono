# Analytics

Analytics is a Lambda Function for collecting web analytics. It's intended for
my blog. This project only contains the source code for the system; the
infrastructure is currently in my private infrastructure repository. This repo's
CI/CD will build and push a container image to an ECR registry and (eventually)
update the Lambda function to reference the new container image.