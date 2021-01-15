---
Title: On Python performance
Date: 2020-08-12
---

At this point in my career, I've been writing Python regularly for about 12
years. As a language, Python fills a fantastic niche: it's very dynamic and
relatively easy to get started with; however, many production applications will
run into performance problems very quickly for which there aren't palatable
solutions. Even still, there are enduring myths that these problems can be
waved away by sprinkling some Pandas on them, or using multiprocessing, or
rewriting in C/Cython/Rust/etc. In fact, I find myself addressing these claims
repeatedly in all manner of fora, and I think it's interesting enough to write
a blog post about.

<!-- more -->

## How slow is Python?

Questions of programming language performance are notoriously difficult to
answer specifically, not least of all because many languages have multiple
implementations and predants will point out that languages aren't fast or slow,
but *language implementations* are fast or slow. Further, the performance of a
language (implementation) depends heavily on the type of application and so on.

Still, there are pretty distinct "ballparks" in which programming languages
compete. The first tier of languages are the fastest: C, C++, Rust, and the
like. The second tier tend to be 2X-10X slower: Java, C#, Go, etc. The next
tier tend to be 50X-1000X slower than the **second** tier of languages, with
these being JavaScript, Ruby, Python, etc. Even in this third tier of slow
languages, JavaScript frequently beats Python because JS has enjoyed lots of
optimization (due to its monopoly on the browser market and also because it
didn't commit itself to the same restrictive design decisions that Python did).

Ultimately, in my entire career I haven't worked on a significant Python
product for which performance wasn't eventually a significant concern, and for
which the landscape of possible optimizations wasn't unpalatably expensive. In
many cases, these were SaaS applications with a 90th percentile response time
that exceeded the 60s HTTP timeout ceiling, and the only optimizations
available were rewriting a significant part of our application into a separate
service written in a more performant language or rewriting a significant part
of our application to leverage something like Apache Spark. In either case, the
rewrites were hugely expensive, and in one case it doomed the product.

## Why is Python slow?

First of all, if you squint, sometimes Python *isn't* slow. Python enthusiasts
will point out that many applications can take advantage of libraries which are
written in C (such as Pandas) to improve performance. In other cases, the
bottleneck is frequently in a database, not in a Python web application. These
are true enough, so we could say that Python isn't slow when it's delegating to
some C code or a database or similar. If you're looking at a JSON or CSV
microbenchmark, Python is about as fast as C, because the Python JSON/CSV
libraries are virtually pure, optimized C. If you're maintaining a Python CRUD
webapp and thinking that Python's performance problem is overblown, it's
because Postgres, MySQL, etc are doing the heavy lifting for you. Other
applications might simply have very lax performance requirements.

Python is ultimately slow because it's not well-optimized, and it's not
well-optimized because it's slow. At some point (or probably many points), it
was decided that Python would address its performance problems by supporting
C-extensions (the ability to write libraries in C which Python programs could
call into for performance-sensitive work, or otherwise to allow Python to
leverage the existing, massive corpus of C library code). If code was slow, the
solution was to rewrite it in C (there probably was and still is some posturing
about how you shouldn't try to write performance-sensitive programs in Python,
but the problem with that is it's notoriously difficult to predict whether your
product will have some performance-sensitive edge case). However, the
C-extension interface was not well-designed and in practice it was the whole
CPython interpreter (CPython is the reference implementation for Python). Since
Python was so slow, the community leaned heavily on C-extensions to the extent
that any proposed improvement to the CPython interpreter that might break
C-extensions was untenable, and because the C-extension interface was so
expansive, this eliminated a huge swath of potential improvements including
many optimizations. Worse still, any other candidate Python implementation
would likewise have to choose between performance and breaking compatibility
with a huge swath of the Python ecosystem.

### Multiprocessing

Other more specific reasons for Python's relative sloth include its commitment
to a GIL (Global Interpreter Lock) which effectively precludes shared-memory
parallelism ("multi threading") optimizations. There are shared-nothing
parallelism implementations that ultimately amount to forking parallel Python
interpreter processes and communication between the processes via whatever IPC
(inter-process communication) facilities the operating system provides. This
ultimately means that the data that needs to be communicated between processes
must first be marshalled into something like JSON or Python's Pickle format,
written to a Unix pipe or similar, and then unmarshalled by the receiving
process. While these de/serialization libraries are typically highly-optimized
C libraries, the sheer volume of data to be transmitted can render this
approach infeasible for a huge swath of applications. Further, it's often
difficult to determine the feasibility for a given application without
significant work to restructure the application to be amenable to
parallelization in the first place, at which point the testing can begin with
no certainty that this multiprocessing approach will succeed. The applications
that are likely to be amenable to multiprocessing are those for which the
volume of communication is low (volume ~= message frequency times message
size).

### Pandas

So what about Pandas? Pandas is a general-purpose dataframe library built on
top of Numpy, which is a general-purpose "array" library. Both of these
libraries store data in contiguous blocks of memory instead of Python's
haphazard method for allocating memory sporadically throughout the heap. This
way of representing data in memory reduces pressure on garbage collectors
(there's only one big allocation for the whole block instead of one allocation
per element) and it also reduces the improves the CPU cache hit:miss ratio
because CPU cache lines are populated based on contiguous memory. Further,
Pandas and Numpy include a number of optimized C procedures for operating on
these data structures such that the program can do more work without having to
call back into the slower Python code. This is a good solution for code that is
a linear or tabular series of scalar values; however, even still, these
libraries can still make performance worse if, for example, you have to call
into a Python function for each cell, which may happen if the provided C
libraries aren't sufficient for your use case. At that point, you may well have
worse performance than a naive Python list implementation.

### Rewrite the slow parts in C/C++/Rust/etc

So what about "rewrite it in C/etc"? This approach is a more generalized
form of the "just use Pandas" advice. With Pandas, you have a pre-built,
pre-packaged C library, but it's only good for dataframe-friendly applications.
The "just rewrite it in X" advice allows you to trade off on the "must be
dataframe-shaped" restriction in exchange for the onus of writing, maintaining,
packaging, etc your own C-extension. Ultimately, this is a complex endeavor no
matter which "X" you choose (C, Rust, etc) if only because you're
dealing with the internal C-extension interface. If you (or your team and any
future teammates, in a team context) is not an expert in Python, C, **and** "X"
(Rust, etc), you just shouldn't do this. If you are such an expert, you
know that the engineering overhead for maintaining the native code and the
infrastructure for building and packaging a C-extension is probably not worth
the effort even in the best case. Further, not all cases are ideal--there is
still a lot of overhead in calling between C and Python: effectively you have
to marshal between the inefficient Python data structures and the
efficient C data structures that allow C-extensions to be performant in the
first place. This is the same fundamental constraint as the multiprocessing
problem: the volume of communication becomes the problem. In order to get a
benefit from these C-extensions, the volume of communication must be
sufficiently low, either by making the number of calls across the C/Python
boundary few or by making the volume of data marshalled small (or both). In
many cases, the volume of data is necessarily large (a large, heterogeneous
Python object graph) and in order to make the number of calls across the
C/Python boundary few in number, you must rewrite a huge corpus of Python code
into C/Rust/etc. This means that only your veteran C/Python/etc
programmers can feasibly extend or maintain that code, and it will still be
more costly to do so than the equivalent Python.


### Python vs JS

So why is Python so much slower than JS? As previously discussed, JavaScript
similarly began its life as an interpreted dynamic language; however, it was
optimized to the extent that it's now quite a lot faster than Python for most
applications. Ultimately, JavaScript was originally and principally a browser
scripting language. In the browser environment, there was no permitted analog
to C-extensions (presumably for security reasons). As such, JS didn't
accumulate an ecosystem that limited implementation improvements to the extent
that CPython's C-extension ecosystem did. This allowed Google and others to
build more sophisticated, faster JavaScript runtimes which ubiquitously
leveraged a class of optimizations that are designed to make dynamic languages
fast: JIT (just-in-time) compilation. It wasn't until web browsers made JS
implementations that were fast enough to make JavaScript broadly appealing
outside of the browser, and the canonical implementation of JavaScript as a
general purpose application language is NodeJS, which wraps Chrome's V8
JavaScript engine with a standard library suitable for interfacing with modern
operating systems (and support for native extensions similar to CPython's
C-extensions); however, the native extension interface for Node is not as
sprawling as CPython's and it only applies to NodeJS--one can't run these
extensions in the browser.

## Other Python implementations

There are other Python implementations, some of which are faster than CPython.
Most of those achieve their performance by sacrificing compatibility with the
C-extension part of the ecosystem, relegating them to a sort of fringe status
with such low adoption that virtually no one bothers to make sure their
packages are compatible with those implementations.

### Pypy

The lone exception in my mind is Pypy (not to be confused with the Python
package index, Pypi), which boasts performance on the order of 2X-10X faster
than CPython for pure-Python code (i.e., it doesn't make C-extensions any
faster). It also has the noble goal of compatibility with the C-extension
ecosystem. Pypy is fast because it uses JIT compilation techniques like moder
JS implementations; however, its approach to compatibility has been to
gradually support more and more of the informal CPython C-extension interface.
That said, there are still many C-extension packages which aren't compatible
with Pypy, including popular packages like Psycopg2 (the canonical Python
bindings for Postgres). In some cases, there may be pure-Python (or otherwise
Pypy-friendly C-extensions) alternatives, but as with Psycopg2, these
alternatives are typically too poorly supported to be considered for many
production use cases. What Pypy has accomplished is remarkable, particularly
given the relatively poor investment it has received; still, that's not
consoling to those of us who would like to be able to use it in our production
applications (ideally our organizations would sponsor Pypy development so it
*can* be a serious contender one day).

## So how is $SUCCESSFUL_COMPANY so successful despite Python?

Some combination of "throwing money at the problem" (if this applies to you,
congratulations), lax performance requirements, and a problem space that is
amenable to offloading the performance-sensitive onto a database or Pandas or
similar. Great engineers understand when their specific context allows them to
effectively flout guidelines and broad advice.


## So what to do?

If you're in the enviable position of starting a new project, I strongly advise
against Python. I've seen many projects start on the assumption that they won't
be performance-sensitive applications, and any performance issues they run into
will be solved with Pandas, C, multiprocessing, etc; however, invariably those
projects run into major performance issues right around the time they begin to
experience real customer workloads, and their bottlenecks aren't amenable to
any of the aforementioned solutions (because of the communication volume
constraint discussed in earlier paragraphs). Further, besides the performance
issue, Python has other notorious issues pertaining to building, packaging, and
distributing/deploying programs (which similarly don't have good solutions at
present) which I didn't discuss in this article, but which have been discussed
at length elsewhere. There are now other languages which have many of the
purported benefits of Python--fast developer iteration cycles and a robust
ecosystem. My prefered Python alternative has been Go, and I strongly recommend
it as it addresses all of the problems of Python performance without
introducing severe problems of its own (overwhelmingly, the chief complaint
about Go is that its static type system isn't expressive enough, and that
sometimes it is necessary to drop down into its dynamic type system, which is
to say that sometimes you have to write untyped code, as you do in Python all
the time).
