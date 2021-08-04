<h1 align="left">
  <img src="static/notify-logo.png" alt="notify" width="170px"></a>
  <br>
</h1>


[![License](https://img.shields.io/badge/license-MIT-_red.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/projectdiscovery/notify)](https://goreportcard.com/report/github.com/projectdiscovery/notify)
[![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/projectdiscovery/notify/issues)
[![GitHub Release](https://img.shields.io/github/release/projectdiscovery/notify)](https://github.com/projectdiscovery/notify/releases)
[![Follow on Twitter](https://img.shields.io/twitter/follow/pdiscoveryio.svg?logo=twitter)](https://twitter.com/pdiscoveryio)
[![Chat on Discord](https://img.shields.io/discord/695645237418131507.svg?logo=discord)](https://discord.gg/KECAGdH)

Notify is an helper utility written in Go that allows you to pipe the output from various tools (or read from a file) and post the same to slack, discord, telegram etc (we call them providers)

# Resources
- [Resources](#resources)
- [Usage](#usage)
- [Installation Instructions](#installation-instructions)
    - [From Binary](#from-binary)
    - [Using go CLI](#download-using-go-cli)
    - [From Source](#build-from-source)
- [Running notify](#running-notify)
  - [Provider config](#provider-config)
  - [Example provider config](#example-provider-config)

# Usage

```sh
▶ notify -h
```

This will display help for the tool. Here are all the switches it supports.

| Flag | Description | Example |
|------|-------------|---------|
| -config string            | Notify configuration file| notify -config config.yaml
| -silent                  | Don't print the banner| notify -silent 
| -version                 | Show version of notify| notify -version 
| -v                       | Show Verbose output| notify -v 
| -no-color                | Don't Use colors in output| notify -no-color
| -data string             | File path to read data from| notify -data test.txt
| -bulk                    | Read the input and send it in bulk, character limit can be set using char-limit flag| notify -bulk 
| -char-limit int          | Character limit for message (default 4000)| notify -char-limit 2000 
| -provider-config string  | provider config path (default: $HOME/.config/notify/provider-config.yaml)| notify -provider-config testproviderconfig.yaml
| -provider string[]       | provider to send the notification to (optional)| notify -provider slack -provider telegram
| -profile string[]        | profile to send the notification to (optional)| notify -profile recon


# Installation Instructions

### From Binary

The installation is easy. You can download the pre-built binaries for your platform from the [releases](https://github.com/projectdiscovery/notify/releases/) page. Extract them using tar, move it to your `$PATH`and you're ready to go.

```sh
Download latest binary from https://github.com/projectdiscovery/notify/releases

▶ tar -xvf notify-linux-amd64.tar
▶ mv notify-linux-amd64 /usr/local/bin/notify
▶ notify -version
```

### Download using go cli

Notify requires **go1.14+** to install successfully. Run the following command to download and install notify -


```sh
▶ GO111MODULE=on go get -v github.com/projectdiscovery/notify/cmd/notify
```


### Build From Source

```sh
▶ git clone https://github.com/projectdiscovery/notify.git; cd notify/cmd/notify; go build; mv notify /usr/local/bin/; notify -version
```


# Running notify

Notify supports piping output of any tool and send it to configured provider/s (e.g, discord, slack channel) as notification.

Following command will enumerate subdomains using [SubFinder](https://github.com/projectdiscovery/subfinder) and probe for alive URLs and sends the notifications of alive URLs using [httpx](https://github.com/projectdiscovery/httpx) to configured provider/s.

```
subfinder -d hackerone.com | httpx | notify
```

<h1 align="left">
  <img src="static/notify-httpx.png" alt="notify-httpx" width="700px"></a>
  <br>
</h1>

Following command will enumerate subdomains using [SubFinder](https://github.com/projectdiscovery/subfinder) and probe alive URLs using [httpx](https://github.com/projectdiscovery/httpx), runs [Nuclei](https://github.com/projectdiscovery/nuclei) templates and send the nuclei results as a notifications to configured provider/s.


```
subfinder -d intigriti.com | httpx | nuclei -t files | notify
```

In similar manner, output (stdout) of any tool can be piped to **notify** for posting data into slack/discord.


# Provider config


The tool tries to use the default provider config (`$HOME/.config/notify/provider-config.yaml`), it can also be specified via CLI by running. 

To run the tool just use the following command.

```sh
▶ notify -provider-config path/to/testproviderconfig.yaml
```

## Example provider config
The default provider config file can be created at `$HOME/.config/notify/provider-config.yal` and can have the following contents:

```yaml
slack:
  - id: "recon"
    slack_channel: "test"
    slack_username: "test"
    slack_webhook_url: "https://hooks.slack.com/services/XXXXXX"
  - id: "test"
    slack_channel: "test"
    slack_username: "test"
    slack_webhook_url: "https://hooks.slack.com/services/XXXXXX"
discord:
  - id: "test"
    discord_channel: "test"
    discord_username: "Sajad"
    discord_webhook_url: "https://discord.com/api/webhooks/XXXXXXXX"
telegram:
  - id: "recon"
    telegram_api_key: "XXXXXXXXXXXX"
    telegram_chat_id: "XXXXXXXX"
``` 


## References:- 

- [Creating Slack webhook](https://slack.com/intl/en-it/help/articles/115005265063-Incoming-webhooks-for-Slack)
- [Creating Discord webhook](https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks)
- [Creating Telegram bot](https://core.telegram.org/bots#3-how-do-i-create-a-bot)

Notify is made with 🖤 by the [projectdiscovery](https://projectdiscovery.io) team.
