# wgdash

![wgdash](https://user-images.githubusercontent.com/25783780/89130367-7470ed80-d4d2-11ea-8e28-78a22fcfba5f.png)

I like wireguard. But it's a little painful to use via the CLI,
especially when managing peers, and for my small use-case.

Wireguard will be used as a VPN gateway here, in a hub-and-spoke model, not via
a client-to-client (or point-to-point) method. This means that all peers are only
discoverable as long as the server/hub is available. As such, this gives us
the ability to do custom domain name resolution, as long as a DNS server is
installed on our wireguard server.

**NOTE**: this is still a work-in-progress. It is functional and I use it
personally but use at your own care.

## Requirements

- `go 1.14` or higher
- Linux distribution with [WireGuard](https://www.wireguard.com/install/) and
  `systemd`

## Installation & Usage

1. Clone repository
    - if you want to use a different virtual IP/port, copy over the
        `example_server_config.json` to `server_config.json` and make your
        changes there.
2. Run `go build`
3. Start via: `sudo ./wgdash`, sudo is required for a few reasons:
      - to save generated `wg0.conf` into `/etc/wireguard/`
      - to create/update `server_config.json`, this stores all public/private keys
        for clients and should only be root permissions.
      - to activate wireguard via `systemctl start wg-quick@wg0`
4. The dashboard can now be opened at `localhost:3100` in the browser. You can
   add/remove peers here. If a Virtual IP is not assigned, it will
   auto-increment based on the server Virtual IP and CIDR.
5. Download or use the QR button to grab the generated configuration files.
   Place them on your peers.
6. Test the connection from the peer first!

The dashboard does not need to be running constantly for wireguard to be
active. Once you are done making changes, you can `^C` out and your changes are
saved.

## TODO

- [ ] Edit peer configuration
- [ ] Add ability to insert custom DNS
- [ ] Add button to stop service
- [ ] Add better peer information
- [ ] Remove dependency on systemd
- [ ] Remove dependency on jQuery for frontend
- [ ] Create Dockerfile
