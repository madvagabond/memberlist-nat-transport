package memberlist_nat_transport


import (
	"sync"
	"github.com/anacrolix/utp"
	"net"
	yamux "github.com/libp2p/go-yamux"
	"context"
)


type dialer struct {
	mu sync.RWMutex
	conns map[string]*yamux.Session
}


func (d *dialer) getConn(addr string) (*yamux.Session, bool) {
	defer d.mu.RUnlock()
	d.mu.RLock()
	c, b := d.conns[addr]
	return c, b
} 


func newDialer() dialer {
	var mu sync.RWMutex
	conns := make(map[string]*yamux.Session)
	return dialer{mu, conns}
}






func ctx_helper(ctx context.Context, f func() (net.Conn, error)) (net.Conn, error) {
	



	type result struct {
		conn net.Conn
		error error
	}

	c := make(chan result)
	
	fun := func() {
		conn, e :=  f()
		c <- result{conn, e}
	}

	go fun()


	select {
	case <- ctx.Done():
		var c net.Conn
		return c, ctx.Err()  

	case result := <- c:
		return result.conn, result.error
	}
}






func dialTransport(addr string) (*yamux.Session, error) {
	conn, e := utp.Dial(addr)

	if e != nil {
		return nil, e
	}

	return yamux.Client(conn, yamux.DefaultConfig())
}




func (d *dialer) dial(addr string) (net.Conn, error) {
	
	if conn, ok := d.getConn(addr); ok {
		return conn.Open()
	}


	session, e := dialTransport(addr)
	
	if e != nil {
		return nil, e
	}


	d.mu.Lock()

	d.conns[addr] = session
	d.mu.Unlock()

	return session.Open()
}

