package dump

import (
	constants "github.com/mgeri/udptunneler/pkg"
	"github.com/mgeri/udptunneler/pkg/util"
	"github.com/spf13/cobra"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"strings"
)

var (
	udpInterface  string
	udpAddress    string
	serverAddress string

	Cmd = &cobra.Command{
		Use:   "dump",
		Short: "Dump UDP multicast traffic",
		Long:  ``,
		RunE:  dump,
	}
)

func init() {

	Cmd.PersistentFlags().StringVarP(&udpInterface, "interface", "i", "",
		"the network interface used to join the provided multicast channel provided")
	Cmd.PersistentFlags().StringVarP(&udpAddress, "address", "a", "",
		"the udp destination IP and port of the channel we want to join")

	_ = Cmd.MarkPersistentFlagRequired("interface")
	_ = Cmd.MarkPersistentFlagRequired("address")

}

func dump(cmd *cobra.Command, args []string) error {

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

	var buffer = make([]byte, constants.MaxDatagramSize)

	// Loop forever reading from the socket
	for {

		numBytes, cm, srcAddr, err := packetConn.ReadFrom(buffer)
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

		log.Printf(strings.Repeat("-", 80))
		log.Printf("addr: %v, numBytes: %d\n", srcAddr, numBytes)
		util.DumpByteSlice(buffer[:numBytes])
	}
}
