---
public: true
publish_date: 2024-11-27T00:00:00Z
title: "Weaponizing SSO for profit"
---

This article looks the practice of SaaS vendors placing Single Sign-On, SSO, capabilities behind premium pricing tiers, effectively turning a fundamental security feature into a premium offering

# TLDR

Due to a higher awareness and need for security, new regulation and security frameworks (such as ISO 27001 and NIST CF) it has become more or less a must have for many to be able to manage user centrally.

This is a good thing. We have the technology, the standards and the knowledge to do so. It has however become an toxic meme for SaaS vendors to have a pricing tier structure where you need to pay for the highest tier to get to use SSO with there service. 

Effectively this is **Weaponizing SSO** to force companies into a higher tier then needed in order to access best practice security and often demanded by regulation or customers 

> "Ohh, you want long passwords over 8 chars for your accounts, that will be an extra $2 / user / month" \
> \
> "Ohh, you want to enable 2 Factor to login, that will be an extra $3 / users / month" \
> \
> "Ohh, you want to manage access to the application your self, that will be an extra $5 / user / month"


Why is only two of the statement above considered ridiculous?


# Background

I'm the CTO of a fintech company, with about 70 people and many of our customers being regulated. We are a cost aware, frugal company and as most these days we rely heavily on SaaS services. As we have grown the pain of all these per seat/user modeles are becoming more apperent and expensive. There are many aspects to the classical value based pricing tiers of SaaS applications worth criticizing, but what drives me crazy is practice of forcing customers into the highest tier for security reasons.

It is very common these days that SaaS providers charge for SSO through OAuth2 or SAML. SSO has during the past couple of years become more or less mandatory at our company, and many others, for new service, this pricing structures of the services bring pain. Many times we end up paying 2-3x the price for the service, just to be able to add SSO to whatever core fether we were looking for.



# Regulation

During the last 5 years a lot of new regulation coming online in particular from the EU. This includes regulation such as NIS, NIS2 and DORA (and im sure there are more that cover IT security.) 

Most of them boils down to that regulated organizations should have control over there IT security and there **vendors** should as well. Not a bad thing in and of itself, but EU being EU combines the regulations with hefty fines for things that someone somewhere at somepoint may consider non-compliance.      

While NISx and DORA don't specifically says so, its more or less implied that the regulated entities vendors all must be ISO 27001 certified to keep delivering their services. There are of corse ways around it, but in general everyone must start working in accordance to ISO 27001.   

And again, working in accordance with ISO 27001 is not a bad thing, it can bring tremendous value. What it does mean is that to be compliant means having SSO for your organization and strict access control. 

> If you use a SaaS application you dam well better be able to manage your users centrally, or else...   

Simply put, regulation drives the adoption and necessity of SSO

# The Premium Tier and Pricing strategy

There are many ways of price a product but two of the more common once on the web are a value based strategy or a cost based strategy. 

_Examples of a value based pricing model_ is Slack, Intercom and more. An indication of such a price strategy is that you are charged by the seat per month and that there are no clear reason for the single to drive such cost.

_Examples of a cost based pricing model_ is Digital Ocean, Sendgrid, AWS and more. An indication of such a price strategy is that you are charged by what you use with a markup. You pay for a vps at Digital Ocean and you are charged per email sent at Sendgrid

From my quick research and gut feeling, most companies with a cost based model seems to include a sso solution for free. This makes sense, it's easy to implement, it drives no extra cost and its makes their site more secure (handling password is not something you should do if you can avoid it)

However, looking at value based offerings we often see that you need to buy the premium tier in order to get the access to SSO in the service. Tiered pricing is more common than not and when we are looking from the purge of a value based pricing strategy it might make sense. But it does not.

Not that long ago, https and (at that time named) SSL was not that common on the internet. People wanting to provide a secure session needed to pay up to get a certificate. That time is gone once TLS certificates became a commodity and TLS encryption can be enjoyed by all for free these days.  

Similarly, not that long ago SSO once were kind of a fringe need for many, but these days it is not. Not for individuals and not for businesses. The implementation has also become ubiquitous through third party services, such as Auth0, or through open source liberties for OAuth2 and SAML in every considerable language

It seems like this rising need for managing user accounts at companies that ought to turn SSO in to a commodity instead has been hijacked by SaaS companies to push tiers and function on customer that they do not need nor want. 


> A side note, there are some, in terms of strategies, that want to eat the cake and have it to. They want to get paid for both the value and the cost as separate items. \
> Some good examples are
> 
> **Mailchimp**, who charges for \
> - Emails sent (cost) and \
> - Contacts you can send emails to (value) and \
> - Users that can write and send emails (value)
> 
> **Betterstack**, who charges you for \
> - Users that can view logs (value) \
> - Storage of logs (cost) \
> - Uptime checks (cost)
> - Integrating with Slack (value)
> - SSO (value)



# Shitty Service Offerings

At one point, [plaintextoffenders.com](https://plaintextoffenders.com) shamed companies that had bad security practices regarding passwords to do better. I think that companies that weaponize SSO in there pricing structure need to be called out. It makes for a shitty experience, us all less secure and purchasing experience leaving feeling taken advantage of

 \
 \ 

- **Github**\
Enterprise accounts only, `5.25x` the price of a regular team account, or an extra `$17/user/month` for Enterprise\

- **Bitbucket / Atlassian**\
An extra `$5/user/month` the price for SSO with Atlassian Guard

- **1 Password**\
An extra `$6/user/month`

- **Bitwarden**\
An extra `$2/user/month` 

- **Betterstack**\
An extra `$5/user/month` to use Okta or Azure 

- **Intercom**\
Expert account only, starting at an extra `$100/user/month`

- **Superhuman**\
Enterprise accounts only

- **OpenAI**\
Enterprise accounts only

- **Anthropic**\
Enterprise accounts only

- **Slack**\
`$5.25/user/month` for Business+ to support SAML and SCIM\

- **Gitbook**\
You need to be pro with an extra `$4/user/month` for SAML

- **Sendgrid**\
`4.5x` the price for SSO

  

 \
 \
_\* As of 2024-11-27_

_\* With most of the offerings above you get more then SSO. But to get SSO you must pay for the premium tier_

DM [me](https://x.com/c_r_holm) on twitter if you want something added or changed