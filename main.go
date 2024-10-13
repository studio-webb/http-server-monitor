package main

import (
	"context"
	"fmt"
	"github.com/coder/websocket"
	"github.com/studio-webb/http-server-monitor/internal/hardware"
	"log"
	"net/http"
	"sync"
	"time"
)

type subscriber struct {
	messages chan []byte
}

type Server struct {
	subscriberMessageBuffer int
	mux                     http.ServeMux
	subscribersMutex        sync.Mutex
	subscribers             map[*subscriber]struct{}
}

func NewServer() *Server {
	s := &Server{
		subscriberMessageBuffer: 10,
		subscribers:             make(map[*subscriber]struct{}),
	}

	s.mux.Handle("/", http.FileServer(http.Dir("./htmx")))
	s.mux.HandleFunc("/ws", s.subscriberHandler)
	return s
}

func (s *Server) subscriberHandler(write http.ResponseWriter, req *http.Request) {
	err := s.subscriber(req.Context(), write, req)
	if err != nil {
		log.Printf("subscriberHandler %v\n", err)
	}
}

func (s *Server) removeSubscriber(sub *subscriber) {
	s.subscribersMutex.Lock()
	defer s.subscribersMutex.Unlock()
	delete(s.subscribers, sub)
	close(sub.messages) // Закрываем канал, если это необходимо
	fmt.Println("Removed subscriber", sub)
}

func (s *Server) addSubscriber(subscriber *subscriber) {
	s.subscribersMutex.Lock()
	//defer s.subscribersMutex.Unlock()
	//s.subscribers[subscriber] = struct{}{}
	s.subscribers[subscriber] = struct{}{}
	s.subscribersMutex.Unlock()
	fmt.Println("Added subscriber", subscriber)
}

func (s *Server) subscriber(ctx context.Context, write http.ResponseWriter, req *http.Request) error {
	var c *websocket.Conn
	subscriber := &subscriber{
		messages: make(chan []byte, s.subscriberMessageBuffer),
	}
	s.addSubscriber(subscriber)
	defer s.removeSubscriber(subscriber)
	c, err := websocket.Accept(write, req, nil)
	if err != nil {
		return err
	}
	defer func(c *websocket.Conn) {
		err := c.CloseNow()
		if err != nil {
			log.Fatalf("%v\n", err)
		}
	}(c)
	c.CloseRead(ctx)
	for {
		select {
		case msg := <-subscriber.messages:
			ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			err := c.Write(ctx, websocket.MessageText, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *Server) broadcast(msg []byte) {
	s.subscribersMutex.Lock()
	for sub := range s.subscribers {
		select {
		case sub.messages <- msg:
			// Сообщение успешно отправлено
		default:
			// Канал заблокирован или заполнен, удаляем подписчика
			delete(s.subscribers, sub)
			close(sub.messages)
			fmt.Println("Removed subscriber due to full buffer", sub)
		}
	}
	s.subscribersMutex.Unlock()
}

func main() {
	fmt.Println("Print system monitor")
	srv := NewServer()
	go func(s *Server) {
		for {
			systemSection, err := hardware.GetSystemSection()
			if err != nil {
				log.Println(err)
			}

			diskSection, err := hardware.GetDiskSection()
			if err != nil {
				log.Println(err)
			}

			cpuSection, err := hardware.GetCPUSection()
			if err != nil {
				log.Println(err)
			}
			//fmt.Println(systemSection)
			s.broadcast([]byte(systemSection))
			//fmt.Println(diskSection)
			s.broadcast([]byte(diskSection))
			//fmt.Println(cpuSection)
			s.broadcast([]byte(cpuSection))

			timeStamp := time.Now().Format("2006-01-02 15:04:05")
			html := `<div hx-swap-oob="innerHTML:#update-timestamp">` + timeStamp + `</div`
			s.broadcast([]byte(html))

			time.Sleep(3 * time.Second)
		}
	}(srv)

	err := http.ListenAndServe("localhost:8080", &srv.mux)
	if err != nil {
		log.Fatalf("%v\n", err)
		return
	}
}
