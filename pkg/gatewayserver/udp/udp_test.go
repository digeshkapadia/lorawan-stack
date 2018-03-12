// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package udp_test

import (
	"fmt"
	"net"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/udp"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

var (
	downlink ttnpb.DownlinkMessage

	extracter    = dummyExtractor{}
	gatewayStore = udp.NewGatewayStore(udp.DefaultWaitDuration)
)

type dummyExtractor struct{}

func (d dummyExtractor) RxPacket(p udp.RxPacket) (ttnpb.UplinkMessage, error) {
	return ttnpb.UplinkMessage{}, nil
}
func (d dummyExtractor) Status(p udp.Stat) (ttnpb.GatewayStatus, error) {
	return ttnpb.GatewayStatus{}, nil
}
func (d dummyExtractor) TxPacket(downlink ttnpb.DownlinkMessage) (udp.Packet, error) {
	return udp.Packet{}, nil
}
func (d dummyExtractor) TxPacketAck(p udp.TxPacketAck) (ttnpb.UplinkMessage, error) {
	return ttnpb.UplinkMessage{}, nil
}

func Example() {
	udpAddr, err := net.ResolveUDPAddr("udp", ":1700")
	if err != nil {
		panic(err)
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		panic(err)
	}

	conn := udp.Handle(udpConn, gatewayStore, nil)

	go func() {
		var packet *udp.Packet
		for {
			packet, err = conn.Read()
			if err != nil {
				panic(err)
			}
			err := packet.Ack()
			if err != nil {
				panic(err)
			}

			switch packet.PacketType {
			case udp.PushData:
				if packet.Data == nil {
					continue
				}
				// Handle the data
				for _, packet := range packet.Data.RxPacket {
					uplink, err := extracter.RxPacket(*packet)
					if err != nil {
						continue
					}

					Forward(uplink)
				}
				if packet.Data.Stat != nil {
					status, err := extracter.Status(*packet.Data.Stat)
					if err != nil {
						continue
					}

					Forward(status)
				}
			case udp.TxAck:
				// Handle the data
				if packet.Data.TxPacketAck != nil {
					txInfo, err := extracter.TxPacketAck(*packet.Data.TxPacketAck)
					if err != nil {
						continue
					}

					Forward(txInfo)
				}
			}
		}
	}()

	go func() {
		time.Sleep(10 * time.Second)
		packet, err := extracter.TxPacket(downlink)
		if err != nil {
			fmt.Println("Couldn't convert Tx packet")
			return
		}

		if err := conn.Write(&packet); err != nil {
			fmt.Println("Error when sending downlink: ", err)
		}
	}()
}

func Forward(interface{}) {}