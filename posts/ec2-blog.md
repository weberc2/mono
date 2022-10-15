---
Title: Moving blog to EC2 Spot Instance
Date: 2022-10-15
Tags: ['aws', 'meta', 'homelab']
---

We recently moved from Chicago to Des Moines, and we're staying in an AirBnB
for a couple months while we look for the right house to buy. In the meanwhile,
most of our stuff (including critical components of my homelab) are in storage,
which means my blog wasn't running. In this transient period, I figured I would
try to run my blog in the cloud, and while there are easier and even cheaper
options, I decided to try out running it on EC2 in order to learn a bit more
about traditional Linux system administration. This post will document the
approach I arrived at.

<!-- more -->

My blog is just a static site served by Caddy, all packaged up into a Docker
image. The goal is to run my blog as well as logs and metrics exporters on a
single EC2 spot instance to keep costs down. Since this is just one instance,
if it goes down (and it likely will because it's a spot instance), my blog will
be unavailable, which is to say that this approach is *not* highly available
and thus not a good production architecture[^1].

Further, I intended to bind a static IP address directly to the spot instance
rather than proxying traffic through a $20/month (or whatever the cost is these
days) load balancer. This means that if my host goes down, a new instance will
be brought up to replace it; however, that instance will not automatically get
the Elastic IP address bound to it (the binding will die with the original
host). As far as I can tell, AWS doesn't offer any kind of automation for this
sort of scenario (apart from load balancers). This is a risk I'm willing to
accept, and if it becomes overly inconvenient, I could build a little lambda
function that periodically checks to make sure the elastic IP address is bound
properly and rebinds it if the spot instance goes down; however, in my
experience these kinds of interruptions are rare.

# Base Image

I decided to use Ubuntu for my base VM image, mostly because I'm more familiar
with it; however, as we'll see later Amazon Linux 2 would likely have been the
better choice. I'm also doing all of the configuration management[^2] through
[cloud-init][cloud-init] whereas I'm guessing more professional sysadmins would
just use cloud-init to bootstrap some other configuration management system
like Ansible. But I don't know Ansible or Puppet or etc, I'm not particularly
interested in learning them, and my use case seems simple enough.

# Process Management

A process manager is basically a controller whose job is to make sure that the
desired processes are running on the host. If a process terminates
unexpectedly, the process manager should respawn it. It is the main process on
a Linux system, and it spawns and controls all other processes. Since I picked
Ubuntu, I'm using the [systemd][systemd] process manager.

The application's systemd configuration is in a `blog.service` file:

<details>

```ini
[Unit]
Description=Caddy

[Service]
Restart=always
ExecStartPre=/usr/bin/docker pull weberc2/blog:latest
ExecStart=/usr/bin/docker run --rm -p 80:80 -p 443:443 weberc2/blog
StandardOutput=file:/var/log/blog.log
StandardError=file:/var/log/blog.log

# Restart every >5 seconds to avoid StartLimitInterval failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

</details>

The `ExecStartPre` line just makes sure the latest image is pulled (in case
the host has previously pulled an older version). Since I wrote this, it seems
the same is achievable with [`docker run ... --pull=always`][pull-always] such
that this `ExecStartPre` line could be elided (I'll probably make this change
in the future). This is necessary because for the time being, I'm just
deploying the latest version of my blog image, which means deploying my blog
just means publishing a new instance and restarting the application; however,
in the future I may deploy explicit image tags rather than using `latest`.

The `RestartSec=5` bit is necessary to keep the service from failing. I don't
understand why, and it's annoying that I have to add this, but if Linux tools
were intuitive then just anyone could run their own instances 🙃.

The `StandardOutput` and `StandardError` blocks are important--they write log
output to a `/var/log/blog.log` file, which will be read by our log exporter
as discussed below. Ideally our log exporter would be able to pull them
directly from `journald` (the logging complement to systemd), but the agent I
chose seems not to support that. There might be a better way to export logs
than writing them to a file, but I couldn't figure it out for my log exporter.

# Logging and Monitoring

I'm only running one host at a time, but I still want a better and more durable
way to access logs and metrics than SSH-ing onto the host and grepping files.
Specifically, if the host goes down, I don't want to lose its logs and metrics.
This means running exporter utilities for shipping the logs and metrics to some
other service where they can be analyzed. The default log/metric analysis tool
in the AWS world is CloudWatch, and I'd rather use a managed service than try
to operate my own (in my limited experience, CloudWatch seems much better than
the self-hosted options anyway). This means running the
`amazon-cloudwatch-agent` utility on the host in addition to the application.

Since I'm using Ubuntu rather than Amazon Linux 2, the agent's installation
package isn't available in the system repository, so I needed to write a small
script to download, install, and start the package:

<details>

```bash
#!/bin/bash

set -eo pipefail

agentDebPath=/amazon-cloudwatch-agent.deb
agentInstalledPath=/agent-installed

if [[ ! -f "$agentInstalledPath" ]]; then
    echo "cloudwatch agent hasn't been installed yet; installing..."
    echo "curling debian package..."
    curl \
        -L https://s3.amazonaws.com/amazoncloudwatch-agent/ubuntu/arm64/latest/amazon-cloudwatch-agent.deb \
        -o "$agentDebPath"
    echo "installing debian package..."
    sudo dpkg -i -E "$agentDebPath"
    echo "touching agent-installed flag file..."
    touch "$agentInstalledPath"
    echo "cloudwatch agent installed successfully"
fi

echo "starting the agent"
sudo amazon-cloudwatch-agent-ctl \
    -a fetch-config \
    -m ec2 \
    -c file:/amazon-cloudwatch-agent.json \
    -s
```

</details>

Note that I'm using the `arm64` version of the package, because my spot
instance is based on the cheaper (better performance/cost) ARM instances. The
debian package handles installing the systemd configuration for the agent, so
I don't have to write or install my own systemd `.service` file.

The configuration for the agent is attached below. Note that I'm collecting the
syslog as well as the aforementioned `/var/log/blog.log` file, and running as
the root user--I could probably run as the `cwagent` user, but I'd need to find
a way to grant that user read access to the syslog. Note also that I'm
collecting disk, memory, and CPU metrics.

<details>

```json
{
  "agent": {
    "debug": false,
    "logfile": "/opt/aws/amazon-cloudwatch-agent/logs/amazon-cloudwatch-agent.log",
    "metrics_collection_interval": 10,
    "run_as_user": "root"
  },
  "logs": {
    "logs_collected": {
      "files": {
        "collect_list": [
          {
            "file_path": "/var/log/syslog",
            "log_group_name": "blog",
            "log_stream_name": "/var/log/syslog - {hostname}",
            "retention_in_days": 60,
            "timezone": "UTC"
          },
          {
            "file_path": "/var/log/blog.log",
            "log_group_name": "blog",
            "log_stream_name": "/var/log/blog.log - {hostname}",
            "retention_in_days": 60,
            "timezone": "UTC"
          }
        ]
      }
    }
  },
  "metrics": {
    "aggregation_dimensions": [
      [
        "InstanceId"
      ]
    ],
    "append_dimensions": {
      "ImageId": "${aws:ImageId}",
      "InstanceId": "${aws:InstanceId}",
      "InstanceType": "${aws:InstanceType}"
    },
    "metrics_collected": {
      "cpu": {
        "measurement": [
          "usage_active",
          "usage_system",
          "usage_user"
        ],
        "metrics_collection_interval": 60
      },
      "disk": {
        "measurement": [
          "used_percent"
        ],
        "metrics_collection_interval": 60,
        "resources": [
          "*"
        ]
      },
      "mem": {
        "measurement": [
          "mem_used_percent"
        ],
        "metrics_collection_interval": 60
      }
    }
  }
}
```

</details>

As we will see, I'm going to ship this script and the agent installation file
in the cloud-init user-data file. CloudWatch also requires configuring a role
for the instance giving it permissions to write to CloudWatch--see the section
on infrastructure below.

# User Data

When an instance boots for the first time, `cloud-init` runs, and one of its
first activities is to find the "user-data", which is the configuration file
provided by the cloud provider for configuring that instance. In the case of
AWS, cloud-init calls the [metadata endpoint][metadata-endpoint] to download
this file. This file is so-called because it's something that we (users of AWS)
provide when we request an AWS EC2 instance. Mine looks like this (note this
YAML was generated by Terraform--more on that below--hence the formatting and
order of keys):

<details>

```yaml
#cloud-config
"packages":
- "curl"
- "docker.io"
"runcmd":
- - "/usr/bin/install-cloudwatch-agent.sh"
- - "systemctl"
  - "start"
  - "blog"
"ssh_import_id":
- "gh:weberc2"
"ssh_pwauth": false
"users":
- "groups": "sudo, users, admin"
  "lock_passwd": true
  "name": "weberc2"
  "shell": "/bin/bash"
  "ssh_import_id":
  - "gh:weberc2"
  "ssh_pwauth": false
  "sudo": "ALL=(ALL) NOPASSWD:ALL"
"write_files":
- "content": |
    [Unit]
    Description=Caddy

    [Service]
    Restart=always
    ExecStartPre=/usr/bin/docker pull weberc2/blog:latest
    ExecStart=/usr/bin/docker run --rm -p 80:80 -p 443:443 weberc2/blog
    StandardOutput=file:/var/log/blog.log
    StandardError=file:/var/log/blog.log

    # Restart every >5 seconds to avoid StartLimitInterval failure
    RestartSec=5

    [Install]
    WantedBy=multi-user.target
  "path": "/etc/systemd/system/blog.service"
- "content": "{\"agent\":{\"debug\":false,\"logfile\":\"/opt/aws/amazon-cloudwatch-agent/logs/amazon-cloudwatch-agent.log\",\"metrics_collection_interval\":10,\"run_as_user\":\"root\"},\"logs\":{\"logs_collected\":{\"files\":{\"collect_list\":[{\"file_path\":\"/var/log/syslog\",\"log_group_name\":\"blog\",\"log_stream_name\":\"/var/log/syslog
    - {hostname}\",\"retention_in_days\":60,\"timezone\":\"UTC\"},{\"file_path\":\"/var/log/blog.log\",\"log_group_name\":\"blog\",\"log_stream_name\":\"/var/log/blog.log
    - {hostname}\",\"retention_in_days\":60,\"timezone\":\"UTC\"}]}}},\"metrics\":{\"aggregation_dimensions\":[[\"InstanceId\"]],\"append_dimensions\":{\"ImageId\":\"${aws:ImageId}\",\"InstanceId\":\"${aws:InstanceId}\",\"InstanceType\":\"${aws:InstanceType}\"},\"metrics_collected\":{\"cpu\":{\"measurement\":[\"usage_active\",\"usage_system\",\"usage_user\"],\"metrics_collection_interval\":60},\"disk\":{\"measurement\":[\"used_percent\"],\"metrics_collection_interval\":60,\"resources\":[\"*\"]},\"mem\":{\"measurement\":[\"mem_used_percent\"],\"metrics_collection_interval\":60}}}}"
  "path": "/amazon-cloudwatch-agent.json"
- "content": |
    #!/bin/bash

    set -eo pipefail

    agentDebPath=/amazon-cloudwatch-agent.deb
    agentInstalledPath=/agent-installed

    if [[ ! -f "$agentInstalledPath" ]]; then
        echo "cloudwatch agent hasn't been installed yet; installing..."
        echo "curling debian package..."
        curl \
            -L https://s3.amazonaws.com/amazoncloudwatch-agent/ubuntu/arm64/latest/amazon-cloudwatch-agent.deb \
            -o "$agentDebPath"
        echo "installing debian package..."
        sudo dpkg -i -E "$agentDebPath"
        echo "touching agent-installed flag file..."
        touch "$agentInstalledPath"
        echo "cloudwatch agent installed successfully"
    fi

    echo "starting the agent"
    sudo amazon-cloudwatch-agent-ctl \
        -a fetch-config \
        -m ec2 \
        -c file:/amazon-cloudwatch-agent.json \
        -s
  "path": "/usr/bin/install-cloudwatch-agent.sh"
  "permissions": "0744"
```

</details>

Note that the `write_files` section contains the aforementioned
`blog.service`, `install-cloudwatch-agent.sh`, and
`amazon-cloudwatch-agent.json` files, and where to write them to disk. It also
specifies which packages are to be installed from the system package
repository--in this case, I'm installing Docker and curl (required by the
application systemd unit and the cloudwatch agent installation script,
respectively).

The user-data also contains some stuff for configuring users and
how they can SSH onto the instance (notably password authentication is
disabled, my weberc2 user is a sudoer, and my user's pubkey is pulled from
GitHub).

Lastly, the `runcmd` section tells `cloud-init` to start the blog application
and run the cloudwatch agent installation script.

# Infrastructure

A major goal for any project I work on is that the infrastructure is immutable
and reproducible. I should be able to tear down my infrastructure and stand it
back up again with relative ease. I certainly don't want "a human pokes around
the AWS console or SSHes onto an instance" to be part of the process, because
I will make mistakes, and I want to minimize tedium (this is for fun, after
all). To that end, I'm using Terraform to describe my infrastructure and
reconcile that description with the current state.

The infrastructure contains a security group for allowing traffic to reach the
instance (ports 80, 443, and 22 for HTTP, HTTPS, and SSH traffic,
respectively).

<details>

```hcl
resource "aws_security_group" "self" {
  name        = var.aws_resource_prefix
  description = "Allow HTTP, TLS, and SSH inbound traffic"

  ingress {
    description      = "TLS"
    from_port        = 443
    to_port          = 443
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  ingress {
    description      = "HTTP"
    from_port        = 80
    to_port          = 80
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  # TODO: use tailscale for SSH
  ingress {
    description      = "SSH"
    from_port        = 22
    to_port          = 22
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}
```

</details>

It also defines the Elastic IP address and the Route53 DNS record descriptions:

<details>

```hcl
resource "aws_eip" "self" {
  instance = aws_spot_instance_request.self.spot_instance_id
}

data "aws_route53_zone" "primary" {
  name = "weberc2.com."
}

resource "aws_route53_record" "self" {
  zone_id = data.aws_route53_zone.primary.zone_id
  name    = "${coalesce(var.subdomain, var.app_name)}.weberc2.com"
  type    = "A"
  ttl     = "120"
  records = [aws_eip.self.public_ip]
}
```

</details>

And the IAM stuff for granting the instance permissions to send logs and
metrics to CloudWatch:

<details>

```hcl
resource "aws_iam_role" "self" {
  name = "${var.aws_resource_prefix}_instance"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.amazonaws.com"
      }
    }]
  })

  inline_policy {
    name = "write-logs"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [{
        Effect = "Allow",
        Action = [
          "cloudwatch:PutMetricData",
          "ec2:DescribeTags",
          "logs:PutLogEvents",
          "logs:PutRetentionPolicy",
          "logs:DescribeLogStreams",
          "logs:DescribeLogGroups",
          "logs:CreateLogStream",
          "logs:CreateLogGroup",
        ],
        Resource = "*"
      }]
    })
  }
}

resource "aws_iam_instance_profile" "self" {
  name = var.aws_resource_prefix
  role = aws_iam_role.self.name
}
```

</details>

Lastly, it contains the description of the spot instance itself:

<details>

```hcl

resource "aws_spot_instance_request" "self" {
  ami                  = data.aws_ami.ubuntu.id
  instance_type        = "t4g.nano"
  key_name             = var.key_name
  security_groups      = [aws_security_group.self.name]
  user_data            = local.user_data
  wait_for_fulfillment = true
  iam_instance_profile = aws_iam_instance_profile.self.id
}
```

</details>

Note the `user_data` is set to `local.user_data`. Previously, this was just a
reference to a local `user-data.yaml` file contained on disk (`user_data =
file("user-data.yaml")`), but as we will see in the next section, I've
abstracted the user-data to make this Terraform more flexible for applications
besides my blog.

## Abstracting User-Data

The user-data referenced above is only suitable for my blog, but conceivably I
may want to configure other instances running a different suite of
applications. Indeed, I have another instance that I was also running with its
own hard-coded user-data, and I wanted to abstract out the similarities. To
that end, I created a module (Terraform verbiage for a template) and factored
out the bits that differ into parameters ("input variables" in Terraform
parlance):

<details>

```hcl
variable "services" {
  type = list(object({
    name           = string
    description    = string
    after          = string
    exec_start_pre = string
    exec_start     = string
    packages       = list(string)
  }))
  description = "Descriptions of each service to run on the machine"
}
```

</details>

This allows me to stamp out multiple Spot Instances complete with DNS and
logging/metrics support by just passing in a few parameters of information
about each service that will run on the instance (excluding the logging agent,
which is provided by default). Further, it gives me one place to make changes
which can then benefit all of my spot instances.

This `services` variable is used to dynamically generate the user-data
and the cloudwatch agent configuration files:

<details>

```hcl
locals {
  user_data = join(
    "\n",
    [
      "#cloud-config",
      yamlencode({
        ssh_import_id = ["gh:weberc2"]
        ssh_pwauth    = false
        users = [
          {
            name          = "weberc2"
            sudo          = "ALL=(ALL) NOPASSWD:ALL"
            groups        = "sudo, users, admin"
            shell         = "/bin/bash"
            lock_passwd   = true
            ssh_pwauth    = false
            ssh_import_id = ["gh:weberc2"]
          },
        ]

        packages = distinct(concat(
          ["curl"],
          flatten([for x in var.services : x.packages]),
        ))

        write_files = concat(
          [
            for x in var.services :
            {
              path    = "/etc/systemd/system/${x.name}.service"
              content = <<EOF
[Unit]
Description=${x.description}
%{if x.after != ""}After=${x.after}%{endif}

[Service]
Restart=always
%{if x.exec_start_pre != ""}ExecStartPre=${x.exec_start_pre}%{endif}
ExecStart=${x.exec_start}
StandardOutput=file:/var/log/${x.name}.log
StandardError=file:/var/log/${x.name}.log

# Restart every >5 seconds to avoid StartLimitInterval failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
            }
          ],
          [
            {
              path = "/amazon-cloudwatch-agent.json"
              content = jsonencode({
                agent = {
                  metrics_collection_interval = 10
                  logfile                     = "/opt/aws/amazon-cloudwatch-agent/logs/amazon-cloudwatch-agent.log"
                  debug                       = false
                  run_as_user                 = "root" # necessary for reading syslog
                }
                logs = {
                  logs_collected = {
                    files = {
                      collect_list = concat(
                        [{
                          file_path         = "/var/log/syslog"
                          timezone          = "UTC"
                          retention_in_days = 60
                          log_group_name    = var.app_name
                          log_stream_name   = "/var/log/syslog - {hostname}"
                        }],
                        [
                          for x in var.services :
                          {
                            file_path         = "/var/log/${x.name}.log"
                            timezone          = "UTC"
                            retention_in_days = 60
                            log_group_name    = var.app_name
                            log_stream_name   = "/var/log/${x.name}.log - {hostname}"
                          }
                        ],
                      )
                    }
                  }
                }
                metrics = {
                  aggregation_dimensions = [["InstanceId"]]
                  append_dimensions = {
                    ImageId      = "$${aws:ImageId}"
                    InstanceId   = "$${aws:InstanceId}"
                    InstanceType = "$${aws:InstanceType}"
                  }
                  metrics_collected = {
                    disk = {
                      measurement                 = ["used_percent"]
                      metrics_collection_interval = 60
                      resources                   = ["*"]
                    }
                    mem = {
                      measurement                 = ["mem_used_percent"]
                      metrics_collection_interval = 60
                    }
                    cpu = {
                      measurement = [
                        "usage_active",
                        "usage_system",
                        "usage_user",
                      ]
                      metrics_collection_interval = 60
                    }
                  }
                }
              }),
            },
            {
              path        = "/usr/bin/install-cloudwatch-agent.sh"
              permissions = "0744"
              content     = <<EOF
#!/bin/bash

set -eo pipefail

agentDebPath=/amazon-cloudwatch-agent.deb
agentInstalledPath=/agent-installed

if [[ ! -f "$agentInstalledPath" ]]; then
    echo "cloudwatch agent hasn't been installed yet; installing..."
    echo "curling debian package..."
    curl \
        -L https://s3.amazonaws.com/amazoncloudwatch-agent/ubuntu/arm64/latest/amazon-cloudwatch-agent.deb \
        -o "$agentDebPath"
    echo "installing debian package..."
    sudo dpkg -i -E "$agentDebPath"
    echo "touching agent-installed flag file..."
    touch "$agentInstalledPath"
    echo "cloudwatch agent installed successfully"
fi

echo "starting the agent"
sudo amazon-cloudwatch-agent-ctl \
    -a fetch-config \
    -m ec2 \
    -c file:/amazon-cloudwatch-agent.json \
    -s
EOF
            },
          ]
        )

        runcmd = concat(
          ["/usr/bin/install-cloudwatch-agent.sh"],
          [for x in var.services : ["systemctl", "start", x.name]],
        )
      })
    ]
  )
}
```

</details>

Since I've modularized my concept of a spot instance application, I've also
ported my older EC2 instance to it so that it can benefit from the logging
changes.

# Next Steps

Now that I have my EC2 spot instance module, the next thing I'd like to add to
it is some sort of alerting automation so I can detect when it goes down (at a
minimum) and possibly if it looks like it's about to run out of resources
(e.g., excessive use of CPU, memory, or disk) although I'm not very worried
because Caddy is pretty simple/reliable. Mostly, my biggest concern is that AWS
will temporarily take down my spot instance, and I'll need to manually
re-associate the Elastic IP address with the new instance it brings up--and an
automated alert means I can know about it immediately rather than needing to
periodically check the blog myself.

Another improvement would be to use TailScale (a VPN) for SSH access rather
than exposing port 22 to the public Internet.

At some point, I would also like to build a CI job to automatically apply my
Terraform changes, because at the moment my process is just running these
Terraform applies on my local laptop (which also holds the Terraform state).
This works well enough for my single-developer flow, and if I'm concerned about
losing the Terraform state, Terraform has first class support for writing it to
a cloud data store (which is also a precursor for building a CI job).

[cloud-init]: https://cloudinit.readthedocs.io/en/latest/
[systemd]: https://www.freedesktop.org/wiki/Software/systemd/
[pull-always]: https://stackoverflow.com/a/57587374/483347
[metadata-endpoint]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html

[^1]: This could be remedied easily enough by running multiple hosts behind a
  load balancer.

[^2]: For some reason, all of the documentation I'd seen in the past for
  "configuration management" seems to imply that the term is self-describing
  or that everyone already knows what it means, but as far as I can tell, it
  specifically refers to preparing a host (i.e., installing packages,
  configuration files, etc) at boot time to run the application. So we're
  specifically talking about hosts (as opposed to all of the other stuff we
  configure in the infrastructure world), and for some reason it seems more
  common to do this at boot time rather than baking it into the VM image as we
  do in the container world (presumably because VM image building tooling is
  even worse than Docker image building tooling).
