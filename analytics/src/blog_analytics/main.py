import json

from nimbus_core import Sub, Template, AccountID
from nimbus_resources.apigatewayv2.api import Api
from nimbus_resources.iam.role import Role, Policy
from nimbus_resources.lambda_.function import Function, Code, Environment
from nimbus_resources.s3.bucket import (
    Bucket,
    BucketEncryption,
    ServerSideEncryptionRule,
    ServerSideEncryptionByDefault,
)
from nimbus_resources.crawler import (
    Crawler,
    Targets,
    S3Target,
    SchemaChangePolicy,
    Schedule,
)
from nimbus_resources.glue.database import Database, DatabaseInput
from nimbus_util.iam import PolicyDocument, Statement, Principal


def main():
    source_bucket = Bucket(
        BucketName=Sub("${AWS::AccountId}-${AWS::StackName}-analytics"),
        BucketEncryption=BucketEncryption(
            ServerSideEncryptionConfiguration=[
                ServerSideEncryptionRule(
                    ServerSideEncryptionByDefault=ServerSideEncryptionByDefault(
                        SSEAlgorithm="AES256",
                    ),
                )
            ],
        ),
    )

    function_role = Role(
        RoleName=Sub("${AWS::StackName}-lambda-execution"),
        Description="Execution role for analytics lambda function",
        AssumeRolePolicyDocument=PolicyDocument(
            Version="2012-10-17",
            Statement=[
                Statement(
                    Effect="Allow",
                    Principal=Principal(Service=["lambda.amazonaws.com"]),
                    Action=["sts:AssumeRole"],
                )
            ],
        ),
        Policies=[
            Policy(
                PolicyName=Sub("${AWS::StackName}-lambda-logs"),
                PolicyDocument=PolicyDocument(
                    Version="2012-10-17",
                    Statement=[
                        Statement(
                            Action=[
                                "logs:CreateLogGroup",
                                "logs:CreateLogStream",
                                "logs:PutLogEvents",
                            ],
                            Effect="Allow",
                            Resource="*",
                        )
                    ],
                ),
            ),
            Policy(
                PolicyName=Sub("${AWS::StackName}-lambda-s3"),
                PolicyDocument=PolicyDocument(
                    Version="2012-10-17",
                    Statement=[
                        Statement(
                            Action=["s3:PutObject"],
                            Effect="Allow",
                            Resource=[
                                Sub("${Bucket}/*", Bucket=source_bucket.GetArn())
                            ],
                        ),
                    ],
                ),
            ),
        ],
    )

    # 15 second timeout; minimum memory size (128MB); TODO: DeadLetterConfig?
    function = Function(
        FunctionName=Sub("${AWS::StackName}-persist-analytics-event"),
        Role=function_role.GetArn(),
        MemorySize=128,  # 128 is the minimum
        Runtime="python3.6",
        Timeout=15,  # 15 second timeout to be safe
        Handler="index.handler",
        Environment=Environment(Variables={"BUCKET": source_bucket}),
        Code=Code(
            ZipFile="""import json
from datetime import datetime
import os

import boto3

def zeropad(i: int, digits: int = 2) -> str:
    if i < (10 ** (digits-1)):
        return f"0{i}"
    return str(i)

def handler(event, context):
    print("Receiving event:", json.dumps(event))
    now = datetime.utcnow()
    bucket = os.environ["BUCKET"]
    key=(
        f"{now.year}/{zeropad(now.month)}/{zeropad(now.day)}/"
        f"{zeropad(now.hour)}:{zeropad(now.minute)}:{zeropad(now.second)}."
        f"{zeropad(now.microsecond, 6)}"
    )
    print(f"Inserting data into s3://{bucket}/{key}")
    # NOTE: Taking care to make sure the querystring parameters don't override
    # the other event values, since qstring parameters are arbitrary and can
    # easily be faked by a malicious client.
    data = event.get("queryStringParameters", {})
    data.update({
        "user_agent": event["requestContext"]["http"]["userAgent"],
        "source_ip": event["requestContext"]["http"]["sourceIp"],
        "time": now.isoformat(),
    })
    data = json.dumps(data)
    print("Data: ", data)
    boto3.client("s3").put_object(Bucket=bucket, Key=key, Body=data)
    return {
        "isBase64Encoded": False,
        "statusCode": 200,
        "headers": {},
        "multiValueHeaders": {},
        "body": ""
    }""",
        ),
    )

    gateway_role = Role(
        RoleName=Sub("${AWS::StackName}-invoke-lambda"),
        Description="API Gateway role to invoke the analytics lambda function",
        AssumeRolePolicyDocument=PolicyDocument(
            Version="2012-10-17",
            Statement=[
                Statement(
                    Effect="Allow",
                    Principal=Principal(Service=["apigateway.amazonaws.com"]),
                    Action=["sts:AssumeRole"],
                )
            ],
        ),
        Policies=[
            Policy(
                PolicyName=Sub("${AWS::StackName}-invoke-lambda"),
                PolicyDocument=PolicyDocument(
                    Version="2012-10-17",
                    Statement=[
                        Statement(
                            Effect="Allow",
                            Action="lambda:InvokeFunction",
                            Resource=[function.GetArn()],
                        ),
                    ],
                ),
            )
        ],
    )

    api = Api(
        Name=Sub("${AWS::StackName}"),
        Description="The HTTP API for the analytics backend",
        ProtocolType="HTTP",
        Target=function.GetArn(),
        CredentialsArn=gateway_role.GetArn(),
    )

    database = Database(
        CatalogId=AccountID,
        DatabaseInput=DatabaseInput(
            Description=Sub("Database for web analytics for ${AWS::StackName} stack.",),
        ),
    )

    crawler = Crawler(
        Name=database,  # name the crawler the same as the database.
        # Crawl the whole bucket
        Targets=Targets(S3Targets=[S3Target(Path=source_bucket)]),
        SchemaChangePolicy=SchemaChangePolicy(
            UpdateBehavior="UPDATE_IN_DATABASE", DeleteBehavior="DELETE_FROM_DATABASE"
        ),
        Schedule=Schedule(ScheduleExpression=""),  # TODO
    )

    t = Template(
        description="Analytics backend",
        parameters={},
        resources={
            "SourceBucket": source_bucket,
            "FunctionRole": function_role,
            "Function": function,
            "GatewayRole": gateway_role,
            "API": api,
            "Database": database,
            "Crawler": crawler,
        },
    )
    print(json.dumps(t.template_to_cloudformation(), indent=4))
