# go-dns

## Introduction

`go-dns` dynamically generates `AAAA` and `PTR` dns records for an IPv6 prefix. In cases where SLAAC, and other dynamic IPv6 methods are being used to assign addresses, it is not feasible to maintain reverse DNS records for all assignable addresses. `go-dns` will dynamically return `AAAA` and `PTR` records based on the queried IPv6 address.

Given the following IPv6 prefix and address the following reverse address will be generated:

```
Prefix: 
2001:db8:1:1::/64

Address:
2001:db8:1:1:14f5:d526:ec16:ba83 -> 14f5d526ec16ba83.dynamic.ipv6.example.com
```

where the original network prefix is: `2001:db8:1::/48`

## Configuration

The configuration file `config.toml` is as follows:

```toml
[dns]
port = 15353
protocol = "udp"

[dns.domain]
response_type = "NxError"
domain = "ipv6.example.com"
prefix = "2001:db8:1::"
mask = 48

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

[static."a.dynamic.ipv6.example.com"]
prefix = "2001:db8:1:1::100"
```

### `[dns]` section

The `dns` section contains the configurable options for the DNS server. The configuration options are:

| Option        | Description                            | Options        |
|---------------|----------------------------------------|----------------|
| port          | the port the DNS server will listen on | Any value port |
| protocol      | the protocol to use                    | `tcp` or `udp` |

### `[dns.domain]` section

The `dns.domain` section contain the top-level response for this domain prefix. If your network is broken into a set of smaller networks, then this can be an `NxError` type which will act as a catch all for top-level dns queries and return an `NXERROR`

| Option        | Description                            |
|---------------|----------------------------------------|
| prefix        | The IPv6 prefix for the reverse lookup |
| comain        | The top-level domain being managed     |
| response_type | The response type for this subdomain   |
| mask          | The network mask                       |

### `[dns.soa]` section

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

### `[dns.ns]`

The `dns.ns` sections contains the defaults for the ns records for the domain:

| Option  | Description                            |
|---------|----------------------------------------|
| server  | The list of namesevers for this domain |

### `[subdomain]` sections

The `subdomain` section contains the defintions of the prefix and the domain names for the generated responses. Each sub-domain should be contains in a section of the following form `[subdomain."<sub.domain.name>"]`, where each secion has the following form:

| Option        | Description                            |
|---------------|----------------------------------------|
| prefix        | The IPv6 prefix for the reverse lookup |
| response_type | The response type for this subdomain   |
| mask          | The network mask                       |

## Response types

There are currently three types of responses, these are `NxError`, `Dynamic`, or `Static`. These response types determine how the responses are constructed.

### `NxError`

The `NxError` response type will return `NXERROR` for every query. It can be uses for fall through responses or where you need to return an `NXERROR` for every query.

### `Dynamic`

The `Dynamic` response type will return an `AAAA` and corresponding `PTR` for a DNS query. The domain name that is returned is of the form:

```
xxxxxxxxxxxxxxxx.domain.example.com
```

where the first part of the domain is the hex representation of the IP address after the IPv6 prefix. The domain name is reversable so the IPv6 address can be reconstructed from the DNS name.

### `Static`

Returns an `AAAA` and `PTR` record for a specificed DNS name. This will override any dynamic prefix responses that may include a the specified IP address. The mask for a static response will always be defaulted to `128`, regardless of any prefix value that has been set.

## Example

The example config above will create a valid instance of `go-dns`, however it will not be useful outside a few examples:

1. Query the reverse domain for `2001:db8:1:1::1`:

```bash
$ dig @127.0.0.1 -p 15353 -x 2001:db8:1:1::1

; <<>> DiG 9.11.5-P1-1ubuntu2.5-Ubuntu <<>> @127.0.0.1 -p 15353 -x 2001:db8:1:1::1
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 6047
;; flags: qr rd; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0
;; WARNING: recursion requested but not available

;; QUESTION SECTION:
;1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.1.0.0.0.1.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa. IN PTR

;; ANSWER SECTION:
1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.1.0.0.0.1.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa. 3600 IN PTR 0000000000000001.dynamic.ipv6.example.com.

;; Query time: 0 msec
;; SERVER: 127.0.0.1#15353(127.0.0.1)
;; WHEN: Sun Aug 18 13:37:43 SAST 2019
;; MSG SIZE  rcvd: 217
```

2. Query the reverse domain for `2001:db8:1:1::100` (a static assignment in config)

```bash
$ dig @127.0.0.1 -p 15353 -x 2001:db8:1:1::100

; <<>> DiG 9.11.5-P1-1ubuntu2.5-Ubuntu <<>> @127.0.0.1 -p 15353 -x 2001:db8:1:1::100
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 12620
;; flags: qr rd; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0
;; WARNING: recursion requested but not available

;; QUESTION SECTION:
;0.0.1.0.0.0.0.0.0.0.0.0.0.0.0.0.1.0.0.0.1.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa. IN PTR

;; ANSWER SECTION:
0.0.1.0.0.0.0.0.0.0.0.0.0.0.0.0.1.0.0.0.1.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa. 3600 IN PTR a.dynamic.ipv6.example.com.

;; Query time: 0 msec
;; SERVER: 127.0.0.1#15353(127.0.0.1)
;; WHEN: Sun Aug 18 13:38:12 SAST 2019
;; MSG SIZE  rcvd: 202
```

3. Perform a lookup dyanmic domain:

```bash
$ dig @127.0.0.1 -p 15353 AAAA 0000000000000001.dynamic.ipv6.example.com

; <<>> DiG 9.11.5-P1-1ubuntu2.5-Ubuntu <<>> @127.0.0.1 -p 15353 AAAA 0000000000000001.dynamic.ipv6.example.com
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 36486
;; flags: qr rd; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0
;; WARNING: recursion requested but not available

;; QUESTION SECTION:
;0000000000000001.dynamic.ipv6.example.com. IN AAAA

;; ANSWER SECTION:
0000000000000001.dynamic.ipv6.example.com. 3600	IN AAAA	2001:db8:1:1::1

;; Query time: 0 msec
;; SERVER: 127.0.0.1#15353(127.0.0.1)
;; WHEN: Sun Aug 18 13:44:25 SAST 2019
;; MSG SIZE  rcvd: 128
```

4. Perform a lookup for a static domain

```bash
$ dig @127.0.0.1 -p 15353 AAAA a.dynamic.ipv6.example.com

; <<>> DiG 9.11.5-P1-1ubuntu2.5-Ubuntu <<>> @127.0.0.1 -p 15353 AAAA a.dynamic.ipv6.example.com
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 50346
;; flags: qr rd; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0
;; WARNING: recursion requested but not available

;; QUESTION SECTION:
;a.dynamic.ipv6.example.com.	IN	AAAA

;; ANSWER SECTION:
a.dynamic.ipv6.example.com. 3600 IN	AAAA	2001:db8:1:1::100

;; Query time: 0 msec
;; SERVER: 127.0.0.1#15353(127.0.0.1)
;; WHEN: Sun Aug 18 13:38:46 SAST 2019
;; MSG SIZE  rcvd: 98

```

## Running

Running `go-dns` is as simple as running:

```
$ ./go-dns -c <config_file>
```

The server will start listening on the specified port and will start to respond to dns requests.

