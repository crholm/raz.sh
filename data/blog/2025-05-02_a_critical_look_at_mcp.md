---
public: true
publish_date: 2025-05-02T00:00:00Z
title: "A Critical Look at MCP"
---


> "MCP is an open protocol that standardizes how applications provide context to LLMs. Think of MCP like a USB-C port for AI
> applications. Just as USB-C provides a standardized way to connect your devices to various peripherals and accessories, MCP
> provides a standardized way to connect AI models to different data sources and tools."
>
> ― Anthropic

## Edit

This post was picked up at quite a few places and discussed 
- https://news.ycombinator.com/item?id=43945993
- https://www.reddit.com/r/programming/comments/1kg6zws/a_critical_look_at_mcp/
- https://lobste.rs/s/eusyqs/critical_look_at_mcp

## TL;DR

I would like for this to turn out to be a skill issue on my part, and hope that I'm missing something.

During the past month,[MCP (Model Context Protocol)](https://modelcontextprotocol.io/specification/2025-03-26), which would enable
LLMs to become agents and interact with the world, has really been blowing up. The idea is straightforward: let's standardize an
API for LLM/Agents to interact with the world and how to inform the LLM/Agent about it.

Things are moving really fast, and IBM recently released their own "orthogonal standard" to MCP
called [Agent Communication Protocol (ACP)](https://agentcommunicationprotocol.dev), followed closely by Google
announcing [Agent2Agent (A2A)](https://google.github.io/A2A).

MCP Servers and Clients are being built and published daily, and can be found at sites like [mcp.so](https://mcp.so/)
and [pulsemcp.com](https://www.pulsemcp.com/).

However, I'm astonished by the apparent lack of mature engineering practices. All the major players spend billions of dollars on
training and tuning their models, only to turn around and, from what I can tell, have an interns write the documentation,
providing subpar SDKs and very little in terms of implementation guidance.

This trend seems to have continued with MCP, resulting in some very strange design decisions, poor documentation, and an even
worse specification of the actual protocols.

My conclusion is that the whole suggested setup for HTTP transport (SSE+HTTP and Streamable HTTP) should be thrown out and
replaced
with something that mimics stdio... Websockets.

## Background

About three weeks ago, I decided to jump on the MCP bandwagon to give it a try and see how it could be used in our own
environment. I'm very much a person who wants to understand how things actually work under the hood before I start using
abstractions. Here we have a new protocol that works over different transports — how exciting!

Anthropic is the company behind the MCP standardization effort, and MCP seems to be one of the major reasons Anthropic's CEO is
thinking that most code will be written by LLMs within a year or so. The bet on coding tooling in particular seems to have been
the guiding principle of the standardization effort based on how it feels to work with it.

## Protocol

Simply put, it is a JSON-RPC protocol with predefined methods/endpoints designed to be used in conjunction with an LLM. This is
not really the focus of this post, but there are things to be criticized about the protocol itself.

## Transport

As with many applications post-2005, they're supposedly "local first" (*ironically*), and this seems to be very much the case with
MCP. Looking at the transport protocol, you get a sense of where
they're coming from—if their intention is to build LLM tools for coding on your laptop. They're probably looking at local IDEs (or
more realistically, Cursor or Windsurf) and how to have the LLM interact with the local file system, databases, editors, language
servers, and so on.

There are essentially two main transport protocols (or three):

1. stdio
2. "Something over HTTP, the web seems to be a thing we probably should support."

### Stdio

Using stdio essentially means starting a local MCP Server, hooking up `stdout` and `stdin` pipes from the server to the client,
and starting to send JSON and using `stderr` for logging. It kind of breaks the Unix/Linux piping paradigm using these streams for
bidirectional communication. When bidirectional communication is needed, we usually reach for a socket, unix socket or even a net
socket.

However, it is straightforward and easy to reason about, works out of the box in all OSes, no need to deal with sockets, and so
on. So even if there is a critique to be made, I get it.

### HTTP+SSE / Streamable HTTP

The HTTP transport is another story. There are two versions of the same mistake: HTTP+SSE (Server-Sent Events) transport, which is
being replaced by "Streamable HTTP" (a made-up term) that uses REST semantics with SSE. But with a whole lot of extra confusion
and corner cases on top.

It can be summarized as: "Since we like SSE for LLM streaming, we're not using WebSockets. Instead, we're effectively implementing
WebSockets on top of SSE and calling it 'Streamable HTTP' to make people think it's an accepted/known way of doing things."

They discuss the problems with WebSockets (and the reason for Streamable HTTP) in this
PR: [modelcontextprotocol/pull/206](https://github.com/modelcontextprotocol/modelcontextprotocol/pull/206), making some very
strange contortions and straw-man arguments to **not** use WebSockets. At least one other person in the thread seems to agree with
me: [modelcontextprotocol/pull/206#issuecomment-2766559523](https://github.com/modelcontextprotocol/modelcontextprotocol/pull/206#issuecomment-2766559523).

## A Descent into Madness

I set out to implement an MCP server in Golang. There isn't an official Go SDK, and I wanted to understand the protocol. This
turned out to be a mistake for mental health...

### The Warning Signs...

Looking at https://modelcontextprotocol.io, the documentation is poorly written (all LLM vendors seem to have an internal
competition in writing confusing documentation). The specification glosses over or ignores important aspects of the protocol and
provides no examples of conversation flow. In fact, it seems the entire website is not meant for reading the standard; instead, it
pushes you toward tutorials on how to implement their SDKs.

All example servers are implemented in Python or JavaScript, with the intention that you download and run them locally using
stdio. Python and JavaScript are probably one of the worst choices of languages for something you want to work on anyone else's
computer. The authors seem to realize this since all examples are available as Docker containers.

> Be honest... when was the last time you ran `pip install` and didn't end up in dependency hell?

Am I being pretentious/judgmental in thinking that people in AI only really know Python, and the "well, it works on my computer"
approach is still considered acceptable? This should be glaringly obvious to anyone that ever tried to run anything from
Hugging Face.

If you want to run MCP locally, wouldn't you prefer a portable language like Rust, Go, or even VM-based options such as Java or
C#?

### The Problem

When I started implementing the protocol, I immediately felt I had to reverse-engineer it. Important aspects of the SSE portion
are missing from the documentation, and no one seemed to have implemented the "Streamable HTTP" yet; not even their own tooling
like `npx @modelcontextprotocol/inspector@latest`. (To be fair, it might have been a skill issue on my part, pulling the wrong
version, since it was available when I checked again a few weeks later. You can also find version at
[inspect.mcp.garden](https://inspect.mcp.garden), which might be more convenient.)

Once you grasp the architecture, you quickly realize that implementing an MCP server, or a client, could be a huge effort. The
problem is that the SSE/Streamable HTTP implementations are trying to act like sockets, emulating stdio, without being one and is
trying to do _Everything Everywhere All at Once_.

#### HTTP+SSE Mode

[modelcontextprotocol.io/specification/2024-11-05/basic/transports](https://modelcontextprotocol.io/specification/2024-11-05/basic/transports)

In **HTTP+SSE mode**, to achieve full duplex, the client sets up an SSE session to (e.g.) `GET /sse` for reads. The first read provides
a URL where writes can be posted. The client then proceeds to use the given endpoint for writes, e.g., a request to
`POST /a-endpoint?session-id=1234`. The server returns a 202 Accepted with no body, and the response to the request should be read
from the pre-existing open SSE connection on `/sse`.

#### "Streamable HTTP" Mode

[modelcontextprotocol.io/specification/2025-03-26/basic/transports](https://modelcontextprotocol.io/specification/2025-03-26/basic/transports)

In **"Streamable HTTP" mode**, they realized that instead of providing a new endpoint in the first request, they could use an
HTTP header for the session ID and REST semantics for the endpoint. For example, `GET` or `POST /mcp` can open an SSE session and
return an `mcp-session-id=1234` HTTP header. To send data, the client does requests to `POST /mcp` and adds the HTTP header
`mcp-session-id=1234`. The response may:

- Open a new SSE stream and post the reply
- Return a 200 with the reply in the body
- Return a 202, indicating the reply will be written to one of any pre-existing SSE stream

To end the session, the client may or may not send a `DELETE /mcp` with the header `mcp-session-id=1234`. The server must maintain
state with no clear way to know when the client has abandoned the session unless the client nicely ends it properly.

### What Are the Implications for SSE Mode?

This is such a problematic design that I don't know where to begin.

While some key features of SSE mode are undocumented, it's fairly straightforward once you reverse-engineer it. But this still
puts a huge and unnecessary burden on the server implementation, which needs to "join" connections across calls. Doing anything
real will pretty much force you to use a message queue to reply to any request. E.g., running the server in any redundant way will
mean that the SSE stream might come from one server to the client, while the requests are being sent to a completely different
server.

### What Are the Implications for "Streamable HTTP"?

The **Streamable HTTP** approach takes it to another level with a host of security concerns and obfuscated control flow. While
keeping all the bad parts from SSE mode, Streamable HTTP seems to be more of a super-set of confusion over SSE mode.

In terms of implementation I have just scratched the surface, but from what I understand in the docs...

**A new session can be created in 3 ways:**

- An empty `GET` request
- An empty `POST` request
- A `POST` request containing the RPC call

**An SSE can be opened in 4 different ways:**

- A `GET` to initialize
- A `GET` to join an earlier session
- A `POST` to initialize a session
- A `POST` that contains a request and answers with an SSE

**A request may be answered in any of 3 different ways:**

- As an HTTP response to a `POST` with an RPC call
- As an event in an SSE that was opened as a response to the `POST` RPC call
- As an event to any SSE that was opened at some earlier point

#### General implications

With its multiple ways to initiate sessions, open SSE connections, and respond to requests, this introduces significant
complexity.
This complexity has several general implications:

- **Increased Complexity**: The multiple ways of doing the same thing (session creation, SSE opening, response delivery) increases
  the cognitive load for developers. It becomes harder to understand, debug, and maintain code.
- **Potential for Inconsistency**: With various ways to achieve the same outcome, there's a higher risk of inconsistent
  implementations across different servers and clients. This can lead to interoperability issues and unexpected behavior. Clients
  and servers just implementing the parts they feel are necessary.
- **Scalability Concerns**: While Streamable HTTP aims to improve efficiency, with a charitable interpretation, the complexity
  will introduce scalability bottlenecks that need to be overcome. Servers might struggle to manage the diverse connection states,
  response mechanisms over a large number of machines.

#### Security Implications

The "flexibility" of Streamable HTTP introduces several security concerns, and here are just a few of them:

- **State Management Vulnerabilities**: Managing session state across different connection types (HTTP and SSE) is complex. This
  could lead to vulnerabilities such as session hijacking, replay attacks or DoS attacks by creating state on the server that
  needs to be managed and kept around waiting for a session to be resumed.
- **Increased Attack Surface**: The multiple entry points for session creation and SSE connections expand the attack surface. Each
  entry point represents a potential vulnerability that an attacker could exploit.
- **Confusion and Obfuscation**: The variety of ways to initiate sessions and deliver responses can be used to obfuscate malicious
  activity.

### Authorization

The latest version of the protocol contains some very opinionated requirements on how authorization should be done.

[modelcontextprotocol.io/specification/2025-03-26/basic/authorization](https://modelcontextprotocol.io/specification/2025-03-26/basic/authorization)

> - Implementations using an HTTP-based transport SHOULD conform to this specification.
> - Implementations using an STDIO transport SHOULD NOT follow this specification, and instead retrieve credentials from the
    environment.

I'm reading it like, for stdio, do whatever. For HTTP, you better fucking jump through these OAuth2 hoops. Why do I need to
implement OAuth2 if I'm using HTTP as transport, while an API key is enough for stdio?

## What Should Be Done

I don't know, just kind of feel sad about it all...
It seems like the industry is peeing their pants at the moment ― _it feels great now, but it's going to be hard to deal with
later._

There is one JSON RPC protocol, and Stdio is clearly preferred as the transport protocol. Then we should try to have the HTTP
transport be as much like Stdio as we can make it, and only really deviate if we really, really need to.

- In Stdio, we have Environment Variables, in HTTP we have HTTP Headers
- In Stdio, we have _socket-like_ behavior with input and output streams, in HTTP we have WebSockets

That's it really. We should be able to accomplish the same thing on WebSockets as we do on Stdio. WebSockets are the appropriate
choice for transport over HTTP. We can do away with complex cross-server state management for sessions. We can do away with a
multitude of corner-cases and on and on.

Sure some things, like authorization, might be a bit more complicated in some instances (and easier in some); some firewalls out
there might block WebSockets; there might be extra overhead for small sessions; it might be harder to resume a broken session. But
as
_they_ say:

> Clients and servers MAY implement additional custom transport mechanisms to suit their specific needs. The protocol is
> transport-agnostic and can be implemented over any communication channel that supports bidirectional message exchange
>
> [modelcontextprotocol.io/specification/2025-03-26/basic/transports#custom-transports](https://modelcontextprotocol.io/specification/2025-03-26/basic/transports#custom-transports)

As an industry, we should optimize for the most common use-cases, not the corner-cases.

## Side note: Alternatives and Additions

As discussed above, there seem to be more protocols emerging. MCP is effectively "a protocol to expose an API to an LLM" (which
can create an agent). The more recent protocols from IBM and Google (ACP and A2A) are effectively "protocols to expose an Agent to
an LLM" (which can create an agent of agents).

Looking through the A2A specification, it seems like there's a very limited need for them. Even though they claim to be
orthogonal, most things in A2A could be accomplished with MCP as is or with small additions.

It boils down to two entire protocols that could just as well be tools in an MCP server. Even IBM seems to acknowledge that their
protocol isn't really necessary:

> "Agents can be viewed as MCP resources and further invoked as MCP tools. Such a view of ACP agents allows MCP clients to
> discover and run ACP agents..."
>
> ― IBM / [agentcommunicationprotocol.dev/ecosystem/mcp-adapter](https://agentcommunicationprotocol.dev/ecosystem/mcp-adapter)

My initial feeling is that the ACP protocol mostly seems like an attempt for IBM to promote their "
agent-building-tool" [BeeAI](https://beeai.dev/)

What both of the A** protocols bring to the table is a sane transport layer and a way to discover agents.

![14 competing standards](https://imgs.xkcd.com/comics/standards.png)