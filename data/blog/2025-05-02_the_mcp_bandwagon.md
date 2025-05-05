---
public: true
publish_date: 2025-05-02T00:00:00Z
title: "The MCP Bandwagon"
---


> MCP is an open protocol that standardizes how applications provide context to LLMs. Think of MCP
> like a USB-C port for AI applications. Just as USB-C provides a standardized way to connect your
> devices to various peripherals and accessories, MCP provides a standardized way to connect AI
> models to different data sources and tools.\
> -- Anthropic

# TL;DR

During the past month MCP (Model Context Protocol),
that would enable LLMs to become Agents and interact with the world, really have been blowing up.
The idea is straightforward,
let us standardize on api for LLM/Agents to interact with the world,
and how to inform the LLM/Agent about it.

Things are moving really fast
and Google recently released their own, as they claim, orthogonal standard
called [Agent2Agent](https://google.github.io/A2A).
MCP Server are being build ant published daily,
and can be found at eg [mcp.so](https://mcp.so/) and [pulsemcp.com](https://www.pulsemcp.com/).

However, I'm astonished by the apparent lack of grownups in the room.
All the major players spend billions of dollars on training and tuning their models,
just to turn around and, from what I can tell, have an intern write the documentation,
providing bad sdks and on and on.

This seems to have continued when it comes to MCP and some very strange design decisions,
poor documentation and even wors specification the actual protocols.

# Background

About three weeks ago, I decided to jump on the MCP bandwagon to give it a try and see how it could
be used in our own environment. I'm very much of a person who wants to understand how things
actually
work under the hood before I start using abstractions.
Here we have a new protocol that works over different transports, how fun!

Anthropic is the company behind the MCP standardization effort and MCP, imho, seems to be one of the
major reasons Anthropics CEO is thinking that most code will be written by LLMs in a year or so. The
bet on Coding tooling in particular seems
to have been the guiding principle of the standardization effort
in how it feels like working with it and from what i
tell.

## Protocol

Simply,
it is a JSON-RPC protocol with predefined methods/endpoints
that is designed to be used in conjunction with an LLM. This is not really the topic of this post,
but there are things to be criticized when it comes to the protocol as well.

_Eg._ a mandatory `id` field in
the [json struct](https://modelcontextprotocol.io/specification/2025-03-26/basic#requests) being
sent, that can be a `number` or a `string` for no apparent reason, just to fuck with all statically
typed languages. Essentially needing to support, `ints`, `floats` and `strings` for this field.

## Transport

As all with cool applications, post 2005, they’re local first (_\*irony\*_), and this seems to be
very
much the case of MCP. Looking at the transport protocol you kind of get where there comming from, if
their intention is to build out LLM tools for coding on your laptop. You probably are looking at the
local IDEA (or let's be real, Cursor or Windsurf) and how to have the LLM interact with the local
file system, your databases, your editor, your language server and so on

There are essentially two main transport protocols (or three).
Its stdio and "Something over HTTP we didn't really put that much thought into,
but the web seems to be a thing probably should support."

### Stdio

Using std**io** is essentially to start a local MCP Server,
hooking up `std-out` and `std-in` pipes from the server to the client and start sending json.
This is not really the way things are usually done in unix/linux land.
Using them for bidirectional communication kind of breaks the classic pipe pattern.
When bidirectional communication is needed, we usually reach for a socket. 
 
Kind of sounds like someone were looking for a socket

### SSE / Streamable HTTP

Using the HTTP transport is another story.
There are two versions of the same mistake, SSE (Server Side Events) transport and Streamable HTTP
(a made-up term) that using REST semantics with SSE.

It can be summarized as "Since we like SEE for the LLM steaming thiningi we are not using
websockets,
instead we are effectively implementing websockets on top of SSE
and calling it _Streamable HTTP_ to gaslight people
into thinking it's an accepted/known way of doing things."

They discuss the problem with Websocket in this PR
[modelcontextprotocol/pull/206](https://github.com/modelcontextprotocol/modelcontextprotocol/pull/206)
and they are making some very stange contortions and straw-man arguments to **not** use websocket.
It seems that I have a least one kindred spirit further down the
thread, [modelcontextprotocol/pull/206#issuecomment-2766559523](https://github.com/modelcontextprotocol/modelcontextprotocol/pull/206#issuecomment-2766559523)
that picks up on the gaslighting.

# A decent into madness

I set out to implement an MCP server in golang.
There isn't an official golang sdk, and I wanted to understand the protocol.
A big mistake for my sanity.

## The warning signs...

Looking at https://modelcontextprotocol.io the documentation of the MCP is poorly written,
(all LLM vendors seem to have an internal competition in writing confusing documentations),
the specification glazes over or just ignores important aspects of the protocol and provides no
examples in the conversation flow.
In fact it seems that the entire web page is not meant to read the standard,
but the webpage instead pushes you towards tutorials in how to implement their SDKs.

Looking at all example servers being implmented in python or js,
where the intention to download and run locally using stdio.
Python and js are probably the worst pick of languages to write something in,
if you want it to work on anyone else's computers.
In fact,
the authors seem to realize this since all examples are wrapped in a docker container.

> Be honest when was the last time you ran `pip install` and didn’t end up in dependency hell.

Am I pretentious / judgmental
in thinking that people in AI only really know Python and the "well it works on my computer"
is still an acceptable way of doing things?

If you want to run MCP locally,
wouldn't you prefer a portable language, like rust, golang or even Java or C#?

## The problem

When I started to implement the protocol,
and right away, it felt I had to start to backwards engineer the protocol.
Important aspects of the SSE portion is missing from the documentation,
and no one seemed to have implemented the _Streamable HTTP_ yet.
Not even their own tooling     
`npx @modelcontextprotocol/inspector@latest`.
(To be fair, it might have been a skill issue on my part
pulling the wrong version since its was there now when I came back to it.)
It's also available at https://inspect.mcp.garden

Once you grasp the architecture, you realize quickly that implementing an MCP server might become a
huge effort. The problem is that the SSE/Streamable HTTP implementations it is trying to act like a
socket, emulating stdio. To accplish a full duplex, the client hooks up to a SSE session from the
server (reads), and then proceeds to use another endpoint to do writes. A request should be posted
to `/a-endpoint`, and the response should then be read form a pre-existing open SSE connection.

# What should be done.

My attempt to fight windmills

## HTTP Transport

- HTTP = request-response
- SSE = one-way server push
- WebSocket = two-way full-duplex