---
Title: Securely expose private service for cheap
Date: 2021-10-29
Tags: ['meta', 'homelab']
---

_Disclaimer: This is not a production-grade solution_

At home I have a Raspberry Pi Kubernetes cluster running, among other things,
this blog (or at least at the time of this writing). One of my goals for this
cluster is to keep my cloud/SaaS/etc costs down below $5/month. Another goal is
to avoid poking holes in my home router's firewall.

<!-- more -->

A much more robust solution would be [inlets][0], but the price is a bit higher
than my budget goal allows. Instead, I'm running an ec2 spot instance
(t4g-nano) with a public IP address, DNS names, etc for a total of
~$2.25/month. From there, a node in my local cluster initiates a reverse SSH
tunnel with the gateway ec2 instance--a reverse tunnel simply means that
traffic flows _to_ the initiator, and the cluster (rather than the gateway)
initiates because my router's firewall permits all egress traffic but denies
all ingress traffic, which is to say that the gateway couldn't initiate a
tunnel into my cluster because it would be stopped by the firewall.

## Tunnel Deployment

On the cluster side, the tunnel is managed by a Kubernetes Deployment that
looks like this:

<details>
<summary>Tunnel Deployment</summary>

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tunnel
  namespace: ingress-nginx
  labels:
    app: tunnel
spec:
  replicas: 1
  selector:
    matchLabels:
      app: tunnel
  template:
    metadata:
      labels:
        app: tunnel
    spec:
      hostNetwork: true
      volumes:
      - name: tunnel-private-key
        secret:
          secretName: tunnel-private-key
          defaultMode: 0400
          items:
            - key: tunnel
              path: tunnel
              mode: 0400
      containers:
        - name: tunnel
          image: ubuntu
          command:
            - bash
            - "-c"
            - |
                apt update && apt install -y autossh
                autossh -N \
                    -i /private-key/tunnel \
                    -o "StrictHostKeyChecking no" \
                    -R 8080:localhost:80 tunnel@api.weberc2.com \
                    -R 8443:localhost:443 tunnel@api.weberc2.com
          volumeMounts:
            - name: tunnel-private-key
              mountPath: /private-key

```

</details>

Some things of note:

1. There's a private key for initiating the tunnel with the gateway. The
   gateway has an `~/.ssh/authorized_keys` entry containing the public key that
   corresponds to this private key. This key must be mounted read-only or the
   SSH client will rightly complain for security reasons.
2. I'm using `autossh` instead of the vanilla `ssh` client. I don't know why,
   but SSH tunnels are flaky things and `autossh` will restart broken tunnels.
   Perhaps you're thinking, as I did initially, that Kubernetes would restart
   the broken tunnel, but (1) broken tunnels don't actually terminate the
   client process, so as far as k8s is concerned everything is hunky dory and
   (2) it's probably faster for autossh to restart the tunnel rather than
   Kubernetes spinning up an entirely new pod.
3. api.weberc2.com points to the gateway instance
4. I'm forwarding traffic from api.weberc2.com:8443 to localhost:443 and from
   gateway:8080 to localhost:80. Ideally the tunnel could bind to ports 80 and
   443 on the remote, but the OpenSSH server (running on the gateway) doesn't
   allow binding to privileged ports. To work around, I'm running a process
   on the gateway that forwards 80->8080 and 443->8443 (details to follow).
5. `hostNetwork: true`--the pod binds to the node's network rather than the
   pod's virtual network. This is important for getting traffic from the tunnel
   pod to the ingress controller (to be routed to the target services).
6. `replicas: 1`--I only want one tunnel running, which is among the reasons
    that this is not a production-grade setup (no redundancy).

## Ingress Controller

For an ingress controller, I'm using [ingress-nginx][1].  To get traffic from
the `tunnel` deployment into the ingress controller, I had to make a couple of
changes to the default templates:

1. Change the `Deployment` to a `DaemonSet`. Specifically, we want to run this
   on every node, so no matter where the `tunnel` pod is scheduled, it will be
   on the same host as an ingress controller.
2. Configure the `DaemonSet` to listen on the host network
   (`DaemonSet.spec.template.spec.hostNetwork: true`). Since the `tunnel` pod
   drops traffic on the host network, the ingress controller has to listen on
   the host network to pick it up.
3. Change the `ingress-nginx` controller `Service` type to `NodePort`, set the
   ports to 80 and 443. These are the ports on the host network that the
   `tunnel` pod drops traffic. Note that we're using ports 80 and 443 so that
   the services are available via private network, but the most important part
   for our purposes is that the port values match the port values for the
   `tunnel` deployment.

This covers the changes on the cluster side of the firewall.

## Gateway

The gateway and associated infrastructure are managed with Terraform. There are
3 main resources:

1. a security group: deny everything except for ports 80, 443, and 22 (http,
   https, and ssh respectively). Port 22 allows inbound SSH connections from
   our cluster (or my laptop, for debugging).
2. an elastic ip: a public IP address that we will associate with the ec2
   instance.
3. a "spot instance request" which is effectively the ec2 instance itself.

There's also the private key Kubernetes `Secret`, which is just an SSH private
key (e.g., `ssh-keygen -t rsa ...`); however, this is created manually rather
than managed by Terraform (in retrospect there's probably a way to
generate/manage this with Terraform as well).

<details>
<summary>main.tf</summary>

```hcl
locals {
  key_name = "weberc2-ec2-key-pair"
}

# Find an official ubuntu AMI for ARM64
data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-arm64-server-*"]
  }

  filter {
    name   = "architecture"
    values = ["arm64"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_security_group" "gateway" {
  name        = "gateway"
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

resource "aws_eip" "gateway" {
  instance = aws_spot_instance_request.gateway.spot_instance_id
}

resource "aws_spot_instance_request" "gateway" {
  ami                  = data.aws_ami.ubuntu.id
  instance_type        = "t4g.nano"
  key_name             = local.key_name
  security_groups      = [aws_security_group.gateway.name]
  user_data            = file("./user-data.yaml")
  wait_for_fulfillment = true
}

# This zone was created manually, so we're referencing it rather than managing
# it directly in this Terraform project.
data "aws_route53_zone" "primary" {
  name = "weberc2.com."
}

resource "aws_route53_record" "gateway" {
  zone_id = data.aws_route53_zone.primary.zone_id
  name    = "api.weberc2.com"
  type    = "A"
  ttl     = "120"
  records = [aws_eip.gateway.public_ip]
}
```

</details>

## Conclusion

To reiterate, this is _not_ a production solution, but frankly I'm surprised at
how well it has worked. AWS hasn't nuked my spot instance (to my knowledge),
and the tunnel hasn't had any issues. As always, if you're reading this and you
have questions about how I implemented this, feel free to reach out. See the
"Contact" section in the footer.


[0]: https://inlets.dev
[1]: https://github.com/kubernetes/ingress-nginx
