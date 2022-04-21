package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

// NewClient 创建客户端对象
func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		Name:       "",
		conn:       nil,
		flag:       -1,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial() occurs an error: ", err)
		return nil
	}
	client.conn = conn

	return client
}

// DealResponse 处理server回应的消息，直接显示到标准输出
func (c *Client) DealResponse() {
	io.Copy(os.Stdout, c.conn) // 永久阻塞监听
}

func (c *Client) menu() bool {
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	var flag int
	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		c.flag = flag
		return true
	} else {
		fmt.Println(">>>> 请输入合法范围内的数字 <<<<")
		return false
	}
}

func (c *Client) PrivateChat() {
	PrintOnlineUser := func() {
		_, err := c.conn.Write([]byte("who\n"))
		if err != nil {
			fmt.Println("c.conn.Write() occurs an error: ", err)
			return
		}
	}
	PrintOnlineUser()

	var remoteName string
	var msg string
	fmt.Println(">>>>> 请输入聊天对象[用户名]，exit退出：")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>> 请输入消息内容，exit退出：")
		fmt.Scanln(&msg)
		for msg != "exit" {
			if len(msg) > 0 {
				_, err := c.conn.Write([]byte("to|" + remoteName + "|" + msg + "\n"))
				if err != nil {
					fmt.Println("c.conn.Write() occurs an error: ", err)
					return
				}
			}

			fmt.Println(">>>>> 请输入消息内容，exit退出：")
			fmt.Scanln(&msg)
		}

		fmt.Println(">>>>> 请输入聊天对象[用户名]，exit退出：")
		fmt.Scanln(&remoteName)
	}
}

func (c *Client) PublicChat() {
	var msg string

	fmt.Println(">>>>> 请输入聊天内容：")
	fmt.Scanln(&msg)

	for msg != "exit" {
		if len(msg) > 0 {
			_, err := c.conn.Write([]byte(msg + "\n")) // 发送到服务器中
			if err != nil {
				fmt.Println("c.conn.Write() occurs an error: ", err)
				return
			}
		}

		msg = ""
		println(">>>>> 请输入聊天内容：")
		fmt.Scanln(&msg)
	}
}

func (c *Client) UpdateName() bool {
	fmt.Println(">>>>> 请输入用户名：")
	fmt.Scanln(&c.Name)

	msg := fmt.Sprintf("rename|%s", c.Name)
	_, err := c.conn.Write([]byte(msg + "\n"))
	if err != nil {
		fmt.Println("c.conn.Write() occurs an error: ", err)
		return false
	}
	return true
}

func (c *Client) Run() {
	for c.flag != 0 {
		for !c.menu() { // 直到输入合法值
		}

		// 根据不同模式进行不同业务
		switch c.flag {
		case 1:
			// 公聊模式
			c.PublicChat()
		case 2:
			// 私聊模式
			c.PrivateChat()
		case 3:
			// 更新用户名
			c.UpdateName()
		}
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口")
}

func main() {
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>> 链接服务器失败...")
		return
	}
	fmt.Println(">>>>> 链接服务器成功...")

	go client.DealResponse()

	client.Run()
}
