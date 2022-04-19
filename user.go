package main

import (
	"fmt"
	"net"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// NewUser 创建一个新用户
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String() // 用户的ip:port

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	go user.ListenMessage() // 启动用户的监听

	return user
}

func (u *User) Online() {
	// 将新用户添加到OnlineMap中
	u.server.OnlineMapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.OnlineMapLock.Unlock()

	// 广播新用户上线消息
	u.server.Broadcast(u, "已上线")
}

func (u *User) Offline() {
	// 将新用户OnlineMap中删除
	u.server.OnlineMapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.OnlineMapLock.Unlock()

	// 广播新用户下线消息
	u.server.Broadcast(u, "已下线")
}

func (u *User) DoMessage(msg string) {
	u.server.Broadcast(u, msg)
}

// ListenMessage 用户监听channel中是否有信息，若有信息则写到客户端中
func (u *User) ListenMessage() {
	for {
		select {
		case msg := <-u.C:
			_, err := u.conn.Write([]byte(msg + "\n"))
			if err != nil {
				fmt.Println("u.conn.Write([]byte(msg + \"\\n\")) occurs an error", err)
				return
			}
		default:

		}
	}
}
