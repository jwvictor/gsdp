## The GSD Protocol: a new communication protocol for the greater good

*WIP: in no way suitable for public consumption*

### Introduction

GSDP is based on an idea I had to build a new communications protocol for structured messages, primarily inspired by Paul Graham's essay on frighteningly big ideas. Both developmnt and feature planning are in their infancy, and I haven't thought entirely through what kind of structured messaging would be useful in both the work and personal spheres. My hope is to develop a system that is scalable, secure, and makes communication within and across organizations and groups efficient and seamless. 

### Techincal description

GSDP is an open-source protocol for the communication of "structured messages" (discussed later). It can be run privately on any domain (much like email) and communication can take place both within a domain and across domains. Unlike email, this is accomplished by running a single daemon -- there is no complex web of old tech that causes everyone to just pay the $5 for gSuite. Unlike Slack, inter-domain communication is a key functional point. And unlike both, the protocol is based around the idea of having a sophisticated, modern client application, capable of making sense of enormous information inflows by way of detailed semantic markup in the structure of the message.

To be clear, GSDP is in no way an aggregator of other things. It is its own, separate protocol. It is also *not* meant to be an email or Slack clone. The messages are highly structured, unlike these platforms, whereupon the messages are largely unstructured.

Furthermore, GSDP implements a cryptographically-secured identity system that makes it easy to keep in touch with contacts if they move domains (e.g. companies), restrict identity lookups via privacy settings, and classify different contacts into different "buckets," each with a logical set of permissions (e.g. your boss can't send you a personal message, but your friend can't assign your work). It is, furthermore, the basis upon which an admin permits an address for a given user on their GSDP server -- they register their public identity (a byte stream that includes, among other things, a public key) with their GSDP server.

Technically speaking, we plan to emphasize a few important technical points.

1. The importance of semantic markup and "structured messages" suitable for organization and digestion in a client-side application.
2. Privacy as a central point of concern, with end-to-end encryption protecting all types of messages, and identity management that is cryptographically secured and heavily customizable.
3. Design for the modern era, and for the shortfalls of existing communication mechanisms like email, Slack, and Facebook.
4. Speed. Modern communications move at the speed of light, and the protocol needs to support that. For this reason, persistent connections are established to domains and shared among clients, and are closed only after they remain unused for a period of time. This greatly shortens the overhead for the delivery of a message. 

### The state of things

1. Crypto: pretty good spot
2. Identity management and name servers: surprisingly solid
3. Client application: zero
4. Message types: plain only - biggest current point to focus on
5. Feature roadmap: pretty non-existent so far, except for some basic ideas about the first message types to build
6. Streaming to clients: needs to be spec'd, put into protobuf, and implemented 

This puts us very far from even a 0.1-alpha type release, so *contributors welcome*!

### This implementation

This implementation is based on Google's protobuf/gRPC stack and provides two things:

1. The proto3 definition needed to generate bindings for a variety of modern languages; and
2. A definition (in code, no spec yet -- sorry) of how the application should work. This is a golang implementation.

That said, it is the most barebones of implementations. It is CLI-only, and can only send and receive plaintext messages at the moment.

### Contributor instructions

Setup isn't really setup just yet (no pun intended). The easiest way is to `go get` this repo, `cd` into it, and run `make`. Then `cd` into the `driver` directory and run `make` to build the driver. This will generate a separate repo for just the generated protocol code. Test scripts are there for you to use the driver.

You'll probably want to make a `~/.gsdp.toml` file with an appropriate identity directory and root identity file path, so you don't have to deal with CLI flags or environment variables. (See the sample in `driver`.) Use the `domain_newid.sh` script to generate your first (root) identity that the server will run under for your domain.

This assumes you have the gRPC and protobuf compilers installed (if you don't, they're deps you'll need for compilation). If you need to get this stuff, follow the golang instructions at `grpc.io`.

After you're setup, just shoot a pull request my way!

### Shortcomings list 

This is intended to be a list of the shortcomings of existing mesaging platforms, so that what we build here is better than what exists currently. Feel free to add to it.

1. Poor handling of threads/chains and ad-hoc "sub-groups" of people
1. Difficult or impossible to effectively leave individual threads of communication
2. Bad security
3. No real semantic markup, which makes it hard to develop interesting/useful clients
4. Hard to find archived or deleted stuff of huge importance
5. Can't easily move identities between email addresses (e.g. workplace domains)
6. Hard to distinguish things you're still working on and things that are completed
7. Hard to track the path a task or question takes through an organization (or several)
8. Poor support for computer-generated messages
9. Open doors to spam
10. Hackish or zero support for bots
12. No concept of structured responses (e.g. answering a yes/no question, accepting a task assignment)

## Copyright and license

Copyright (c) 2017 Jason Victor. All rights reserved, except as follows. Code is released under the Modified BSD License in the included LICENSE file. 
