package nat_transport

import (
	"github.com/anacrolix/utp"
	"github.com/hashicorp/memberlist"
	"net"
	"time"
	"github.com/ccding/go-stun/stun"
	ms "github.com/multiformats/go-multistream"
	"context"
	"fmt"
	"io"
	yamux "github.com/libp2p/go-yamux"
	"errors"
)




const (
	packetBufSize = 65536
	packet_proto = "/memberlist-nat/packet/v1"
	stream_proto = "/member-list-nat/v1"
)


type NatTransport struct {
	packets chan *memberlist.Packet
	streams chan net.Conn

	dialer dialer 
	utp *utp.Socket
	public_address string
}













func udpHolePunch(host string, port int) (*net.UDPConn, *stun.Host, error) {
	udp_addr, _ := net.ResolveUDPAddr("udp", host + ":" + string(port))
	conn, err := net.ListenUDP("udp", udp_addr)
	stun_client := stun.NewClientWithConnection(conn)
	stun_client.SetServerHost("stun.sipgate.net", 10000)
	_, s_host, err := stun_client.Discover()
	return conn, s_host, err
}







// port to config later
func NewNatTransport(host string, port int) (*NatTransport, error) {
	var t NatTransport
	
	pc, h, error := udpHolePunch(host, port)

	if error != nil {
		return &t, error
	}


	utp, err := utp.NewSocketFromPacketConn(pc)

	if err != nil {
		return &t, err
	}




	
	t.packets = make(chan *memberlist.Packet, 10)
	t.streams = make(chan net.Conn, 10)

	t.public_address = h.String()
	t.utp = utp
	t.dialer = newDialer()

	go t.listen()

	return &t, err 
}





func (t *NatTransport) StreamCh() <- chan net.Conn {
	return t.streams
}

func (t *NatTransport) PacketCh() <- chan *memberlist.Packet {
	return t.packets
}




func (t *NatTransport) Address() string {
	return t.public_address
}

func (t *NatTransport) WriteTo(b []byte, addr string) (time.Time, error) {
	e := t.sendPacket(b, addr)
	ts := time.Now()
	return ts, e
}



/*
Initiates a yamux over utp connection. 
Then selects the packet protocol via yamux 
sends the packet
Closes the yamux stream. 
*/

func (t *NatTransport) sendPacket(b []byte, addr string) error {
	conn, e := t.dialer.dial(addr)
	stream := ms.NewMSSelect(conn, packet_proto)
	_, e = stream.Write(b)
	
	stream.Close()
	return e 	
}




func (t *NatTransport) DialTimeout(addr string, timeout time.Duration) (net.Conn, error) {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	


	dial := func () (net.Conn, error){
		



		conn, e := t.dialer.dial(addr)

		if e != nil {
			conn.Close()
			return conn, e
		}
		
		err := ms.SelectProtoOrFail(stream_proto, conn)
	
	
		return conn, err
	}

	//return dial()

	return ctx_helper(ctx, dial)
}







func (t *NatTransport) handlePacket(s string, rwc io.ReadWriteCloser) error {

	defer rwc.Close()

	conn, err := rwc.(net.Conn)

	if !err {
		return errors.New("type mismatch") 
	}
	
	buf := make([]byte, packetBufSize)
	i, e := conn.Read(buf)

	if e != nil {
		return e
	}
	
	addr := conn.RemoteAddr()

	ts := time.Now()

	t.packets <- &memberlist.Packet{
		From: addr,
		Buf: buf[:i],
		Timestamp: ts,
	}

	return nil

}





func (t *NatTransport) handleStream(proto string, rwc io.ReadWriteCloser) error {
	conn := rwc.(net.Conn)

	fmt.Println("accepted stream")


	t.streams <- conn
	return nil
}







func (t *NatTransport) streamListener(conn net.Conn) {
	session, e := yamux.Server(conn, yamux.DefaultConfig())



	if e != nil {
		conn.Close()
		return 
	}


	

	mux := ms.NewMultistreamMuxer()

	mux.AddHandler(packet_proto, t.handlePacket)
	mux.AddHandler(stream_proto, t.handleStream)



	
	for {


	

		stream, err := session.Accept()


		if err != nil {
			session.Close()
			return 
		}

		
		//maybe implement error channel
		go mux.Handle(stream) 
	}

}



func (t *NatTransport) listen() error {





	for {

		conn, e := t.utp.Accept()

		if e != nil {
			return e 
		}




	
		go t.streamListener(conn)
	}
	
}


