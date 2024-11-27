---
public: true
publish_date: 2024-10-10T00:00:00Z
title: "Whats up at Port 25"
---

> _An attempt to bring some clarity to email_.


## TL;DR


## Introduction

Email and SMTP (Simple Mail Transfer Protocol) has been around for a long long time. In fact it has been around, in one form or another, since the 70s.
It is the most widely used form of communication on the internet and a large fraction of all internet traffic is in fact email. 

The protocol itself was design in an era and for the explicit purpose for you to send a text message to the only other guy using the mainframe. This meant that things like confidentiality, security and authenticity was not really a big concern, or even reasoned about, when the protocol was designed. Nor was it really a problem for the first 10-20 years of its existence. But as we will see, it is a big short coming today and has been since the rise of spam in the 90s when people started to get online.

## The Basics

The SMTP protocol is in its basic form, accomplishing 80% of what you need, very simple.

You connect to a server on port `25`. A raw tcp socket will do, even if TLS nowadays is both supported and recommended.

Once connected you simply send you do some had wavy protocol stuff and then you just write the email to the server. The server will thank you and go about its business of delivering the email. No authentication, No verification, no encryption, no nothing.

There are two portions to keep track of, the protocol it self and the email format. 
The protocol, how to talk to the server, we inform the server of the who the email is from and who at the server it is intended for. This in fact has nothing to do what is later in the email it self. The email format is the actual email that will be delivered and consists of two parts the headers and the body (which might be split up into sections.).


First we need to figure out what server to talk to. Luckly there is a DNS record type for that, the MX record. 
If we want to deliver an email to `bob@gmail.com` we can look up the MX record for the domain gmail.com and get a list of servers that are responsible for receiving emails for users on gmail.com.

```bash 
dig mx gmail.com
gmail.com.		60	IN	MX	5 gmail-smtp-in.l.google.com.
gmail.com.		60	IN	MX	10 alt1.gmail-smtp-in.l.google.com.
gmail.com.		60	IN	MX	20 alt2.gmail-smtp-in.l.google.com.
gmail.com.		60	IN	MX	30 alt3.gmail-smtp-in.l.google.com.
gmail.com.		60	IN	MX	40 alt4.gmail-smtp-in.l.google.com.
```

pick the one with the highest priority (lowest number), in this case `gmail-smtp-in.l.google.com` with 5, and connect to it on port `25` using `telnet`.


```bash 
$ telnet gmail-smtp-in.l.google.com 25
S: 220 mx.google.com ESMTP 38308e7fff4ca-2fb247434a1si5559721fa.352 - gsmtp
C: EHLO example.com
S: 250-mx.google.com Hello example.com [192.168.1.2]
S: 250-SIZE 14680064
S: 250-PIPELINING
S: 250 HELP
C: MAIL FROM:<bounce-to-me@example.org>
S: 250 Ok
C: RCPT TO:<bob@gmail.com>
S: 250 Ok
C: DATA
S: 354 End data with <CR><LF>.<CR><LF>
C: From: "The Boss" <the-boss@example.com>
C: To: "Bob Bobson" <bob@gmail.com>
C: Date: Tue, 10 Oct 2024 16:06:16 +0200
C: Subject: Test message
C:
C: Hello Bob.
C: This is a test message
C: Your friend,
C: The Boss
C: <CR><LF>.<CR><LF>  ## really just the chars {\r\n.\r\n}
S: 250 Ok: queued as 12345
C: QUIT
S: 221 Bye
```

> ignore the `S:` and `C:` they are just to show who is talking, the server or the client.


There are some things to note in the SMTP conversation above. 

`MAIL FROM:` command specifies the server that is relying the email and who to send a bounce to. Its not the actual / logical sender of the email. Well they probably were the same back in the days, but not anymore. If you inspect a raw email you recived and look at the headers, the `MAIL FROM` addres will appear as the `Return-Path` header which is added by the receiving server.  

`RCPT TO:` is in which mailbox that the email should be delivered to, and again has nothing to do with what is displayed when looking at the email in regards to who the recipient is.

`From:` header within the actual email is the one that is displayed to the recipient, and usaly the one that is used to reply to the email. But again, this address i arbitrary and can be anything.

`To:` header is the one that is used to determine who the email is "for", and is the one that is displayed to the recipient. Just as with the `From` header the `To` header can be anything and is complete arbitrary and has nothing to do with who the email was delivered to. In practice however, the `To` or `Cc` header will contain the `RCPT TO` address, but this is not a requirement. Eg if you are Bcc on an email, your email won't appear in the headers, by design. 


## It cant be that easy?

Well yes and no. You can definitely send emails, like above, but using command line tools such as `sendmail` or `curl` (in which cases the `DATA` portion is written to a text file prior to engaging in some cli karate) is much easier. Or for that matter most programming languages have libraries that can do this for you, eg Go standard lib, [pkg.go.dev/net/smtp](https://pkg.go.dev/net/smtp).

Example using `curl`:

```bash 
curl smtp://gmail-smtp-in.l.google.com \
   --mail-from bounce-to-me@example.org \
   --mail-rcpt bob@gmail.com \ 
   --upload-file email.txt
```
and `email.txt` would contain the email body which is written under the `DATA` command.
```bash 
cat email.txt
From: "The Boss" <the-boss@example.com>
To: "Bob Bobson" <bob@gmail.com>
Date: Tue, 10 Oct 2024 16:06:16 +0200
Subject: Test message

Hello Bob.
This is a test message
Your friend,
The Boss
```

Most of the emails sent this way may even end up your inbox, if you "de-spamed" them a few times and are using gmail.

### The not so easy part.

Since the world has been drowning in spam since the 90s, for the very reason described above and ease of use, the world has responded with RFCs amending the SMTP protocol to retrofit functionalities that gives you like confidentiality, integrity and authentication along with massive black box efforts in trying to deal with spam classifications. 

> One of the reasons that gmail got to be so popular was that they were the first to implement a spam filter that actually worked using bayesian filtering. Cutting edge machine learning in the early 2000s.

Alot of effort has been put in to make the email system more secure and reliable. It is however still a mess in many regards. SMTP is a federated protocol
and there is no central authority. But in practice Google, Yahoo, Microsoft and a few other have captured the majority of the users and in their egernes to protect them from spam and phishing, made it very very hard to send emails to their users if your on the outside. 

The discurraging part of trying to send emails your self is that, even if you follow all the rules and standards, you will still end up in the spam folder or just have your emails dropped at times.
Then trying to figure it out why will lead you down a path of talking with support tech at Microsoft (and others, but mostly Microsoft) that will gas-light-the-shit-out-of-you and having you sign for different "delivery programs", registering your IP address, washing blacklists or out right try to have you pay third partis to unblock you. But it all boils down to that the Outlook-AI-Spam-Classifier (TM) told the system the mail is spam and people at Microsoft have no idea why.
 
The thing you should do to have you emails delivierd boils down to three things, SPF, DKIM and DMARC and there is a fun resource when it in [learndmarc.com](https://www.learndmarc.com/)

#### SPF - Sender Policy Framework

Boils down to that if you eg. want to send emails from `@gmail.com` you have to define what ip addresses are allowed to send emails containing `@gmail.com` in the `From` header. This is done by deining a DNS TXT record pointing out the ip addresses  

[wikipedia.org/wiki/Sender_Policy_Framework](https://en.wikipedia.org/wiki/Sender_Policy_Framework)


Example
```bash 
dig txt gmail.com
gmail.com.              60      IN      TXT     "v=spf1 redirect=_spf.google.com"
## So we then keep resolving
dig txt _spf.google.com
_spf.google.com.        60      IN      TXT     "v=spf1 include:_netblocks.google.com include:_netblocks2.google.com include:_netblocks3.google.com ~all"
## and keep going
dig txt _netblocks.google.com
_netblocks.google.com.  60      IN      TXT     "v=spf1 ip4:35.190.247.0/24 ip4:64.233.160.0/19 ip4:66.102.0.0/20 ip4:66.249.80.0/20 ip4:72.14.192.0/18 ip4:74.125.0.0/16 ip4:108.177.8.0/21 ip4:173.194.0.0/16 ip4:209.85.128.0/17 ip4:216.58.192.0/19 ip4:216.239.32.0/19 ~all"
```

When a mail server recives an email, it looks at the IP address of the connection and compares it with the SPF record of the domain. If the IP address is not in the SPF record something is not right it might be spam.

The receiving email adds a header to the received email with the result of the SPF check and can look as follows. 

```txt
Received-SPF: pass (google.com: domain of bob@gmail.com designates 209.85.220.41 as permitted sender) client-ip=209.85.220.41;
```

#### DKIM - DomainKeys Identified Mail

DKIM is a way to sign an outgoing email and have any receiving party be able to verify the signature  

[wikipedia.org/wiki/DomainKeys_Identified_Mail](https://en.wikipedia.org/wiki/DomainKeys_Identified_Mail)

This is accoplished by genereating a public/key pair and publishing the public key in a DNS TXT record. The private key is then used to sign the email resulting in a DKIM-Signature header in the email. That can look as follows

```txt
DKIM-Signature: v=1; a=rsa-sha256; c=relaxed/relaxed;
        d=gmail.com; s=20230601; t=1728592864; x=1729197664;
        h=to:subject:message-id:date:from:mime-version:from:to:cc:subject
         :date:message-id:reply-to;
        bh=r9RC7dfXtv2b+fhZcFKQ70ls0Q/QqCn/kv82AjqJC+U=;
        b=PajBEgFH5y/JnSAYKF221wJP1u72PRAnBhFB7oeziaSSk+DH4L2r+ox9ZDDj+5MjDt
         IAxosznJgM2YiC78pMozyPkQ12WWh7bLgltAorpE6xghCJYzXUrziwVAZXKkiWbUVq/c
         w+AP90HrF7hZ6S1FUEUgMli5SyN6a5bVuL8rnz/6+hYBMaA2ac0b6khOqutep0ix5Wfn
         kbBvy9O5c4UnZwwkvGEcVem+768FqMgafCuTqvYEQwP/xzOqsDGYiD4Jr574vCioZf6Y
         ks1gA2e7HZkaojjTKCI5qeLac+WuhAaM0I4daEDVxXbXm/1Q4pYOHgiKUiF6Cn3oBuH4
         mxrw==
```

The receiving email server can then verify the signature by looking up the public key in the DNS record and verify the signature.
The public key can be found by constructing a DNS query for the `d` and `s` fields in the DKIM-Signature header. 

```bash
## <s>._domainkey.<d>
dig txt 20230601._domainkey.gmail.com
20230601._domainkey.gmail.com. 60 IN    TXT     "v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAntvSKT1hkqhKe0xcaZ0x+QbouDsJuBfby/S82jxsoC/SodmfmVs2D1KAH3mi1AqdMdU12h2VfETeOJkgGYq5ljd996AJ7ud2SyOLQmlhaNHH7Lx+Mdab8/zDN1SdxPARDgcM7AsRECHwQ15R20FaKUABGu4NTbR2fDKnYwiq5jQyBkLWP+LgGOgfUF4T4HZb2" "PY2bQtEP6QeqOtcW4rrsH24L7XhD+HSZb1hsitrE0VPbhJzxDwI4JF815XMnSVjZgYUXP8CxI1Y0FONlqtQYgsorZ9apoW1KPQe8brSSlRsi9sXB/tu56LmG7tEDNmrZ5XUwQYUUADBOu7t1niwXwIDAQAB"

```

Using the public key and the rest of the information of the DKIM-Signature header we can now do so crypto (no, not Bitcoin) to verify that the sender of the email indeed does control the domain in the `d` section.

#### DMARC - Domain-based Message Authentication, Reporting and Conformance

DMARC is a way to combine SPF and DKIM and add some policy to it. What should happen if the SPF or DKIM checks fail.

[wikipedia.org/wiki/DMARC](https://en.wikipedia.org/wiki/DMARC)

The DMARC policy is published in a DNS TXT record and can look as follows

```bash
dig txt _dmarc.gmail.com
_dmarc.gmail.com.       60      IN      TXT     "v=DMARC1; p=none; sp=quarantine; rua=mailto:mailauth-reports@google.com"
```

* The `p` field is the policy that should be applied if the SPF or DKIM checks fail. 
* The `sp` is the subdomain policy, if the email is sent from a subdomain of the domain in the `d` field of the DKIM-Signature header. 
* The `rua` field is where reports of failures and successes should be sent to.

In this case gmail is saying, if the SPF or DKIM checks fail, do nothing. If the email is sent from a subdomain of gmail.com, quarantine the email. And send reports to mailauth-reports@google.com



## The Setup

We want to be able to send emails from my domain `raz.sh` to notfiy people of new blog posts. It is in fact the "blog enging" / software that will send the emails from the same server as the blog post is published.

Simply, we need to setup SPF, DKIM and DMARC for the domain `raz.sh`

### SPF

For SPF it is easy, all we need is to add the `v=spf1 a -all` DNS TXT record to the domain `raz.sh`. 

* the `a` part indicates that ip address that is contained in the A/AAAA records of raz.sh are allowed to send emails from the domain.
* the `-all` part indicates all other ip addresses shall fail (ie. email that was not sent from `A raz.sh`). The `-` indicates that the failing spf checked emails should be rejected.

### DKIM

A bit more work is needed for DKIM. We need to generate a public/private key pair and publish the public key in a DNS TXT record. 

```bash

mkdir -p ./data/dkim
cd ./data/dkim

## Creating the private key, 2048 bits RSA
openssl genrsa -out dkim_private.pem 2048

## Extracting the public key into a DNS TXT record
echo "v=DKIM1; p=$( \
    openssl ec -in dkim_private.pem -pubout -outform der | openssl base64 -A \
)" > 202410._domainkey.raz.sh.txt

## This should be added to the DNS as a TXT record at 202410._domainkey.raz.sh
cat 202410._domainkey.raz.sh.txt  
v=DKIM1; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwvaKRca/zMD2eK+5RFDY2ZVjCeGB8P/HcLktXHBunkNf87wfg3NA2c3ohHunUygYdCsOG5j/dmHEza75hBS+agPVn+YnG2Je6EomWbCw62jsxWyOZTKtrf/VnzTKdGsPN1PJGJLJzt1EhGZE6fDdXLQ9EnLUwloIQCM7wR0afvdB5PUhZvKIZMlM9HToIk3t73ivhITVnjF4FoeMXuY+4AsjYhdKnA9Lxto3mccqCb9DW44RgGQOhPLATFYKeIvtkuhfWdaKoNASkoM067oLst+7fKl6VU/9pXmwMDv0lQgakEPiWAbjPlO5n3d1YLSVjYu/WDH1VaG33RrRGJMdwQIDAQAB

```

### DMARC

For DMARC we need to publish a DNS TXT record with the policy.

it should look like the following
```bash
dig txt _dmarc.raz.sh
_dmarc.raz.sh.       60      IN      TXT     "v=DMARC1; p=quarantine; sp=quarantine; rua=mailto:shrubs-ebony-jeep@duck.com"
```