# circular [![Tests](https://github.com/giovannizotta/circular/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/giovannizotta/circular/actions/workflows/tests.yml) ![GitHub](https://img.shields.io/github/license/giovannizotta/circular)

`circular` is a Core Lightning plugin that helps lightning nodes rebalance their channels.

It optimizes on **fees**, and it's designed to be used by routing nodes who do not need reliability in their payments and just want to rebalance their channels at the cheapest possible rate.
It features a custom pathfinding algorithm that remembers liquidity information about the graph thanks to failed payments. Initially it doesn't know anything about the graph, but it will learn about it as it fails payments.

## Features
* Lightweight
* No invoices
* Liquidity information is stored in `graph.json`
* Usage data is stored in the database

## Endpoints
* `circular`: Rebalance a channel by scid
* `circular-node`: Rebalance a channel by node id
* `circular-parallel`: Rebalance a channel using many channels as source in parallel
* `circular-stats`: Get stats about the usage of the plugin

Detailed explanation of the endpoints follows in the Usage section.

## Building
You need Go 1.18 or higher to build this plugin.

```bash
git clone https://github.com/giovannizotta/circular.git
cd circular
go build -o circular cmd/circular/*.go
```

## Running
This plugin is dynamic, meaning that you can start and stop it via the CLI. For general plugin installation instructions see [How to install a plugin](https://github.com/lightningd/plugins/blob/master/README.md#Installation).

The executable that you have just built is called `circular`.

## Usage
There are two options for running a circular rebalance at the moment:

### Rebalance one channel at a time
via Short Channel ID:
```bash
lightning-cli circular -k inscid=123456x1x1 outscid=345678x1x1 amount=200000 maxppm=10 attempts=1
```

via Node ID:
```bash
lightning-cli circular-node -k outnode=123abc innode=345def amount=200000 maxppm=10 attempts=1
```

Required parameters:
* `outnode` or `outscid`: the node/scid that you want to use to send the payment
* `innode` or `inscid`: the node/scid where you want to receive the payment

Optional parameters:
* `amount`(sats, default=200000) is the amount that you want to rebalance
* `maxppm`(default=10) is the maximum ppm that you are willing to pay
* `attempts`(default=1) is the number of payment attempts that will be made once a path is found
* `maxhops`(default=8) is the maximum number of hops that a path is allowed to have

### Rebalance a channel from many sources in parallel
```bash
lightning-cli circular-parallel -k inscid=123456x1x1 amount=500000 splits=5 splitamount=20000 maxppm=10 maxoutppm=50 attempts=1 maxhops=8 depleteuptopercent=0.5 depleteuptoamount=2000000
```

Required parameters:
* `inscid`: the Short Channel Id where you want to receive the payment.

Optional parameters:
* `amount`(sats, default=400000) is the **total** amount that you want to rebalance
* `splits`(default=4) is the maximum number of rebalances that will happen in parallel
* `splitamount`(sats, default=100000) is the amount that each rebalance will carry
* `maxoutppm`(default=50) is the maximum ppm of the outgoing channels that `circular` is allowed to use to rebalance `inscid`. Useful to avoid rebalancing a channel from channels where you can profit
* `maxppm`(default=10), `attempts`(default=1) and `maxhops`(default=8) are the same as for the `circular` command
* `outlist` is a JSON array of node ids that you want to use as sources. If this is specified, `maxoutppm` is ignored. An example of how to use this parameter is the following:
```bash
cli circular-parallel -k inscid=123456x1x1 outlist='["03700917a25f79a3e427fe86e49b5041b583c73dd223cfa9a87cd6be5076b7b7a5", "025614be3600e9899bc044d331ab58a9fe1ccf30e75ae35943cdd11218a0a55dba"]' amount=800000 splitamount=80000 splits=4 maxppm=5000
```

`depleteuptopercent` and `depleteuptoamount` are a bit special: 
* `depleteuptopercent`(default=0.2) is a threshold percentage for the amount to leave in the outgoing channels. This must be between 0 and 1.
* `depleteuptoamount`(sats, default=1000000) is a value in sats for the amount to leave in the outgoing channels.
The actual amount that is going to be left in the outgoing channels is the minimum of `depleteuptopercent` and `depleteuptoamount`.

Example: you have a 10M channel and you set `depleteuptopercent` to 0.2 (20%) and `depleteuptoamount` to 1000000. The actual amount that will be left in that channel will be the minimum of 0.2 and 1000000. So in this case, at least 1000000 sats will be left in that channel.

### Get stats about the usage of the plugin
```bash
lightning-cli circular-stats
```
This command will return the following stats:
* `graph_stats`: stats about the graph that `circular` has learned
* `successes`: successful rebalances done by `circular`
* `failures`: failed rebalances done by `circular`
* `routes`: routes taken by `circular`

## Contributing
If you have any problems feel free to open an issue or join our [Telegram group](https://t.me/+u_R8kAfpSJBjMjI0). Pull requests are welcome as well.

Special thanks to devzorLN🐸 for helping me test the plugin.
