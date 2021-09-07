package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatalln(err)
	}
	defer listener.Close()

	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConn(conn)
	}
}

//代表每個客戶端
type client chan<- string

var (

	//客戶端進入
	entering = make(chan client)

	//客戶端離開
	leaving = make(chan client)

	//聊天室廣播
	message = make(chan string)
)

func broadcaster() {
	clients := make(map[client]bool) //每個client的登入狀態
	for {
		select {
		case msg := <-message:
			for cli := range clients {
				cli <- msg
			}

		case cli := <-entering:
			clients[cli] = true

		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}

	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	defer conn.Close()
	
	go writeToClient(conn, ch)
	who := conn.RemoteAddr().String()
	ch <- who + "歡迎你進入聊天室"

	message <- who + "已進入聊天室"

	entering <- ch

	input := bufio.NewScanner(conn)

	//監聽客戶端輸入
	for input.Scan() {
		message <- who + input.Text()
	}

	//客戶端斷開連接
	leaving <- ch
	message <- who + "離開聊天室"

}

//系統發送給連入者的訊息
func writeToClient(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}
