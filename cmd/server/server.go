package server

import (
	"bufio"
	"fmt"
	"github.com/bytedance/gopkg/lang/mcache"
	constants "github.com/mgeri/udptunneler/pkg"
	"github.com/mgeri/udptunneler/pkg/frame"
	"github.com/mgeri/udptunneler/pkg/packet"
	"github.com/mgeri/udptunneler/pkg/util"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

var (
	listenerAddress string
	udpAddress      string
	dumpBytes       bool

	udpConn        *net.UDPConn
	udpConnections = make(map[string]*net.UDPConn)

	Cmd = &cobra.Command{
		Use:   "server",
		Short: "Start UDP tunneler server",
		Long:  ``,
		RunE:  server,
	}
)

func init() {
	Cmd.PersistentFlags().StringVarP(&listenerAddress, "listener", "l", ":5055",
		"the tcp server listener address and port used to listen for client connections")
	Cmd.PersistentFlags().StringVarP(&udpAddress, "address", "a", "",
		"the udp destination address (ip:port) where the server is publishing the forwarded datagrams. If not provided, datagrams are published on the same channel joined by the client")
	Cmd.PersistentFlags().BoolVarP(&dumpBytes, "dump", "d", false,
		"dump the raw bytes of the message")
}

func server(cmd *cobra.Command, args []string) error {

	l, err := net.Listen("tcp", listenerAddress)
	if err != nil {
		log.Println(err)
	}

	defer l.Close()

	if udpAddress != "" {
		addr, err := net.ResolveUDPAddr("udp4", udpAddress)
		if err != nil {
			return err
		}
		udpConn, err = net.DialUDP("udp4", nil, addr)
		defer udpConn.Close()
	}

	// udp connections cleanup
	defer func() {
		for _, value := range udpConnections {
			if value != nil {
				value.Close()
			}
		}
	}()

	log.Printf("listening: %s", listenerAddress)

	for {
		c, err := l.Accept()
		if err != nil {
			log.Println("accept error:", err)
			break
		}
		// start a new goroutine to handle the new connection.
		go handleConn(c)
	}
	return nil
}

func handleConn(c net.Conn) {
	defer c.Close()
	frameCodec := frame.NewFrameCodec()
	rbuf := bufio.NewReader(c)
	wbuf := bufio.NewWriter(c)

	log.Printf("handleConn[%s <-> %s] new connection", c.RemoteAddr(), c.LocalAddr())

	for {
		// read from the connection

		// decode the frame to get the payload the payload is not decoded packet
		c.SetReadDeadline(time.Now().Add(constants.DefaultHeartbeatTimeout * time.Second))
		// buffer will be created by frame codec decode before read
		framePayload, err := frameCodec.Decode(rbuf)
		if err != nil {
			if err == io.EOF {
				log.Printf("handleConn[%s] disconnected", c.RemoteAddr().String())
			} else {
				log.Printf("handleConn[%s] frame decode error: %s", c.RemoteAddr(), err)
			}
			return
		}
		p, err := handlePacket(c, framePayload)
		mcache.Free(framePayload)
		if err != nil {
			log.Printf("handleConn[%s] packet handle error: %s", c.RemoteAddr(), err)
		}

		// write response
		if p != nil {
			buf := mcache.Malloc(p.Length())
			err = p.Encode(buf)
			if err != nil {
				log.Printf("handleConn[%s] packet encode error: %s", c.RemoteAddr(), err)
			}
			err = frameCodec.Encode(wbuf, buf)
			mcache.Free(buf)
			if err != nil {
				log.Printf("handleConn[%s] frame encode error: %s", c.RemoteAddr(), err)
			}
		}
	}
}

func handlePacket(clientCon net.Conn, framePayload []byte) (res packet.Packet, err error) {
	var p packet.Packet
	p, err = packet.Decode(framePayload)
	if err != nil {
		return nil, err
	}

	switch p.(type) {
	case *packet.Heartbeat:
		return p, nil
	case *packet.Datagram:
		datagram := p.(*packet.Datagram)
		var c = udpConn
		if c == nil {
			addr := net.UDPAddr{
				IP:   datagram.UdpIP,
				Port: int(datagram.UdpPort),
			}
			conn, ok := udpConnections[addr.String()]
			if !ok {
				c, err = net.DialUDP("udp4", nil, &addr)
				if err != nil {
					return nil, err
				}
				udpConnections[addr.String()] = c
			} else {
				c = conn
			}
		}

		// make sure all data will be written to outbound stream
		var f = datagram.DatagramPacket
		for {
			n, err := c.Write(f) // write the frame payload to outbound stream
			if err != nil {
				return nil, err
			}
			if n >= len(f) {
				break
			}
			if n < len(f) {
				f = f[n:]
			}
		}

		if dumpBytes {
			log.Printf(strings.Repeat("-", 80))
			log.Printf("src: %v, addr: %v, numBytes: %d\n",
				clientCon.RemoteAddr().String(), c.RemoteAddr(), len(datagram.DatagramPacket))
			util.DumpByteSlice(datagram.DatagramPacket)
		}

		return nil, nil
	default:
		return nil, fmt.Errorf("unknown packet type")
	}
}
