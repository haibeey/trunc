package connection

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"runtime"
	"time"
	"flag"
)

var (
	//default header for all request
	header       []byte
	byteReadSize = 20
)

type Conn struct {
	id       uint64
	conn     net.Conn
	fetching bool
	done     bool
}

func (c *Conn) GetId() uint64 {
	return c.id
}
//*service, *resource, *username, *password, *concLevel, *protocol
func (c *Conn) Connect() {
	service:=flag.Lookup("service").Value.(flag.Getter).Get().(string)

	tcpAddr, err := net.ResolveTCPAddr("tcp4", flag.Lookup("service").Value.(flag.Getter).Get().(string))
	if err != nil {
		panic(err)

	}
	usetls:=flag.Lookup("tls").Value.(flag.Getter).Get().(bool)

	c.conn, err = resolveConn(usetls, service, *tcpAddr)

	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	_, err = c.conn.Write(header)
	if err != nil {
		panic(err)
	}
}

func resolveConn(dotls bool, host string, tcpAddr net.TCPAddr) (net.Conn, error) {

	if !dotls {
		return net.DialTCP("tcp", nil, &tcpAddr)
	}

	data, err := ioutil.ReadFile(flag.Lookup("certpath").Value.(flag.Getter).Get().(string))
	if err != nil {
		panic(err)
	}
	
	certs := x509.NewCertPool()
	certs.AppendCertsFromPEM(data)

	return tls.Dial("tcp", host, &tls.Config{
		RootCAs: certs,
	})
}

func NewConn(id uint64) *Conn {
	conn := &Conn{
		id: id,
	}
	conn.Connect()
	return conn
}

//read data in chucks
//returns true is we reach end of stream
func (c *Conn) Read() {
	//haven't finish reading previous data. abort
	if c.fetching {
		return
	}
	c.fetching = true
	data := make([]byte, byteReadSize)
	_, err := c.conn.Read(data)
	if err == io.EOF {
		fmt.Println("Conn with id %v return error %v", c.GetId(), err)
		c.done = true
	}
	//clear data
	c.fetching = false

}

func (c *Conn) Write(b []byte) {
	_, err := c.conn.Write(b)
	if err != nil {
		panic(err)
	}

}

func (c *Conn) Close() {
	c.conn.Close()
}

type ConnGroup struct {
	Connections map[uint64]*Conn
	peek        uint64
	reads       int
}

func NewConnGroup(rsr string, encAuth string) *ConnGroup {

	header = buildByteBuffer(rsr, encAuth)
	return &ConnGroup{
		Connections: make(map[uint64]*Conn),
	}
}
func buildByteBuffer(rsr string, encAuth string) []byte {
	service:=flag.Lookup("service").Value.(flag.Getter).Get().(string)
	var buf bytes.Buffer
	buf.Write([]byte("GET "))
	buf.Write([]byte(rsr))
	buf.Write([]byte(" HTTP/1.0"))
	buf.Write([]byte(newLineSep()))
	buf.Write([]byte("Host: " + service))
	buf.Write([]byte(newLineSep()))
	if encAuth != "" {
		buf.Write([]byte("Authorization: Basic "))
		buf.Write([]byte(encAuth))
		buf.Write([]byte(newLineSep()))
	}
	return buf.Bytes()
}

func newLineSep() string {
	if runtime.GOOS == "windows" {
		return "\r\n\r\n"
	}
	return "\n\n"
}
func (cg *ConnGroup) AddConn(c *Conn) {
	cg.peek++
	cg.Connections[cg.peek] = c
}

func (cg *ConnGroup) Read() {
	for _, conn := range cg.Connections {
		go conn.Read()
		time.Sleep(time.Second * 1)
		cg.reads++
	}
}

func (cg *ConnGroup) Done() bool {
	for _, conn := range cg.Connections {
		if !conn.done {
			return false
		}
	}
	return true
}

func (cg *ConnGroup) Release(connId uint64) {
	delete(cg.Connections, connId)
}

func (cg *ConnGroup) ReleaseAll() {
	for _, c := range cg.Connections {
		c.Close()
		delete(cg.Connections, c.GetId())
	}
}
