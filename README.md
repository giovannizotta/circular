# circular

Circular is a Core Lightning plugin that helps routing nodes rebalance their channels.

It optimizes on fees, and it's designed to be used by routing nodes who do not need reliability in their payments and just want to rebalance their channels at the cheapest possible rate.
It features a custom pathfinding algorithm that remembers liquidity information about the graph thanks to failed payments.

