# wgdash (WIP)

This will eventually be a wireguard management dashboard.

I like wireguard a lot. But it's a little painful to use via the CLI,
especially when adding new peers, and for my small-ish use-case.

Wireguard will be used as a VPN gateway here, in a hub-and-spoke model, not via
a Point-to-Point method. As such, I assume we will be running our own DNS.
Whether to bundle the DNS resolver here is still up in the air.
