set -eo pipefail
image=988080168334.dkr.ecr.us-east-2.amazonaws.com/gobuildlambda:latest
docker build -t $image .
docker push $image
aws lambda update-function-code --function-name gobuilder --image-uri $image
