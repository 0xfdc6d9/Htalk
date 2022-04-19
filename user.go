package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
}

// NewUser 创建一个新用户
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String() // 用户的ip:port

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
	}

	go user.ListenMessage() // 启动用户的监听

	return user
}

// ListenMessage 用户监听channel中是否有信息，若有信息则写到客户端中
func (u *User) ListenMessage() {
	for {
		select {
		case msg := <-u.C:
			u.conn.Write([]byte(msg + "\n"))
		default:

		}
	}
}
