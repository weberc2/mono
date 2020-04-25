from nimbus_core import Template, Sub
from nimbus_resources.sqs.queue import Queue
from nimbus_resources.apigateway.restapi import RestApi
from nimbus_resources.apigateway.method import (
    Method,
    MethodResponse,
    Integration,
    IntegrationResponse,
)
from nimbus_resources.iam.role import Role, Policy
from nimbus_util.iam import PolicyDocument, Statement, Principal


def main():
    queue = Queue(
        FifoQueue=True,
        QueueName=Sub("${AWS::StackName}.fifo"),
        ContentBasedDeduplication=True,
    )

    api = RestApi(
        Name=Sub("${AWS::StackName}-analytics-pipeline"),
        Description="The SQS proxy for the analytics pipeline",
    )

    gateway_role = Role(
        AssumeRolePolicyDocument=PolicyDocument(
            Version="2012-10-17",
            Statement=[
                Statement(
                    Action=["sts:AssumeRole"],
                    Effect="Allow",
                    Principal=Principal(Service=["apigateway.amazonaws.com"]),
                )
            ],
        ),
        Path="/",
        Policies=[
            Policy(
                PolicyName=Sub("${AWS::StackName}-apigateway-sqs-send-msg"),
                PolicyDocument=PolicyDocument(
                    Version="2012-10-17",
                    Statement=[
                        Statement(
                            Action=["sqs:SendMessage"],
                            Effect="Allow",
                            Resource=queue.GetArn(),
                        ),
                        Statement(
                            Action=[
                                "logs:CreateLogGroup",
                                "logs:CreateLogStream",
                                "logs:PutLogEvents",
                            ],
                            Effect="Allow",
                            Resource="*",
                        ),
                    ],
                ),
            ),
        ],
    )

    t = Template(
        description="Analytics pipeline",
        parameters={},
        resources={
            "Queue": queue,
            "RestAPI": api,
            "GatewayRole": gateway_role,
            "SQSMethod": Method(
                AuthorizationType="NONE",
                HttpMethod="POST",
                Integration=Integration(
                    Credentials=gateway_role.GetArn(),
                    IntegrationHttpMethod="POST",
                    IntegrationResponses=[IntegrationResponse(StatusCode="200")],
                    PassthroughBehavior="NEVER",
                    RequestParameters={
                        "integration.request.header.Content-Type": (
                            "'application/x-www-form-urlencoded'"
                        ),
                    },
                    RequestTemplates={
                        "application/json": (
                            "Action=SendMessage&MessageGroupId=foo&MessageBody=$input.body"
                        ),
                    },
                    Type="AWS",
                    Uri=Sub(
                        (
                            "arn:aws:apigateway:${AWS::Region}:sqs:path/"
                            "${AWS::AccountId}/${Queue}"
                        ),
                        Queue=queue.GetQueueName(),
                    ),
                ),
                MethodResponses=[
                    MethodResponse(
                        ResponseModels={"application/json": "Empty"}, StatusCode="200",
                    )
                ],
                ResourceId=api.GetRootResourceId(),
                RestApiId=api,
            ),
        },
    )
    print(json.dumps(t.template_to_cloudformation(), indent=4))
