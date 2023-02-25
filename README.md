# udptunneler

A simple UDP tunneling tool to forward UDP multicast traffic through a TCP connection (aka tunnel), written in Go.

## How it works
The purpose of the `udotunneler` is to transfer udp multicast data between two different network when that multicast channel is not available.

A typical use case is when you have a multicast channel available in your lan, and you want to access it from a remote location,
for example through a VPN connection where the multicast traffic is not allowed.

The `udptunnel` is composed of two parts: the client and the server.</b>
The client joins a multicast group and forwards the received datagrams to the server, which in turns multicasts them on its own subnet.

## Build

You can simply build the udptunneler project with the provided Makefile.

```
# build the binary
make build

# build the multi arch binaries (default linux and windows)
make clean build_all

# build the multi arch binaries with specific version
VERSION=0.0.1 make clean build_all
```

See the makefile for details.

You can download the current binaries from the [release page].

## Usage
In order to use the `udptunneler` you need to start the server side first, then the client side.
If client can not connect to the server i will exit.

### Server
The `server` command listens to a TCP listener address and publish the received datagrams to a multicast channel.

```shell
$ udptunneler server -h
Start UDP tunneler server

Usage:
  udptunneler server [flags]

Flags:
  -a, --address string    the udp destination address (ip:port) where the server is publishing the forwarded datagrams. If not provided, datagrams are published on the same channel joined by the client
  -d, --dump              dump the raw bytes of the message
  -h, --help              help for server
  -l, --listener string   the tcp server listener address and port used to listen for client connections (default ":5055")
```

Example:

```shell
$  udptunneler server -d -l :5055 -a 231.1.1.102:10202
```

### Client
The `client` command connects to the `udptunnler` server and send to it the received datagrams.

```shell
$ udptunneler client -h
Start UDP tunneler client

Usage:
  udptunneler client [flags]

Flags:
  -a, --address string     the udp destination IP and port of the channel we want to join
  -d, --dump               dump the raw bytes of the message
  -h, --help               help for client
  -i, --interface string   the network interface used to join the provided multicast channel provided
  -s, --server string      the tcp address (ip:port) of the server to which the datagram will be forwarded
```

Example:

```shell
$ udptunneler client -d -a 231.1.1.101:10101 -i eno1 -s my-server:5055
```

### Ping
The `ping` command publish an `hello, world` message on the multicast channel. It can be used for testing the multicast channel.

```shell
$ udptunneler ping -h
Send multicast 'Hello world' ping message. Can be used for testing the multicast channel.

Usage:
  udptunneler ping [flags]

Flags:
  -a, --address string   the udp destination IP and port of the channel we want to join
  -h, --help             help for ping
```

Example:

```shell
$ udptunneler ping -a 231.1.1.101:10101
```

### Dump
The `dump` command listen to the multicast channel and dump the received datagrams. Ic an be used for testing the multicast channel.
```shell
$ udptunneler dump -h
Dump UDP multicast traffic

Usage:
  udptunneler dump [flags]

Flags:
  -a, --address string     the udp destination IP and port of the channel we want to join
  -h, --help               help for dump
  -i, --interface string   the network interface used to join the provided multicast channel provided

```

Example:

```shell
$ ./bin/udptunneler dump -a 231.1.1.102:10202 -i eno1 
```
## UdpTunneler Protocol
The `udptunnler`  uses a simple framed TCP binary protocol, with little endian byte order.

```
+----------------+--------------------+------------------------------+
+ Frame Length   | Packet Header      | Packet Body                  |
+----------------+--------------------+------------------------------+
```

**Frame Length**: a unit16 representing the length of the frame (including the header length)

**Packet Header**: a byte containing the packet type

**Packet Body**: the packet body depends on the packet type and it's optional

There are 2 packet types:

**Heartbeat Packet**: type 0x01, no body

**Datagram Packet**: type 0x02,with following packet body:
 * Datagram Length (uint16): number of bytes of the datagram packet
 * UDP Channel Address (uint32, ipv4): destination address of the multicast group which the client joined to receive that datagram
 * UDP Channel Port (uint16): destination port of the multicast group which the client joined to receive that datagram
 * Datagram Packet (variable byte array): actual datagram received by the client from the multicast channel
 
