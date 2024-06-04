#!/usr/bin/python

import logging
import time

from pyln.testing.fixtures import *  # noqa: F403
from pyln.testing.utils import only_one, sync_blockheight, wait_for
from util import get_plugin  # noqa: F401

LOGGER = logging.getLogger(__name__)


def test_circular(node_factory, bitcoind, get_plugin):  # noqa: F811
    l1, l2, l3 = node_factory.get_nodes(3, opts={"allow-deprecated-apis": True})
    l1.fundwallet(10_000_000)
    l2.fundwallet(10_000_000)
    l3.fundwallet(10_000_000)
    l1.rpc.connect(l2.info["id"], "localhost", l2.port)
    l2.rpc.connect(l3.info["id"], "localhost", l3.port)
    l3.rpc.connect(l1.info["id"], "localhost", l1.port)
    l1.rpc.fundchannel(l2.info["id"], 1_000_000, mindepth=1, announce=True)
    l2.rpc.fundchannel(l3.info["id"], 1_000_000, mindepth=1, announce=True)
    l3.rpc.fundchannel(l1.info["id"], 1_000_000, mindepth=1, announce=True)

    bitcoind.generate_block(6)
    sync_blockheight(bitcoind, [l1, l2, l3])

    cl1 = l1.rpc.listpeerchannels(l2.info["id"])["channels"][0][
        "short_channel_id"
    ]
    cl2 = l2.rpc.listpeerchannels(l3.info["id"])["channels"][0][
        "short_channel_id"
    ]
    cl3 = l3.rpc.listpeerchannels(l1.info["id"])["channels"][0][
        "short_channel_id"
    ]

    for n in [l1, l2, l3]:
        for scid in [cl1, cl2, cl3]:
            n.wait_channel_active(scid)

    # expected graph:
    # 1M       0/1M       0/1M      0
    # l1 ------ l2 ------ l3 ------ l1
    #      cl1       cl2       cl3

    # wait for plugin gossip refresh
    time.sleep(5)

    l1.rpc.call(
        "plugin",
        {
            "subcommand": "start",
            "plugin": str(get_plugin),
            "circular-graph-refresh": 1,
            "circular-peer-refresh": 1,
        },
    )

    time.sleep(5)

    l1.rpc.call(
        "circular",
        {
            "inscid": cl3,
            "outscid": cl1,
            "amount": 100_000,
            "splitamount": 25000,
            "maxppm": 1000,
        },
    )
    # expected graph:
    # .9M    .1M/.9M   .1M/.9M     .1M
    # l1 ------ l2 ------ l3 ------ l1
    #      cl1       cl2       cl3
    wait_for(
        lambda: only_one(l1.rpc.listpeerchannels(l3.info["id"])["channels"])[
            "to_us_msat"
        ]
        == 100_000_000
    )

    l1.rpc.call(
        "circular-pull",
        {
            "inscid": cl3,
            "maxoutppm": 1000,
            "amount": 100_000,
            "splitamount": 25000,
            "maxppm": 1000,
        },
    )
    # expected graph:
    # .8M    .2M/.8M   .2M/.8M     .2M
    # l1 ------ l2 ------ l3 ------ l1
    #      cl1       cl2       cl3
    wait_for(
        lambda: only_one(l1.rpc.listpeerchannels(l3.info["id"])["channels"])[
            "to_us_msat"
        ]
        == 200_000_000
    )
    stats = l1.rpc.call("circular-stats")
    LOGGER.info(f"circular-stats: {stats}")
    assert stats["graph_stats"]["nodes"] == 3
    assert stats["graph_stats"]["channels"] == 6
    assert stats["graph_stats"]["active_channels"] == 6
    assert stats["graph_stats"]["liquid_channels"] == 6
    assert stats["graph_stats"]["max_htlc_channels"] == 6
    assert len(stats["successes"]) > 0

    l1.rpc.call(
        "circular-push",
        {
            "outscid": cl1,
            "minoutppm": 0,
            "amount": 100_000,
            "splitamount": 25000,
            "maxppm": 1000,
        },
    )
    # expected graph:
    # .7M    .3M/.7M   .3M/.7M     .3M
    # l1 ------ l2 ------ l3 ------ l1
    #      cl1       cl2       cl3
    wait_for(
        lambda: only_one(l1.rpc.listpeerchannels(l3.info["id"])["channels"])[
            "to_us_msat"
        ]
        >= 290_000_000
    )
