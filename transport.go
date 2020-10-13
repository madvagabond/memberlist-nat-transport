package memberlist-nat-transport


import (
	"github.com/anacrolix/utp"
	"github.com/hashicorp/memberlist"
	"net"
	"time"
	"github.com/ccding/go-stun/stun"
	ma "github.com/multiformats/go-multiaddr"
)




const (
	packetBufSize = 65536
)


type NatTransport struct {
	packets chan *memberlist.Packet
	streams chan net.Conn
	utp utp.Socket
	udp net.UDPConn
	public_address ma.Multiaddr
	shutdownCh chan struct{}
}



func udpHolePunch(host string, port int) (net.UDPConn, stun.Host, error) {
	conn, err := net.ListenUDP(host, port)
	stun_client := stun.NewClientWithConnection(packet_conn)
	host, err := stun_client.Keepalive()
	return conn, host, err
}




func (t *NatTransport) StreamCh() <- chan net.Conn {
	return t.streams
}

func (t *NatTransport) PacketCh() <- chan *memberlist.Packet {
	return t.packets
}



func (t *NatTransport) udpListen() {
	buf := make([]byte, packetBufSize)

	for {
		i, addr, e := t.udp.ReadFrom(buf)

		//add logging 
		if err != nil {
			break 
		}

		ts := time.Now()
		
		t.packets <- &Packet {
			Address: addr, Buf: buf[:i], Timestamp: ts
		}
	}
}


func (t *NatTransport) utpListen() {

	for {
		conn, err := t.utp.Accept()

		if err != nil {
			break 
		}

		t.streams <- conn
	}
}










