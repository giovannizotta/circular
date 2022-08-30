# circular [![Tests](https://github.com/giovannizotta/circular/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/giovannizotta/circular/actions/workflows/tests.yml) ![GitHub](https://img.shields.io/github/license/giovannizotta/circular)

`circular` is a Core Lightning plugin that helps lightning nodes rebalance their channels.

It optimizes on **fees**, and it's designed to be used by routing nodes who do not need reliability in their payments and just want to rebalance their channels at the cheapest possible rate.
It features a custom pathfinding algorithm that remembers liquidity information about the graph thanks to failed payments. Initially it doesn't know anything about the graph, but it will learn about it as it fails payments.

`circular` makes it easy to rebalance large amounts between your channels. Pathfinding is deterministic and straightforward. When there is a payment failure, the failing channel will be marked as unusable for that kind of amount until it is refreshed. Instead, if there is a success the information doesn't change, so if you issue the same command another time it will find the same successful route thanks to determinism in pathfinding. **Shorter routes are prioritized.**

## Features
* Lightweight
* No invoices
* Liquidity information is stored in `graph.json`
* Usage data is stored in the database

## Endpoints
* `circular-pull`: Pull liquidity into a channel using many channels as sources in parallel
* `circular-push`: Push liquidity out of a channel using many channels as destinations in parallel
* `circular`: Rebalance a channel by scid
* `circular-node`: Rebalance a channel by node id
* `circular-stats`: Get stats about the usage of the plugin
* `circular-delete-stats`: Delete stats about the usage of the plugin

Detailed explanation of the endpoints follows in the Usage section.

## Building
You can download the `circular` binary from the releases section. Alternatively, you can build the plugin on your own.
You need Go 1.18 or higher to build this plugin.

```bash
git clone https://github.com/giovannizotta/circular.git
cd circular
go build -o circular cmd/circular/*.go
chmod +x circular
```

## Running
This plugin is dynamic, meaning that you can start and stop it via the CLI. For general plugin installation instructions see [How to install a plugin](https://github.com/lightningd/plugins/blob/master/README.md#Installation).

The executable that you have just built is called `circular`.
The startup options are:
* `circular-graph-refresh` (**minutes**): How often the graph is refreshed. Default is 10.
* `circular-peer-refresh` (**seconds**): How often the list of peers is refreshed . Default is 30.
* `circular-liquidity-refresh` (**minutes**): Period of time after which we consider a liquidity belief not valid anymore. Default is 300.
* `circular-save-stats` (**boolean**): Whether to save stats about the usage of the plugin. Default is true. Save this to false if you are not interested in stats, as this data can grow big if you are running a lot of rebalances. You can delete the stats with the method `circular-delete-stats`.

You can also set a preferred logging level.
For example, with this startup command you would refresh the graph every 5 minutes, peers every 60 seconds, and reset liquidity on channels every 120 minutes. You would also *not* save stats and set the logging level to **DEBUG**.

⚠ This command is meant to be an example of how to use startup options. You should probably use a configuration file for CLN instead of starting it in this way.️
```bash
lightningd --plugin=/path/to/circularexecutable --circular-graph-refresh=5 --circular-peer-refresh=60 --circular-liquidity-refresh=120 --circular-save-stats=false --log-level=debug:plugin-circular
```

## Usage
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

### Pull liquidity into a channel from many sources in parallel
```bash
lightning-cli circular-pull -k inscid=123456x1x1 amount=500000 splits=5 splitamount=20000 maxppm=10 maxoutppm=50 attempts=1 maxhops=8 depleteuptopercent=0.5 depleteuptoamount=2000000
```

Required parameters:
* `inscid`: the Short Channel Id where you want to pull liquidity.

Optional parameters:
* `amount`(sats, default=400000) is the **total** amount that you want to rebalance
* `splits`(default=4) is the maximum number of rebalances that will happen in parallel
* `splitamount`(sats, default=100000) is the amount that each rebalance will carry
* `maxoutppm`(default=50) is the maximum ppm of the outgoing channels that `circular` is allowed to use to rebalance `inscid`. Useful to avoid rebalancing a channel from channels where you can profit
* `maxppm`(default=10), `attempts`(default=1) and `maxhops`(default=8) are the same as for the `circular` command
* `outlist` is a JSON array of node ids that you want to use as sources. If this is specified, `maxoutppm` is ignored. An example of how to use this parameter is the following:
```bash
cli circular-pull -k inscid=123456x1x1 outlist='["03700917a25f79a3e427fe86e49b5041b583c73dd223cfa9a87cd6be5076b7b7a5", "025614be3600e9899bc044d331ab58a9fe1ccf30e75ae35943cdd11218a0a55dba"]' amount=800000 splitamount=80000 splits=4 maxppm=5000
```

`depleteuptopercent` and `depleteuptoamount` are a bit special: 
* `depleteuptopercent`(default=0.2) is a threshold percentage for the amount to leave in the outgoing channels. This must be between 0 and 1.
* `depleteuptoamount`(sats, default=1000000) is a value in sats for the amount to leave in the outgoing channels.
The actual amount that is going to be left in the outgoing channels is the minimum of `depleteuptopercent` and `depleteuptoamount`.

Example: you have a 10M channel and you set `depleteuptopercent` to 0.2 (20%) and `depleteuptoamount` to 1000000. The actual amount that will be left in that channel will be the minimum of 0.2 and 1000000. So in this case, at least 1000000 sats will be left in that channel.

### Push liquidity out of a channel to many destinations in parallel
**Symmetrical to `circular-pull`, but for pushing liquidity out of a channel.**
```bash
lightning-cli circular-push -k outscid=123456x1x1 amount=500000 splits=5 splitamount=20000 maxppm=10 minoutppm=50 attempts=1 maxhops=8 filluptopercent=0.5 filluptoamount=2000000
```

Required parameters:
* `outscid`: the Short Channel Id from which you want to push out liquidity.

Optional parameters:
* `amount`, `splits`, `splitamount`, `maxppm`, `attempts` and `maxhops` are the same as for the `circular-pull` command.
* `minoutppm`(default=50) is the minimum ppm charged by your node that a channel has to charge to be selected by `circular-push`. Useful to avoid rebalancing a channel to channels where you can't profit from.
* `inlist` is a JSON array of node ids that you want to use as destinations. If this is specified, `minoutppm` is ignored. An example of how to use this parameter is the following:
```bash
cli circular-push -k outscid=123456x1x1 inlist='["03700917a25f79a3e427fe86e49b5041b583c73dd223cfa9a87cd6be5076b7b7a5", "025614be3600e9899bc044d331ab58a9fe1ccf30e75ae35943cdd11218a0a55dba"]' amount=800000 splitamount=80000 splits=4 maxppm=5000
```

`filluptopercent` and `filluptoamount` are a bit special:
* `filluptopercent`(default=0.2) is a threshold percentage for the minimum amount that is allowed to stay as remote liquidity in the incoming channels. This must be between 0 and 1.
* `filluptoamount`(sats, default=1000000) is a value in sats for the minimum amount that is allowed to stay as remote liquidity in the incoming channels.
  The actual amount that is going to be left in the incoming channels is the minimum of `filluptopercent` and `filluptoamount`.

Example: you have a 10M channel and you set `filluptopercent` to 0.2 (20%) and `filluptoamount` to 1000000. The minimum amount of remote liquidity that will be left in that channel will be the minimum of 0.2 and 1000000. So in this case, at least 1000000 sats will be left in that channel.


### Get stats about the usage of the plugin
```bash
lightning-cli circular-stats > stats.json
```
This command will return the following stats:
* `graph_stats`: stats about the graph that `circular` has learned
* `successes`: successful rebalances done by `circular`
* `failures`: failed rebalances done by `circular`
* `routes`: routes taken by `circular`

It's a good idea to pipe the output into a file, since it can be quite big.
## Benchmarks
Here is the performance of the pathfinding algorithm on the mainnet lightning network graph as of August 2022 (about 16000 nodes and 80000 channels). The benchmarks consist in finding a route between two random nodes and measuring the time it takes to find the route. Different values of `maxhops` are tested to show that shorter routes take less time to compute. Those routes are preferred by `circular`, since the longer the route, the most likely it is to fail.

On a laptop:
```bash
goos: linux
goarch: amd64
cpu: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz

30 runs sampled
name                                 time/op
Graph_GetRoute/dijkstra_3_maxhops-8  2.44ms ± 2%
Graph_GetRoute/dijkstra_4_maxhops-8  3.25ms ± 5%
Graph_GetRoute/dijkstra_5_maxhops-8  6.57ms ± 5%
Graph_GetRoute/dijkstra_6_maxhops-8  10.6ms ± 5%
Graph_GetRoute/dijkstra_7_maxhops-8  14.6ms ±10%
Graph_GetRoute/dijkstra_8_maxhops-8  17.6ms ± 7%
```

On a Raspberry Pi 4:
```bash
goos: linux
goarch: arm64

30 runs sampled
name                                 time/op
Graph_GetRoute/dijkstra_3_maxhops-4  20.4ms ± 1%
Graph_GetRoute/dijkstra_4_maxhops-4  23.3ms ± 2%
Graph_GetRoute/dijkstra_5_maxhops-4  36.0ms ± 3%
Graph_GetRoute/dijkstra_6_maxhops-4  52.3ms ± 6%
Graph_GetRoute/dijkstra_7_maxhops-4  67.3ms ± 7%
Graph_GetRoute/dijkstra_8_maxhops-4  78.0ms ± 8%
```

To replicate these benchmarks on your own machine, you can use the following command:
```bash
cd graph
go test -bench=. -timeout 0 -count 30 | tee bench.txt

benchstat bench.txt
```
(you need `benchstat` installed)

## Contributing
If you have any problems feel free to open an issue or join our [Telegram group](https://t.me/+u_R8kAfpSJBjMjI0). Pull requests are welcome as well.

Special thanks to devzorLN🐸 for helping me test the plugin.
