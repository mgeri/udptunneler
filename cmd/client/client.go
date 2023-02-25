package client

import (
	"bufio"
	"github.com/bytedance/gopkg/lang/mcache"
	constants "github.com/mgeri/udptunneler/pkg"
	"github.com/mgeri/udptunneler/pkg/frame"
	"github.com/mgeri/udptunneler/pkg/packet"
	"github.com/mgeri/udptunneler/pkg/util"
	"github.com/spf13/cobra"
	"golang.org/x/net/ipv4"
	"time"

	"log"
	"net"
	"strings"
)

var (
	udpInterface  string
	udpAddress    string
	serverAddress string
	dumpBytes     bool

	Cmd = &cobra.Command{
		Use:   "client",
		Short: "Start UDP tunneler client",
		Long:  ``,
		RunE:  client,
	}
)

func init() {

	Cmd.PersistentFlags().StringVarP(&udpInterface, "interface", "i", "",
		"the network interface used to join the provided multicast channel provided")
	Cmd.PersistentFlags().StringVarP(&udpAddress, "address", "a", "",
		"the udp destination IP and port of the channel we want to join")
	Cmd.PersistentFlags().StringVarP(&serverAddress, "server", "s", "",
		"the tcp address (ip:port) of the server to which the datagram will be forwarded")
	Cmd.PersistentFlags().BoolVarP(&dumpBytes, "dump", "d", false,
		"dump the raw bytes of the message")

	_ = Cmd.MarkPersistentFlagRequired("interface")
	_ = Cmd.MarkPersistentFlagRequired("address")
	_ = Cmd.MarkPersistentFlagRequired("server")

}

func client(cmd *cobra.Command, args []string) error {
	// connect to server
	connServer, err := net.Dial("tcp", serverAddress)
	if err != nil {
		log.Printf("error connecting to server %s: %v", serverAddress, err)
		return err
	}
	defer connServer.Close()
	log.Printf("connected to server: [%s <-> %s]", connServer.RemoteAddr(), connServer.LocalAddr())

	dataChannel := make(chan *packet.Datagram, 1024)
	go handleServerConnection(connServer, dataChannel)

	// listen to udp channel
	addr, err := net.ResolveUDPAddr("udp4", udpAddress)
	if err != nil {
		return err
	}

	var intf *net.Interface = nil

	if udpInterface != "" {
		intf, err = net.InterfaceByName(udpInterface)
		if err != nil {
			return err
		}
	}

	conn, err := net.ListenPacket("udp4", udpAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	packetConn := ipv4.NewPacketConn(conn)
	if err := packetConn.JoinGroup(intf, addr); err != nil {
		return err
	}
	defer packetConn.LeaveGroup(intf, addr)

	err = packetConn.SetControlMessage(ipv4.FlagTTL|ipv4.FlagSrc|ipv4.FlagDst|ipv4.FlagInterface, true)
	if err != nil {
		return err
	}

	log.Printf("listening multicast to %s@%s  %v\n", udpAddress, util.StringIfEmpty(udpInterface, "default"), intf)

	var buffer []byte
	// Loop forever reading from the socket
	for {

		if buffer == nil {
			// leave space for the datagram header to avoid reallocation
			buffer = mcache.Malloc(constants.MaxDatagramSize + packet.DatagramPacketHeaderLen)
		}

		numBytes, cm, srcAddr, err := packetConn.ReadFrom(buffer[packet.DatagramPacketHeaderLen:])
		if err != nil {
			log.Fatal("read from udp failed:", err)
		}

		if !cm.Dst.IsMulticast() {
			continue
		}
		if !cm.Dst.Equal(addr.IP) {
			// unknown group, discard
			continue
		}

		if dumpBytes {
			log.Printf(strings.Repeat("-", 80))
			log.Printf("addr: %v, numBytes: %d\n", srcAddr, numBytes)
			util.DumpByteSlice(buffer[packet.DatagramPacketHeaderLen : packet.DatagramPacketHeaderLen+numBytes])
		}

		// send the datagram to the server
		d := packet.Datagram{
			DatagramLength: uint16(numBytes),
			UdpIP:          cm.Dst,
			UdpPort:        uint16(addr.Port),
			DatagramPacket: buffer,
		}
		dataChannel <- &d
		buffer = nil
	}
}

func handleServerConnection(conn net.Conn, in <-chan *packet.Datagram) {
	timer := time.NewTicker(time.Second * constants.DefaultHeartbeatTimeout / 2)

	go handleServerResponse(conn)

	frameCodec := frame.NewFrameCodec()
	wbuf := bufio.NewWriter(conn)

	heartbeatBuffer := make([]byte, packet.HeartbeatPacketHeaderLen)
	p := packet.Heartbeat{}
	p.Encode(heartbeatBuffer)

	for {
		select {
		case <-timer.C:
			err := frameCodec.Encode(wbuf, heartbeatBuffer)
			if err != nil {
				log.Fatalf("write error while sending heartbeat: %v", err)
				return
			}
			err = wbuf.Flush()
			if err != nil {
				log.Fatalf("write error while flushing: %v", err)
				return
			}
		case data := <-in:
			// unwrap buffer from packet to avoid encoding it (is already ready to be sent except for the header)
			buffer := data.DatagramPacket
			data.DatagramPacket = nil
			data.Encode(buffer)
			err := frameCodec.Encode(wbuf, buffer[:packet.DatagramPacketHeaderLen+data.DatagramLength])
			mcache.Free(buffer)
			if err != nil {
				log.Fatalf("write error while sending datagram: %v", err)
				return
			}
			err = wbuf.Flush()
			if err != nil {
				log.Fatalf("write error while flushing: %v", err)
				return
			}
		}
	}
}

func handleServerResponse(conn net.Conn) {
	frameCodec := frame.NewFrameCodec()
	rbuf := bufio.NewReader(conn)
	for {
		framePayload, err := frameCodec.Decode(rbuf)
		if err != nil {
			log.Fatalf("read error: %v", err)
			return
		}
		p, err := packet.Decode(framePayload)
		if err != nil {
			log.Fatalf("packet decode error: %v", err)
			return
		}
		switch p.(type) {
		case *packet.Heartbeat:
			log.Printf("heartbeat received")
		default:
			log.Fatalf("unknown packet received: %v", p)
		}
		mcache.Free(framePayload)
	}
}
