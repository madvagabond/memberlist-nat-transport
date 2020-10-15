package memberlist_nat_transport

import (
	"github.com/anacrolix/utp"
	"github.com/hashicorp/memberlist"
	"net"
	"time"
	"github.com/ccding/go-stun/stun"
	"context"
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
	address ma.Multiaddr
	ctx context.Context
}




func NewNatTransport() {
	
}


func (t *NatTransport) bind(s string) {
}


func udpHolePunch(host string, port int) (*net.UDPConn, *stun.Host, error) {
	udp_addr, _ := net.ResolveUDPAddr("udp", host + ":" + string(port))
	conn, err := net.ListenUDP("udp", udp_addr)
	stun_client := stun.NewClientWithConnection(conn)
	s_host, err := stun_client.Keepalive()
	return conn, s_host, err
}




func (t *NatTransport) StreamCh() <- chan net.Conn {
	return t.streams
}

func (t *NatTransport) PacketCh() <- chan *memberlist.Packet {
	return t.packets
}


func (t *NatTransport) WriteTo(b []byte, addr string) (time.Time, error) {
	maddr, e := ma.NewMultiaddr(addr)
	ip, e := maddr.ValueForProtocol(ma.P_IP4)
	port, e := maddr.ValueForProtocol(ma.P_UDP)

	d_addr := ip + ":" + port
	udp_addr, e := net.ResolveUDPAddr("udp", d_addr)
	_, e = t.udp.WriteTo(b, udp_addr)
	ts := time.Now()
	return ts, e
}


func (t *NatTransport) DialTimeout(addr string, timeout time.Duration) (net.Conn, error) {
	maddr, e := ma.NewMultiaddr(addr)
	ip, e := maddr.ValueForProtocol(ma.P_IP4)

	port, e := maddr.ValueForProtocol(ma.P_UDP)
	addr = ip + ":" + port
	ctx, _ := context.WithTimeout(context.Background(), timeout)

	if e != nil {
		var nc utp.Conn
		return &nc, e 
	}

	
	return utp.DialContext(ctx, addr)
}



func (t *NatTransport) udpListen() {
	buf := make([]byte, packetBufSize)

	for {
		i, addr, e := t.udp.ReadFrom(buf)

		//add logging 
		if e != nil {
			break 
		}

		ts := time.Now()
		
		t.packets <- &memberlist.Packet {
			From: addr, Buf: buf[:i], Timestamp: ts}
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










