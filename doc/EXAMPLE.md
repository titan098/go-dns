# Example

`go-dns` can be used as a simple DNS server to serve dynamic DNS requests for your IPv6 prefix. Depending on what is available to you, you will need:

1. A routable IPv6 Prefix where the reverse lookup can be delegated to a server.
2. A place to run `go-dns`.
3. A top-level domain name where you can delegate DNS requests. A sub-domain is recommended.

## Example Setup

1. An IPv6 prefix from tunnelbroker.net. I make use of a `/48` prefix so the network can be split into multiple `/64` networks, however a `/64` prefix will work just as well.
2. `go-dns` running a Raspberry Pi 3 in a docker container.
3. A domain name registered with AWS Route 53.

## Step 1: Decide on a sub-domain

To make things easier, you can delegate all IPv6 dynamic domains to a sub-domain for example where all domains will fall under the following sub-domain:


```
ipv6.example.com
```

Should you wish to distinguish between different network types, it is possible to create sub-domains which `go-dns` will respond to. For instance, all dynamic addresses could be subdomains under:

```
dynamic.ipv6.example.com
```

It would also be possible to delegate other networks to different sub-domains, such as:

```
vpn.ipv6.example.com
```

## Step 3: Decide the name of the nameserver

This is a domain name which points directly to the host where you are hosting `go-dns`. It does not need to have `ipv6` connectivity, it must just be resolvable and have an `A` record. If you are running in an environment where NAT is being used then the appropriate port forwards will need to be created. Nevertheless, the service **must** be running on port 53 and not restricted by your firewall so that the world can see it.

For the sake of this example, say that the domain name you decide you will host `go-dns` will be `ipv6-ns.example.com` and that will have an `A` record which points to the IP of the host where `go-dns` is running.

In the case of Route 53:

1. Create a Record Set with the name `ipv6-ns` for the domain `example.com`
2. Create an `A` record for that Record Set that contains the IP address of the host
3. **(Optional)** Create an `AAAA` record that contains the IPv6 address of the host

## Step 2: Config

To create the top-level config the following will be required:

1. The routable prefix
2. The name of the sub-domain that the address should be mapped to

In the case where the domain name is `example.com`, where all dynamic IPv6 domain names should appear under `dynamic.ipv6.example.com`, and the routable prefix is `2001:db8:1:1::/64`, the config would appear as follows:

```toml
[dns]
port = 53
protocol = "udp"

[dns.soa]
ttl = 300
refresh = 3600
retry = 1800
expire = 10800
minimum = 300
mname = "ns-ipv6.example.com"
rname = "dns.example.com"

[dns.ns]
servers = ["ns-ipv6.example.com"]

[subdomain."dynamic.ipv6.example.com"]
response_type = "Dynamic"
prefix = "2001:db8:1:1::"
mask = 64
```

## Step 3: Run!

Start `go-dns` by running:

```
$ ./go-dns -c config.toml
```

## Step 4: Set the rDNS for the Subnet

This will be dependent on the provider. For tunnelbroker.net the following be set for the tunnel:

- `rDNS Delegated NS1`: `ipv6-ns.example.com`

This will tell the provider to send any reverse DNS requests for the subnet to `ipv6-ns.example.com`

## (Optional) Step 5: Configure forward lookups

`go-dns` will respond to forward lookups for any of the domains which it manages. This is done simply by setting an `NS` record for the subdomain `ipv6.example.com` to point to `ipv6-ns.example.com`. In the case of Route 53 this can be done as follows:

1. Create a Record Set with the name `ipv6` for the domain `example.com`
2. Create an `NS` record for that Record Set that contains `ipv6-ns.example.com.` *(note the trailing dot)*

This will tell any DNS lookups for `X.ipv6.example.com` to query the domain server running on `ipv6-ns.example.com`.
