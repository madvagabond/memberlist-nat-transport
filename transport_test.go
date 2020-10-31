package nat_transport

import (
	"testing"
	"time"
	"fmt"
 )






func TestConnection(t *testing.T) {
	a, e := NewNatTransport("127.0.0.1", 9000)
	b, e1 := NewNatTransport("127.0.0.1", 9001)

	if e != nil || e1 != nil {
		t.Error(e, e1)
		return 
	}


	fmt.Println(a.Address())
	fmt.Println(b.Address())


	_, e = a.DialTimeout(b.Address(), 10 * time.Second)


	if e != nil {
		t.Error(e)
		return
	}

	fmt.Println("connected")

	
}

