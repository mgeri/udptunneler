package ping

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"net"
	"time"
)

var (
	udpAddress string

	Cmd = &cobra.Command{
		Use:   "ping",
		Short: "Send multicast 'Hello world' ping message. Can be used for testing the multicast channel.",
		Long:  ``,
		RunE:  ping,
	}
)

func init() {

	Cmd.PersistentFlags().StringVarP(&udpAddress, "address", "a", "",
		"the udp destination IP and port of the channel we want to join")

	_ = Cmd.MarkPersistentFlagRequired("address")
}

func ping(cmd *cobra.Command, args []string) error {
	addr, err := net.ResolveUDPAddr("udp4", udpAddress)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		return err
	}

	log.Printf("starting ping loop to %s", udpAddress)
	count := 0
	for {
		count++
		log.Printf("sending ping [%d]...", count)
		_, err := conn.Write([]byte(fmt.Sprintf("hello, world [%d]", count)))
		if err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}
}
