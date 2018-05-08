# go-dns

## Introduction

`go-dns` dynamically generates `AAAA` and `PTR` dns records for an IPv6 prefix. In cases where SLAAC, and other dynamic IPv6 methods are being used to assign addresses, it is not feasible to maintain reverse DNS records for all assignable addresses. `go-dns` will dynamically return `AAAA` and `PTR` records based on the queried IPv6 address.

Given the following IPv6 address, the reverse address will be generated:

```
2001:470:1f2f:23:14f5:d526:ec16:ba83 -> 14f5d526ec16ba83.ipv6.ellefsen.ninja
```

Where the original network prefix is: `2001:470:1f2f:23::/64`

## Configuration

The configuration file `config.toml` is as follows:

```toml
[dns]
port = 15353
protocol = "udp"

    [dns.soa]
    ttl = 300
    refresh = 3600
    retry = 1800
    expire = 10800
    minimum = 300
    mname = "ns.ipv6.domain.example.com"
    rname = "dns.ipv6.domain.example.com"

    [dns.ns]
    servers = ["ns.ipv6.domain.example.com"]

[domains]
    [domains."ipv6.domain.example.com"]
    prefix = "2001:470:1f2f:23::"
    mask = 64
```

### `[dns]` section

The `dns` section contains the configurable options for the DNS server. The configuration options are:

| Option   | Description                            | Options        |
|----------|----------------------------------------|----------------|
| port     | the port the DNS server will listen on | Any value port |
| protocol | the protocol to use                    | `tcp` or `udp` |

### `[dns.soa]`

The `dns.soa` sections contains the defaults for the SOA records for the domain:

| Option  | Description                                            | Example |
|---------|--------------------------------------------------------|---------|
| ttl     | The time to live value of the SOA record               | 300     |
| refresh | Refresh period for secondary lookups                   | 3600    |
| retry   | The retry period for the secondary lookups             | 1800    |
| expire  | The number of seconds before before the record expires | 10800   |
| minimum | Time to live                                           | 300     |
| mname   | The primary nameserver for this zone                   |         |
| rname   | Email address for the administrator for the zone       |         |

### `[dns.soa]`

The `dns.soa` sections contains the defaults for the SOA records for the domain:

| Option  | Description                            |
|---------|----------------------------------------|
| server  | The list of namesevers for this domain |

### `[domains]` sections

The `domains` section contains the defintions of the prefix and the domain names for the generated responses. Each sub-domain should be contains in a section of the following form `[domains."<sub.domain.name>"]`, where each secion has the following form:

| Option  | Description                            |
|---------|----------------------------------------|
| prefix  | The IPv6 prefix for the reverse lookup |
| mask    | The network mask                       |

## Running

Running `go-dns` is as simple as follows:

```
$ ./go-dns -c <config_file>
```

The server will start listening on the specified port and will start to respond to dns requests.

