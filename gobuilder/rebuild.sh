set -eo pipefail
image=988080168334.dkr.ecr.us-east-2.amazonaws.com/gobuilder:latest
export AWS_DEFAULT_REGION=us-east-2
docker build -t $image -f ../docker/gobuilder/Dockerfile .
docker push $image
aws lambda update-function-code \
    --function-name gobuilder \
    --image-uri $image \
    --no-cli-pager
