# BidderD

## RTBKIT Agent(s) using Go and the HTTPInterface

`bidderd` allows you to have several RTBKIT agents, that bid with a
fixed price, behind an HTTP Interface.

To install:

Before installing you should setup your Go environment.

* `go get github.com/alep/bidderd`
* `go install github.com/alep/bidderd`

Usage:

`bidderd --config agents.json`

Where agents.json contains configuration for each agent. For an
example configuration see the agents.json example.

* ``"name"`` this is used to give the agent a name in the ACS
* ``"config"`` this is the configuration that will be registered in the ACS
* ``"price"`` price of a single bid
* ``"balance"`` amount of money used to update the bank account
* ``"period"`` period for bank pacing.


As soon as the program is started it will register the agents in the
ACS. You can quit with `C-c` and this will unregister the agents from
the ACS.


This started as an example and copying ideas from the [RTBKIT's Python example][1]


[1]: https://github.com/rtbkit/rtbkit/blob/master/rtbkit/examples/py-bidder/http_bid_agent_ex.py  "http_bid_agent_ex.py"
