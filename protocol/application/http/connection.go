package http

import (
	"fmt"
	"log"

	"github.com/brewlin/net-protocol/pkg/waiter"
	tcpip "github.com/brewlin/net-protocol/protocol"
)

type connection struct {
	// 客户端连接的socket
	socket tcpip.Endpoint
	// 状态码
	status_code int
	// 接收队列
	recv_buf string
	// HTTP请求
	request *http_request
	// HTTP响应
	response *http_response
	// 接收状态
	recv_state http_recv_state
	// 客户端地址信息
	addr *tcpip.FullAddress
	// 请求长度
	request_len int
	// 请求文件的真实路径
	real_path string

	q         *waiter.Queue
	waitEntry waiter.Entry
	notifyC   chan struct{}
}

//等待并接受新的连接
func newCon(e tcpip.Endpoint, q *waiter.Queue) *connection {
	var con connection
	//创建结构实例
	con.status_code = 0
	con.request_len = 0
	con.socket = e
	con.real_path = ""
	con.recv_state = HTTP_RECV_STATE_WORD1
	con.request = newRequest()
	con.response = newResponse()
	con.recv_buf = ""
	addr, _ := e.GetRemoteAddress()
	log.Println("@application http: new client connection : ", addr)
	con.addr = &addr
	con.waitEntry, con.notifyC = waiter.NewChannelEntry(nil)
	q.EventRegister(&con.waitEntry, waiter.EventIn)
	con.q = q
	return &con

}

//HTTP 请求处理主函数
//从socket中读取数据并解析http请求
//解析请求
//发送响应
//记录请求日志
func (con *connection) handler() {
	<-con.notifyC
	log.Println("@应用层 http: waiting new event trigger ...")
	fmt.Println("@应用层 http: waiting new event trigger ...")
	for {
		v, _, err := con.socket.Read(con.addr)
		if err != nil {
			if err == tcpip.ErrWouldBlock {
				break
			}
			log.Println("@应用层 http:tcp read  got error", err)
			break
		}
		con.recv_buf += string(v)
	}
	fmt.Println("http协议原始数据:")
	fmt.Println(con.recv_buf)
	con.request.parse(con)
	con.response.send(con)
}

// 设置状态
func (c *connection) set_status_code(code int) {
	if c.status_code == 0 {
		c.status_code = code
	}
}

//关闭连接
func (c *connection) close() {
	if c == nil {
		return
	}
	//释放对应的请求
	c.request = nil
	c.response = nil
	//放放客户端连接中的缓存
	c.recv_buf = ""
	//注销接受队列
	c.q.EventUnregister(&c.waitEntry)
	c.socket.Close()
	//关闭socket连接
	c = nil
}
