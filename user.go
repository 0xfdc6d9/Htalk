package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// NewUser 创建一个新用户
func NewUser(conn net.Conn, srv *Server) *User {
	userAddr := conn.RemoteAddr().String() // 用户的ip:port

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: srv,
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
	// 修改用户名消息格式：rename|NewUsername
	if len(msg) > 7 && msg[:7] == "rename|" {
		NewUsername := strings.Split(msg, "|")[1]
		// 一锁二查三更新
		u.server.OnlineMapLock.Lock()
		_, ok := u.server.OnlineMap[NewUsername]
		if ok {
			u.C <- "当前用户名已被使用"
		} else {
			delete(u.server.OnlineMap, u.Name)
			u.server.OnlineMap[NewUsername] = u
			u.Name = NewUsername
			str := fmt.Sprintf("您已更新用户名为：%s", u.Name)
			u.C <- str
		}
		u.server.OnlineMapLock.Unlock()
		return
	}
	switch msg {
	case "who":
		// 查询当前用户所在的server中有哪些用户在线
		for _, client := range u.server.OnlineMap {
			str := fmt.Sprintf("[%s]%s: 在线中", client.Addr, client.Name)
			u.C <- str
		}
	default:
		u.server.Broadcast(u, msg)
	}
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
