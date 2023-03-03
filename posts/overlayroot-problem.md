---
Title: Overlayroot Problem Statement
Date: 2023-03-02
---

This week I've been looking at `overlayroot` as a potential solution to reduce
the effort to make changes to the nodes in my Raspberry Pi cluster. In this
post I want to talk about the problem I'm hoping it solves and the problems
I'm running into with respect to implementing `overlayroot` as well as
potential solutions that I'm exploring.

<!-- more -->

# What am I hoping to solve with `overlayroot`?

For my Raspberry Pi cluster, I've been using cloud-init's NoCloud provider to
configure each node, which entails mounting my SD card to my Mac and writing
new user-data. This worked fine when changes were infrequent, but I'm pivoting
toward pushing more of the complexity into the user-data so I want to remove
the "burn an SD card image" from my iteration loop (as well as the incumbent
searching around for my SD card adapter).

There are a few solutions to this problem, but the one that seems most
appealing is using overlayroot to keep a read-only filesystem layer (the
"lower" layer in overlayfs terminology) that essentially preserves the initial
image as it was at install-time while writing changes to an "upper" layer. The
upper layer can then be erased in order to reset the system to its initial
state.

# Roadblocks & potential solutions

The problem I'm running into is that overlayroot seems to assume that these
layers are running on distinct block devices, and it seems overwhelmingly
people use an in-memory tmpfs for the upper layer (so much so that it seems to
be implicit when people write about overlayroot online). As far as I can tell
from the overlayroot documentation and Googling, I have these options:

1. Use a tmpfs upper layer
2. Partition my SD card such that I have a device for "upper" and "lower"
3. *Maybe* create a loop device from a file on a single primary partition

## Use a tmpfs upper layer

This option is the easiest to get working, but it seems like overlayroot will
use half of the available memory for the tmpfs which is a lot for a raspberry
pi (particularly my 1GB pis). As far as I can tell, there's no way to specify
an amount of memory before it starts swapping to disk, and I'm also not sure
how the performance of swapping a tmpfs to disk will compare to just writing to
a block device (moreover, I would think I still would need a large swap
partition anyway, so I don't think this ends up buying me anything over a
partitioned SD card).

## Partition SD card into "upper" and "lower" partitions

Basically this approach involves writing a Raspberry Pi disk image file to an
SD card and then creating a third "upper" partition with the remaining space.
This will give me devices that I can reference in the `overlayroot.conf` file,
but it's a pretty significant mount of work everytime I want to burn a new SD
card (there's probably some way to automatically grow the third partition, but
I'm not sure how to make that happen at the moment).

## *Maybe* create a loop device from a file on a single primary partition

I'm not sure if this is feasible for a few reasons, but the idea is that I
shouldn't have to create partitions just to get a separate block device. I
should be able to create a file on my primary partition, create a loop device
from that file, and reference that loop device from my `overlayroot.conf` file.
The main challenge here is that loop devices don't persist across boots, so I
would need to inject a script that runs *before* `overlayroot` which creates
the loop device, and I'm not sure when exactly `overlayroot` runs during the
boot process, nor how to hook in my script. I haven't found anything online
about using `overlayroot` with loop devices (a few things about using a DIY
overlayroot-like system with a loop device, but nothing about overlayroot
specifically).

# Conclusion

The `tmpfs` solution seems unworkable for my requirements (not using a bunch of
memory), and the loop device solution seems like a thread that I could pull for
a very long time before getting anything working (mostly because of my poor
understanding of the Linux boot process). The partitioned SD card solution
seems like the fastest path to a working solution, but it also requires a lot
of work each time I'm burning an SD card (any time I want to add a new Pi to
the cluster or replace an SD card).

As such, I'll start by getting the partitioned SD card working and see how that
works out in practice. Hopefully I won't be burning SD cards so often as to be
burdensome, and even if it is painful, I can probably automate away a fair
amount of that pain (scripting the SD card burning process).
