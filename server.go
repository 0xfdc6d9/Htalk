package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
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
	user := NewUser(conn, srv)

	user.Online()

	isLive := make(chan bool) // 判断用户是否在线

	// 广播客户端发送的消息（为了做超时强踢，将这部分异步出去）
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("conn.Read(buf) occurs an error: ", err)
				return
			}

			isLive <- true
			// 去掉最后的回车
			msg := string(buf[:n-1])

			// 广播用户输入的消息
			user.DoMessage(msg)
		}
	}()

	for {
		select {
		case <-isLive:

		case <-time.After(time.Second * 30):
			// 超时强踢
			user.C <- "你被踢了"
			time.Sleep(1 * time.Second)
			close(user.C)
			close(isLive)

			err := conn.Close() // 关闭conn会使server读buf的数据长度为0，触发Offline
			if err != nil {
				fmt.Println("conn.Close() occurs an error: ", err)
				return
			}
			return
		}
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
