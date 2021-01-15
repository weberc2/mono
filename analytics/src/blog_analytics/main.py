import json

from nimbus_core import ParameterString, Sub, Template
from nimbus_resources.apigatewayv2.api import Api
from nimbus_resources.iam.role import Role, Policy
from nimbus_resources.lambda_.function import Function, Code, Environment
from nimbus_resources.s3.bucket import (
    Bucket,
    BucketEncryption,
    ServerSideEncryptionRule,
    ServerSideEncryptionByDefault as Encryption,
)
from nimbus_resources.secretsmanager.secret import Secret
from nimbus_util.iam import PolicyDocument, Statement, Principal


def main():
    source_bucket = Bucket(
        BucketName=Sub("${AWS::AccountId}-${AWS::StackName}-analytics"),
        BucketEncryption=BucketEncryption(
            ServerSideEncryptionConfiguration=[
                ServerSideEncryptionRule(
                    ServerSideEncryptionByDefault=Encryption(
                        SSEAlgorithm="AES256",
                    ),
                )
            ],
        ),
    )

    ipstack_api_key_parameter = ParameterString(
        NoEcho=True,
        Description="The API key for the IPStack IP geolookup service",
    )

    secret = Secret(
        Name=Sub("${AWS::StackName}-secret"),
        Description="The secret for the analytics lambda function",
        SecretString=ipstack_api_key_parameter,
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
                PolicyName=Sub("${AWS::StackName}-secret"),
                PolicyDocument=PolicyDocument(
                    Version="2012-10-17",
                    Statement=[
                        Statement(
                            Action="secretsmanager:GetSecretValue",
                            Effect="Allow",
                            Resource=secret,
                        ),
                    ],
                ),
            ),
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
                                Sub(
                                    "${Bucket}/*",
                                    Bucket=source_bucket.GetArn(),
                                )
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
        Environment=Environment(Variables={
            "BUCKET": source_bucket,
            "SECRET": secret,
        }),
        Code=Code(
            ZipFile="""import json
from datetime import datetime
import os
import urllib3
import logging

import boto3

# Setup logging that works for both AWS lambda and local execution
logging.basicConfig(level = logging.INFO)
logger = logging.getLogger()


IPSTACK_API_KEY = boto3.client("secretsmanager").get_secret_value(
    SecretId=os.environ["SECRET"],
)["SecretString"]

HTTP = urllib3.PoolManager()


def handler(event, context):
    logger.info("Receiving event:", json.dumps(event))
    now = datetime.utcnow()
    bucket = os.environ["BUCKET"]
    key=(
        f"{now.year}/{zeropad(now.month)}/{zeropad(now.day)}/"
        f"{zeropad(now.hour)}:{zeropad(now.minute)}:{zeropad(now.second)}."
        f"{zeropad(now.microsecond, 6)}"
    )
    logger.info(f"Inserting data into s3://{bucket}/{key}")
    # NOTE: Taking care to make sure the querystring parameters don't override
    # the other event values, since qstring parameters are arbitrary and can
    # easily be faked by a malicious client.
    data = event.get("queryStringParameters", {})
    data.update({
        "user_agent": event["requestContext"]["http"]["userAgent"],
        "source_ip": event["requestContext"]["http"]["sourceIp"],
        "time": now.isoformat(),
    })
    data.update(geolookup(data["source_ip"]))
    data = json.dumps(data)
    logger.info("Data: ", data)
    boto3.client("s3").put_object(Bucket=bucket, Key=key, Body=data)
    return {
        "isBase64Encoded": False,
        "statusCode": 200,
        "headers": {},
        "multiValueHeaders": {},
        "body": ""
    }


def geolookup(ip_address):
    try:
        r = HTTP.request(
            "GET",
            f"http://api.ipstack.com/{ip_address}?access_key={IPSTACK_API_KEY}",
        )
        if r.status < 200 or r.status >= 300:
            raise Exception(f"Expected HTTP 2XX; got {r.status}")
        geo = json.loads(r.data)
        return {
            field: geo.get(field) for field in [
                "continent_code",
                "continent_name",
                "country_code",
                "country_name",
                "region_code",
                "region_name",
                "city",
                "zip",
                "latitude",
                "longitude",
            ]
        }
    except Exception as e:
        logger.error(f"Geolookup for '{ip_address}': {e}", exc_info=True)
        return None  # Explicitly return None in case of error


def zeropad(i: int, digits: int = 2) -> str:
    if i < (10 ** (digits-1)):
        return f"0{i}"
    return str(i)
""",
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

    t = Template(
        description="Analytics backend",
        parameters={
            "IPStackAPIKey": ipstack_api_key_parameter,
        },
        resources={
            "SourceBucket": source_bucket,
            "Secret": secret,
            "FunctionRole": function_role,
            "Function": function,
            "GatewayRole": gateway_role,
            "API": api,
        },
    )
    print(json.dumps(t.template_to_cloudformation(), indent=4))
