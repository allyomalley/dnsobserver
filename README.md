# DNSObserver

A handy DNS service written in Go to aid in the detection of several types of blind vulnerabilities. It monitors a pentester's server for out-of-band DNS interactions and sends notifications with the received request's details via Slack. DNSObserver can help you find bugs such as blind OS command injection, blind SQLi, blind XXE, and many more!

![ScreenShot](https://raw.githubusercontent.com/allyomalley/dnsobserver/master/notification.png)

For a more detailed overview and setup instructions, see:

https://www.allysonomalley.com/2020/05/22/dnsobserver/


## Setup

What you'll need:

* Your own registered domain name
* A Virtual Private Server (VPS) to run the script on (I'm using Ubuntu - I have not tested this tool on other systems)
* *[Optional]* Your own Slack workspace and a webhook

### Domain and DNS Configuration

If you don't already have a VPS ready to use, create a new Linux VPS with your preferred provider. Note down its public IP address.

Register a new domain name with your preferred registrar - any registrar should be fine as long as they allow setting custom name servers and glue records.

Go into your new domain's DNS settings and find the 'glue record' section. Add two entries here, one for each new name server, and supply both with the public IP address of your VPS.

Next, change the default name servers to:

```
ns1.<YOUR-DOMAIN>
ns2.<YOUR-DOMAIN>
```

### Server Setup

SSH into your VPS, and perform these steps:

* Install Go if you don't have it already. Installation instructions can be found [here](https://golang.org/doc/install)
* Make sure that the default DNS ports are open - 53/UDP and 53/TCP. Run:
	
	```
	sudo ufw allow 53/udp
	sudo ufw allow 53/tcp
	```

* Get DNSObserver and its dependencies:
	
	```
	go get github.com/allyomalley/dnsobserver/...
	```


### DNSObserver Configuration

There are two required arguments, and two optional arguments:

<hr />

**domain** ***[REQUIRED]***  
Your new domain name.

**ip** ***[REQUIRED]***  
Your VPS' public IP address.

**webhook** *[Optional]*  
If you want to receive notifications, supply your Slack webhook URL. You'll be notified of any lookups of your domain name, or for any subdomains of your domain (I've excluded notifications for queries for any other apex domains and for your custom name servers to avoid excessive or random notifications). If you do not supply a webhook, interactions will be logged to standard output instead. Webhook setup instructions can be found [here](https://api.slack.com/messaging/webhooks).

**recordsFile** *[Optional]*  
By default, DNSObserver will only respond with an answer to queries for your domain name, or either of its name servers. For any other host, it will still notify you of the interaction (as long as it's your domain or a subdomain), but will send back an empty response. If you want DNSObserver to answer to A lookups for certain hosts with an address, you can either edit the config.yml file included in this project, or create your own based on this template:

```
a_records:
  - hostname: ""
    ip: ""
  - hostname: ""
    ip: ""
```
 
Currently, the tool only uses A records - in the future I may add in CNAME, AAAA, etc). Here is an example of a complete custom records file:

```
a_records:
  - hostname: "google.com"
    ip: "1.2.3.4"
  - hostname: "github.com"
    ip: "5.6.7.8"
```

These settings mean that I want to respond to queries for 'google.com' with '1.2.3.4', and queries for 'github.com' with '5.6.7.8'.

<hr />

## Usage

Now, we are ready to start listening! If you want to be able to do other work on your VPS while DNSObserver runs, start up a new tmux session first. 

For the standard setup, pass in the required arguments and your webhook:

```
dnsobserver --domain example.com --ip 11.22.33.44 --webhook https://hooks.slack.com/services/XXX/XXX/XXX
```

To achieve the above, but also include some custom A lookup responses, add the argument for your records file:
```
dnsobserver --domain example.com --ip 11.22.33.44 --webhook https://hooks.slack.com/services/XXX/XXX/XXX --recordsFile my_records.yml
```

Assuming you've set everything up correctly, DNSObserver should now be running. To confirm it's working, open up a terminal on your desktop and perform a lookup of your new domain ('example.com' in this demo):

```
dig example.com
```

You should now receive a Slack notification with the details of the request!
