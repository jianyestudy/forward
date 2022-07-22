package main

import (
	"flag"
	"log"
	"net"
)

const LEN = 1024

func main() {
	var from, toForward, method string
	
	flag.StringVar(&from, "f", "", "-l=0.0.0.0:8282 指定服务监听的IP跟端口")
	flag.StringVar(&toForward, "t", "", "-d=127.0.0.1:1789 指定转发的IP和端口")
	flag.StringVar(&method, "m", "", "-m=tcp 指定转发的方式，tcp或udp或tcpv6")
	flag.Parse()
	
	if from == "" || toForward == "" || method == "" {
		log.Panic("参数不完整，例如：./forward -f=110.110.110.110:8282 -t=120.120.120.120:1789 -m=tcp")
	}
	
	fListen, err := net.Listen(method, from)
	
	if err != nil {
		log.Fatal("监听失败：" + err.Error())
	}
	
	defer func(f net.Listener) {
		err := f.Close()
		if err != nil {
			log.Fatal("关闭监听失败：" + err.Error())
		}
	}(fListen)
	
	for {
		log.Println("监听成功，等待连接...")
		
		fConn, err := fListen.Accept()
		if err != nil {
			log.Printf("接受来自%s的：请求失败："+err.Error(), fConn.RemoteAddr().String())
		} else {
			log.Printf("收到来自%s的：请求\n", fConn.RemoteAddr().String())
		}
		
		// 这边最好也做个协程，防止阻塞
		tConn, err := net.Dial("tcp", toForward)
		if err != nil {
			log.Println("连接转发目的地失败：" + err.Error())
			
			continue
		}
		
		go handleConnection(fConn, tConn)
		go handleConnection(tConn, fConn)
	}
}

func handleConnection(r, w net.Conn) {
	defer func(r net.Conn) {
		err := r.Close()
		if err != nil {
			log.Println("关闭连接失败：" + err.Error())
		}
	}(r)
	defer func(w net.Conn) {
		err := w.Close()
		if err != nil {
			log.Println("关闭连接失败：" + err.Error())
		}
	}(w)
	
	buffer := make([]byte, LEN)
	
	for {
		nLen, err := r.Read(buffer)
		if err != nil {
			log.Println("读取数据失败：" + err.Error())
			
			break
		}
		
		nLen, err = w.Write(buffer[:nLen])
		if err != nil {
			log.Println("写入数据失败：" + err.Error())
			
			break
		}
		
		log.Printf("%s->%s,成功转发%d字节数据\n", r.RemoteAddr().String(), w.RemoteAddr().String(), nLen)
	}
}
