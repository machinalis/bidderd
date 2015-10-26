# bidderd

## RTBKIT Agent(s) using Go and the HTTPInterface


### Features

`bidderd` allows you to have several RTBKIT agents, that bid with a
fixed price, behind an HTTP Interface. `bidderd` is barebones, that
means that it does not handle (yet):

* automatic reload of configuration
* unique bid ids
* tracking of wins and events
* multiple seats.

Although hopefully you can use it as an starting point for your
agents.

### Dependencies

* `"gopkg.in/bsm/openrtb.v1"`


### Install

Before installing you should setup your Go environment.

* `go get github.com/alep/bidderd`
* `go install github.com/alep/bidderd`

### Usage

`bidderd --config agents.json`

Where ``agents.json`` contains configuration for each agent. For an
example configuration see the ``agents.json`` example.

* ``"name"`` this is used to give the agent a name in the ACS
* ``"config"`` this is the configuration that will be registered in the ACS
* ``"price"`` price of a single bid
* ``"balance"`` amount of money used to update the bank account
* ``"period"`` period for bank pacing.


As soon as the program is started it will register the agents in the
ACS. You can quit with `C-c` and this will unregister the agents from
the ACS.

If there are multiple agents that can bid on the request, each agent
will add the bid to the **same** seat in the bid response.


This started as an example and copying ideas from the [RTBKIT's Python example][1]

[1]: https://github.com/rtbkit/rtbkit/blob/master/rtbkit/examples/py-bidder/http_bid_agent_ex.py  "http_bid_agent_ex.py"

### Future work

The ideas is to have managable agents and be able to dynamically start campaigns.
