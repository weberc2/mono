---
Title: Update
Date: 2022-04-30
Tags: ['homelab', 'media', 'update']
---

I'm starting a new series where I briefly discuss what I've been working on,
what I've read, and what I'd like to explore. I'm just calling it "Update"
because I know I can't commit to any particular schedule. To find out what I've
been thinking about recently, read on.

<!-- more -->

# In progress

This week I've been focused on building a media server, specifically learning
about different encodings (and parameters they take), container formats,
subtitle formats, and how to rip DVDs. I've also learned about streaming media
formats, media server protocols (DLNA), and a bit about the landscape of media
servers and clients. It's been dizzying for what seems like a straightforward
problem, but I'm getting my head around it.

One of the big takeaways has been that live transcoding is computationally
intensive and I don't really have any hardware that's up to the task. Instead,
I've been transcoding ahead of time to formats that are agreeable to the
clients that I'll be using (Amazon Fire Stick, Roku, Samsung Smart TV, etc) so
that the media server doesn't need to transcode live. I haven't confirmed, but
I think all of my clients can handle h.265 encoding and mkv containers.

I'm putting all of my media into BackBlaze's b2 (because it's a quarter of the
cost of AWS S3) using rclone, and for the time being, I'm using `rclone serve
dlna` from a spare Raspberry Pi 3B to serve directly from b2. This has been
trivial to run and (almost?) all of my clients support DLNA. Moreover, I can
access my media from anywhere either my Tailscale VPN or frankly just by
running rclone from my laptop (the latter would be easier and more secure than
getting some AirBnB's smart TV to talk to a service on my VPN).

I still need to think about archiving my media (picking an archive format and
finding cheap durable storage--probably S3's Glacier or similar).

# To explore later

Some things I came across along the way that I'd like to explore in greater
detail eventually:

* [Ubuntu's MaaS](https://maas.io) ("metal as a service"), a system for
  managing bare-metal machine installations. This seems like it would scratch
  an itch I have for my Raspberry Pi fleet. Right now I have some hacky
  automation for setting up Raspberry Pi hosts and getting them connected to my
  Kubernetes cluster, but it leaves a lot to be desired. A more scalable,
  structured approach would be great.

* [Tailscale's services](https://tailscale.com/kb/1100/services/)--seems like
  Tailscale can be made aware of individual services rather than just hosts,
  which is something I've wanted for a while, although it remains to be seen if
  what Tailscale built is the thing I was desiring or not, hence "to explore
  later".

* [Helmfile](https://github.com/roboll/helmfile)--this seems like it aspires to
  provide a structured approach to putting your helm chart instantiations into
  source control. I haven't liked helm, in large part due to its abuse of text
  templates, but also because I didn't quite know how to GitOps my helm charts.
  Hopefully this tackles that last problem.

# Interesting reads

* [I've used all the notebooks][0] - a comparison of fancy notebooks. Taking
  notes and sketching ideas is a big part of my process, but I've rarely
  shelled out for anything more than the $0.10 notebooks I used in school as a
  child. I liked the post, and I added a couple of the recommendations to my
  shopping list.

* [Cool things people do with their blogs][1] - Brief list packed with creative
  ideas for blogging. Inspired me to write this post.


[0]: https://tylercipriani.com/blog/2022/04/30/ive-used-all-the-notebooks/
[1]: https://brainbaking.com/post/2022/04/cool-things-people-do-with-their-blogs/
