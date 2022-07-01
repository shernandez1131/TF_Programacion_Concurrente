package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

type Product struct {
	ID   string `json:"ID"`
	Name string `json:"Name"`
}

type Frame struct {
	Cmd    string   `json:"cmd"`
	Sender string   `json:"sender"`
	Data   []string `json:"data"`
}

type Info struct {
	nextNode string
	nextNum  int
	imFirst  bool
	cont     int
}

type InfoCons struct {
	contA, contB int
}

var (
	host         string
	chRemotes    chan []string
	myNum        int
	chInfo       chan Info
	chCons       chan InfoCons
	chProd       chan Product
	readyToStart chan bool
	participants int
)

func main() {
	rand.Seed(time.Now().UnixNano())
	if len(os.Args) == 1 {
		log.Println("Hostname not given")
	} else {
		chRemotes = make(chan []string, 1)
		chInfo = make(chan Info, 1)
		chCons = make(chan InfoCons, 1)
		chProd = make(chan Product, 1)
		readyToStart = make(chan bool)

		host = os.Args[1]
		chRemotes <- []string{}

		if len(os.Args) >= 3 {
			connectToNode(os.Args[2])
		}
		if len(os.Args) == 4 {
			switch os.Args[3] {
			case "agrawalla":
				go startAgrawalla()
			case "consensus":
				go startConsensus()
			}

		}
		server()
	}
}

func startAgrawalla() {
	go func() {
		time.Sleep(5 * time.Second)
		remotes := <-chRemotes
		for _, remote := range remotes {
			send(remote, Frame{"agrawalla", host, []string{}}, nil)
		}
		chRemotes <- remotes
		handleAgrawalla()
	}()
}

func startConsensus() {
	remotes := <-chRemotes
	for _, remote := range remotes {
		log.Printf("%s: sending consensus batsignal to %s\n", host, remote)
		send(remote, Frame{"consensus", host, []string{}}, nil)
	}
	chRemotes <- remotes
	handleConsensus()
}

func connectToNode(remote string) {
	remotes := <-chRemotes
	remotes = append(remotes, remote)
	chRemotes <- remotes
	if !send(remote, Frame{"hello", host, []string{}}, func(cn net.Conn) {
		dec := json.NewDecoder(cn)
		var frame Frame
		dec.Decode(&frame)
		remotes := <-chRemotes
		remotes = append(remotes, frame.Data...)
		chRemotes <- remotes
		log.Printf("%s: friends: %s\n", host, remotes)
	}) {
		log.Printf("%s: unable to connect to %s\n", host, remote)
	}
}

func send(remote string, frame Frame, callback func(net.Conn)) bool {
	if cn, err := net.Dial("tcp", remote); err == nil {
		defer cn.Close()
		enc := json.NewEncoder(cn)
		enc.Encode(frame)
		if callback != nil {
			callback(cn)
		}
		return true
	} else {
		log.Printf("%s: can't connect to %s\n", host, remote)
		idx := -1
		remotes := <-chRemotes
		for i, rem := range remotes {
			if remote == rem {
				idx = i
				break
			}
		}
		if idx >= 0 {
			remotes[idx] = remotes[len(remotes)-1]
			remotes = remotes[:len(remotes)-1]
		}
		chRemotes <- remotes
		return false
	}
}

func server() {
	if ln, err := net.Listen("tcp", host); err == nil {
		defer ln.Close()
		log.Printf("Listening on %s\n", host)
		for {
			if cn, err := ln.Accept(); err == nil {
				go fauxDispatcher(cn)
			} else {
				log.Printf("%s: cant accept connection.\n", host)
			}
		}
	} else {
		log.Printf("Can't listen on %s\n", host)
	}
}

func fauxDispatcher(cn net.Conn) {
	defer cn.Close()
	dec := json.NewDecoder(cn)
	frame := &Frame{}
	dec.Decode(frame)
	switch frame.Cmd {
	case "hello":
		handleHello(cn, frame)
	case "add":
		handleAdd(frame)
	case "agrawalla":
		handleAgrawalla()
	case "num":
		handleNum(frame)
	case "start":
		handleStart()
	case "consensus":
		handleConsensus()
	case "vote":
		handleVote(frame)
	}
}

func handleHello(cn net.Conn, frame *Frame) {
	enc := json.NewEncoder(cn)
	remotes := <-chRemotes
	enc.Encode(Frame{"<response>", host, remotes})
	notification := Frame{"add", host, []string{frame.Sender}}
	for _, remote := range remotes {
		send(remote, notification, nil)
	}
	remotes = append(remotes, frame.Sender)
	log.Printf("%s: friends: %s\n", host, remotes)
	chRemotes <- remotes
}
func handleAdd(frame *Frame) {
	remotes := <-chRemotes
	remotes = append(remotes, frame.Data...)
	log.Printf("%s: friends: %s\n", host, remotes)
	chRemotes <- remotes
}
func handleAgrawalla() {
	myNum = rand.Intn(1000000000)
	log.Printf("%s: my number is %d\n", host, myNum)
	msg := Frame{"num", host, []string{strconv.Itoa(myNum)}}
	remotes := <-chRemotes
	for _, remote := range remotes {
		send(remote, msg, nil)
	}
	chRemotes <- remotes
	chInfo <- Info{"", 1000000001, true, 0}
}
func handleNum(frame *Frame) {
	if num, err := strconv.Atoi(frame.Data[0]); err == nil {
		info := <-chInfo
		if num > myNum {
			if num < info.nextNum {
				info.nextNum = num
				info.nextNode = frame.Sender
			}
		} else {
			info.imFirst = false
		}
		info.cont++
		go func() { chInfo <- info }()
		remotes := <-chRemotes
		chRemotes <- remotes
		if info.cont == len(remotes) {
			if info.imFirst {
				log.Printf("%s: I'm first!\n", host)
				criticalSection()
			} else {
				readyToStart <- true
			}
		}
	} else {
		log.Printf("%s: can't convert %v\n", host, frame)
	}
}
func handleStart() {
	<-readyToStart
	criticalSection()
}
func handleConsensus() {
	time.Sleep(3 * time.Second)
	fmt.Print("Escribe cualquier cosa para agregar un producto:\n")
	var op string
	fmt.Scanf("%s\n", &op)
	info := InfoCons{0, 0}
	getProduct := Product{"", ""}
	getProduct.Name = op
	/*if op == "A" {
		info.contA++
	} else {
		info.contB++
	}*/
	chCons <- info
	chProd <- getProduct
	remotes := <-chRemotes
	participants = len(remotes) + 1
	for _, remote := range remotes {
		log.Printf("%s: sending %s to %s\n", host, op, remote)
		send(remote, Frame{"vote", host, []string{op}}, nil)

		file, err := os.OpenFile("sample_products.csv", os.O_APPEND|os.O_CREATE, os.ModePerm)
		defer file.Close()
		if err != nil {
			panic(err)
		}

		w := csv.NewWriter(file)
		f, _ := file.Stat()
		if f.Size() == 0 {
			w.Write([]string{"ID", "Name"})
		}

		errData := w.Write([]string{"80", op})
		//fmt.Println([]string{"80",op})
		if errData != nil {
			log.Println(errData)
		}

		w.Flush()

	}
	chRemotes <- remotes
}
func handleVote(frame *Frame) {
	vote := frame.Data[0]
	info := <-chCons
	if vote == "A" {
		info.contA++
	} else {
		info.contB++
	}
	chCons <- info
	log.Printf("%s: got %s from %s\n", host, vote, frame.Sender)
	if info.contA+info.contB == participants {
		if info.contA > info.contB {
			log.Printf("%s go A\n", host)
		} else {
			log.Printf("%s go B\n", host)
		}
	}
}

func criticalSection() {
	log.Printf("%s: my time has come!\n", host)
	info := <-chInfo
	if info.nextNode != "" {
		log.Printf("%s: letting %s start\n", host, info.nextNode)
		send(info.nextNode, Frame{"start", host, []string{}}, nil)
	} else {
		log.Printf("%s: I was the last one :(\n", host)
	}
}
