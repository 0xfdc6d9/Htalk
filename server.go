package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	OnlineMap     map[string]*User // 在线用户表
	OnlineMapLock sync.RWMutex

	Message chan string
}

// NewServer 创建一个服务器
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:            ip,
		Port:          port,
		OnlineMap:     make(map[string]*User),
		OnlineMapLock: sync.RWMutex{},
		Message:       make(chan string),
	}
	return server
}

// Broadcast 将待广播消息写入srv的channel中
func (srv *Server) Broadcast(user *User, msg string) {
	str := fmt.Sprintf("[%s]%s: %s", user.Addr, user.Name, msg)
	srv.Message <- str
}

func (srv *Server) Handler(conn net.Conn) {
	// 创建新用户
	user := NewUser(conn)

	// 将新用户添加到 OnlineMap中
	srv.OnlineMapLock.Lock()
	srv.OnlineMap[user.Name] = user
	srv.OnlineMapLock.Unlock()

	// 广播新用户上线消息
	srv.Broadcast(user, "已上线")

	// 广播客户端发送的消息（感觉这里可以不用并发）
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if n == 0 {
			srv.Broadcast(user, "已下线")
			return
		}
		if err != nil && err != io.EOF {
			fmt.Println("conn.Read(buf) occurs an error: ", err)
			return
		}

		// 去掉最后的回车
		msg := string(buf[:n-1])

		// 广播用户输入的消息
		srv.Broadcast(user, msg)
	}
}

// Start 启动服务器的接口
func (srv *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", srv.Ip, srv.Port))
	if err != nil {
		fmt.Println("net.Listen() occurs an error: ", err)
		return
	}

	// close listen socket
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("listener.Close() occurs an error: ", err)
		}
	}(listener)

	go srv.ListenMessage() // 监听新上线的用户消息

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept() occurs an error: ", err)
			continue
		}

		// 到这里则代表有新用户上线

		// do handler
		go srv.Handler(conn)
	}

}

// ListenMessage 监听需要广播的新用户上线信息
func (srv *Server) ListenMessage() {
	for {
		select {
		case msg := <-srv.Message:
			srv.OnlineMapLock.Lock()
			for _, client := range srv.OnlineMap {
				client.C <- msg
			}
			srv.OnlineMapLock.Unlock()
		default:

		}
	}
}
