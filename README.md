# circular

`circular` is a Core Lightning plugin that helps lightning nodes rebalance their channels.

It optimizes on **fees**, and it's designed to be used by routing nodes who do not need reliability in their payments and just want to rebalance their channels at the cheapest possible rate.
It features a custom pathfinding algorithm that remembers liquidity information about the graph thanks to failed payments. Initially it doesn't know anything about the graph, but it will learn about it as it fails payments.

## Building
You need Go 1.13 or higher to build this plugin.

```bash
git clone https://github.com/giovannizotta/circular.git
cd circular
go build -o circular cmd/circular/*.go
```

## Running
This plugin is dynamic, meaning that you can start and stop it via the CLI. For general plugin installation instructions see [How to install a plugin](https://github.com/lightningd/plugins/blob/master/README.md#Installation)

The executable that you have just built is called `circular`.

## Usage
There are two options for running a circular rebalance at the moment:

via scid:
```bash
lightning-cli circular -k inscid=123456x1x1 outscid=345678x1x1 amount=200000 maxppm=10 attempts=1
```

via nodeID:
```bash
lightning-cli circular-node -k outnode=123abc innode=345def amount=200000 maxppm=10 attempts=1
```

The required parameters are:
* `outnode` or `outscid`: the node/scid that you want to use to send the payment.
* `innode` or `inscid`: the node/scid where you want to receive the payment.

Optional parameters are:
`amount`(sats, default=200000) is the amount that you want to rebalance and `maxppm`(default=10) is the maximum ppm that you are willing to pay. `attempts`(default=1) is the number of payment attempts that will be made once a path is found.

## Roadmap
The following is a list of features that will be added in the future:
* Allow the user to omit the `outscid` or `outnode` parameter and let the plugin find the best alternative
* Liquidity aging policy: right now if there is a failure on a channel, the liquidity belief doesn't move until that channel is used again. This information might change over time, and we want to keep that into account.
* Concurrent rebalancing attempts
* More testing

## Contributing
If you have any problems feel free to open an issue. Pull requests are welcome as well.