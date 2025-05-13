---
public: true
publish_date: 2025-05-01T00:00:00Z
title: "The Commoditization of SaaS"
description: |-
  This article examines the evolution of SaaS from a niche market to the dominant business model for tech products, particularly in the B2B sector. Initially offering scalability and flexibility through per-seat models, the author argues that this has led to customer exploitation, inflated costs, and the pricing of commodities as luxury items. The piece analyzes how essential features like SSO, audit logs, and user management are often locked behind expensive "Enterprise" tiers. By comparing the software landscape of the past with the present, the author illustrates how the proliferation of per-user subscriptions for services like communication and file storage has significantly increased costs despite the commoditization of many of these services, using Slack as a prime example of a product priced as a luxury despite readily available and often included alternatives. The article concludes that while the commoditization of SaaS is a positive development, pricing strategies have not adjusted to reflect this reality, leaving many businesses overpaying for services that have become standard.
---

## Background

During the last 15 years, SaaS has gone from an obscure niche to the go-to businesses model for tech
products and especially in the b2b market segments. Back in the days you bought a software that you
installed on your server for a one-time fee and maybe some support agreement with updates ever few
months or so. Today this is unheard of, and the absolute majority of the services that are built and
sold today are done so in a SaaS framing and a per-seat-model. The per-seat-model has been very
successful over the years because it allows companies to scale up as their business grows without
having to worry about buying expensive systems or large one-off payments upfront. However, if we
look at what has happened and think through the business cases, we can see how customers are being
taken advantage of, how costs have skyrocketed and how commodities are being priced as luxury items.

The basic setup with self-serve SaaS services is to have bout three pricing tiers,
e.g. $2/user/month, $19/user/month or $59/user/month. What then happens is that some bare-bones
barely functional version is the first tier, the second is the functional one everyone really needs
to buy, and the third tier is really the same as the second tier but all the things needed by an
organization larger than 10 people are locked up here the "Enterprise" tier. It usually consists of
things like, SSO, Audit Logs, User management, Role management, integrations, SOC reports, etc. (a
blog entry on the SSO part, [raz.sh/blog/2024-11-27_weaponizing_sso_for_profit](https://raz.sh/blog/2024-11-27_weaponizing_sso_for_profit))

> Alright, so what is so bad about this? companies can afford it right?

Well, if you’re a company with knowledge workers, your largest budget item is probably your
employees and your second is all software that you use. (if you’re a SaaS company, your
infrastructure is probably the second one, and software is pushed down to third. But still).
Software has eaten the world and has become very costly. We surely use more than we did 15 years
ago, but we were promised economies of scale, right?

### Then

15-20 years ago you bought

- Windows XP license for $60,
- Word, Excel and Outlook for another one time $100,
  and you were done. Sure, there was an AD and some Samba stuff going on, maybe a SharePoint or two,
  and some central bought software like a ERP and so on.

### Now

Now we buy

- Google Workspace for $15/user/month and Office365 for $20/user/month (because you need Excel)
- You add on Slack for another $15/user/month, since it's the DM service of choice
- Tack on Zoom for $20/user/month because everyone got all exited during covid.
- ... Dropbox
- ... Github
- ... CRM systems like Salesforce or Pipedrive
- ... 1Password for password manager
- ... Monday for project management
- ... and on and on ...\
  and the monthly cost per user is much greater than the lifetime cost for a users software 15 years
  ago.

## Analysis

So what is it that has happened? why am I bashing on Dropbox, Slack and Zoom in particular? There is
something strange about these three in particular from the list above. Can you spot it?

These three products kind of revolutionized their area of business when the first came on the scene.
While they weren’t the first file storage, dm and video conferencing service, they kind of created
the market for it with a niche that convinced the market to start using the service. But what
quickly happened was that competitors copied the functionality offered and integrated into their own
offering. So, for example, Dropbox, Slack and Zoom equivalents are included in all Google Workspace
offerings as Drive, Chat and Meet.

What does this mean? Take Slack as an example, it is priced as a luxury good. However, the service
itself is a commodity, imho, providing little extra value over what is included in Microsoft Teams,
Google Chat, even Campfire or even an IRC instance. They seem to get a way with it because of the
SaaS pricing model and the fetcher bloat that is caused by Enterprise deals and the companies need
to charge for something. (but for some reason, I can still not quote reply to things.)

## The Commoditization

The commoditization of SaaS is a fact in many cases, and it is a good thing. It means that the
market has matured to the point at which services are hard to differentiate, on paper anyway. This
is a good thing, it means that the market is working and that the services are good. However, the
pricing is not reflecting this. The pricing is still reflecting the early days of the services,
where they were new and shiny and the SaaS pricing model has helped it to stay a luxury good for no
seeming reason.

In general, the motivation with the SaaS pricing model usually is that it is value-based, and they
charge for the value the service brings the user. However, the comparative value for many of these
services has really gone down as the service category has become a commodity, but many customers
haven’t really noticed.

I'll continue bashing on Slack, since it's such a good example of a service that has become a
commodity.
However, somehow the price has not really changed,
and we’re still being charged for a luxury good
costing us about `$20/user/month` for a somewhat complete fetcher set.

Comparing this to competitors such as Google Chat
which is included in Google Workspace running you about `$15/user/month` together with Gmail,
Gemini, Google Drive, and on and on.

It is about the same story looking at Microsoft 365 and Teams,
its included, its good,
and the entire offering is less expensive than a license for just Slack in many cases

If you are looking at something stand alone, like Slack, 37
Signals [Campfire](https://once.com/campfire)
is probably a good choice and is offered for a onetime fee of `$299`

## Conclusion

The commoditization of SaaS is a fact in many cases, and it is a good thing.
But many companies are still changing use for a uniq luxury good, and somehow most of us are just
ignoring it.  