---
Title: Weekend Updates
Date: 2021-01-16
---

I've had an unusually productive last few days with respect to my hobby
projects. Since Thursday evening, I published an initial draft of [Homelab Part
I: Hardware][0], did some more work on my Raspberry Pi cluster, added a
syndication and markdown table support to my static site generator. Lastly, I
found and fixed a viewport bug that was breaking making my blog's font size way
too big for iPad-sized devices. I also managed to play too much Battlefield 1,
but I'll stick with the technical details for now.

<!-- more -->

### iPad Viewport Bug

Jen got a new iPad this morning and over the course of playing around with it,
I figured out that this blog looked really bad. The font-sizes were all ~3x too
big. I'm not really sure why it only affected iPad, but the jist is that when I
was adding mobile support, I learned that I needed to add a `<meta
name=viewport ...>` tag to my HTML pages or else a smartphone screen would just
show a bunch of left-margin and the user would have to scroll to pan over the
content. I didn't really understand it, but I figured out that
`content="width=350"` did the trick. After spending this morning hunting for
the solution (not understanding at the time that this viewport thing was the
culprit and chasing down several red herrings) it became apparent that I needed
to set the width to a special `device-width` value. Now it looks great in iPad.


### Syndication

I've never been a big feeds guy personally, but I really like the idea of "the
old web" and smaller-yet-still-open communities of bloggers. Atom/RSS feeds
remind me of that time--even if they were never really pervasive, they fit the
aesthetic. So I bought [Reeder 5][1] for iOS and macOS (and probably for iPadOS
in the coming days), added a couple of feeds, and ... haven't really done
anything with them yet. But I did update my static site generator, `neon` so it
will automatically generate a `feed.atom` file at the root of the output
directory, and I updated the "Contact" section of the blog footer to link to
it.

I've been meaning to do this for years, but there hasn't been much point
because I've never really developed the habbit of writing regularly to this
blog anyway. I also kind of figured it would be more difficult than it was to
figure out how to fit this feature into the architecture of my static site
generator, but it worked out elegantly enough that I was able to crank it out
[soup-to-nuts][2] in just a few hours time (thanks largely to the excellent
[`gorilla/feeds`][3] package.

### Markdown Table

For my [Homelab/Hardware post][0], I wanted a table for my bill of materials.
My static site generator uses another fantastic library, [`blackfriday`][4],
for markdown support, and the library's extensibility is excellent and elegant.
I didn't expect it would be hard to add tables, but I guess I forgot *exactly
how easy* it was to extend:

```go
blackfriday.Run(
    // ...
    blackfriday.WithExtensions(
        blackfriday.CommonExtensions|
            blackfriday.Footnotes|
            blackfriday.Tables,
    ),
    // ...
)
```

With that done, I only had to style the tables in my theme's CSS file, and now
I have tables that don't look half bad. This is a pretty small feature, but it
brings me a lot of joy to see my home-grown tools evolve gradually over time.

### Raspberry Pi Cluster update

I'm presently working on the storage element of my cluster. I have a Pi that is
connected to an SSD, and I configured the host-setup script to add the SSD to
the /etc/fstab table (so the drive would auto-mount on boot). The goal is to
expose the SSD to the whole cluster as an NFS volume driver or perhaps
eventually as a [Longhorn.io][5] volume driver.

Unfortunately, I was running into a networking issue that seems to have been
[Tailscale][6]-related (either a bug or a PEBAK; haven't root caused it yet).
The issue was preventing all network egress outside of my LAN, including DNS
resolution failures. I had to uninstall Tailscale and restart the application
to get egress working on the node at all; I suspect the tailscale package
configures some iptables or something, and uninstalling the package deletes the
tailscale-related rules? The Tailscale team has always been super responsive to
my issues even though I'm not a paying customer, so I'm confident I'll be able
to get to the bottom of it when I have time to root-cause it.

Thankfully, all I had to do was remove the tailscale bits from my host
bootstrap scripts, reflash my SD cards, and rerun the scripts (and since I blew
away my Kubernetes master as well, I also needed to re-apply my various
manifests, but those are few and manageable by hand easily enough). These
scripts make me really happy and I'm glad I'm investing in them early and
often so I know I can always get my cluster back to a good state without having
to plod through a bunch of uninteresting detail that I haven't thought about in
months.

### Tailscale SWAG Box

Lastly, I saw that Tailscale was giving out hoodies and other SWAG, so [I
requested one via Twitter][7]. It finally arrived in the mail and fits great.
Now it's in the daily
trying-to-keep-our-heating-bill-low-and-save-the-planet-by-bundling-up
hoodie/sweater rotation (at least until Jen pilfers it, as she's known to do).

[0]: ./homelab-part-i-hardware.html
[1]: https://reederapp.com/
[2]: https://grammarist.com/idiom/from-soup-to-nuts/
[3]: https://github.com/gorilla/feeds
[4]: https://github.com/russross/blackfriday
[5]: https://longhorn.io
[6]: https;//tailscale.io
[7]: https://twitter.com/weberc2/status/1346578091536232453?s=20
