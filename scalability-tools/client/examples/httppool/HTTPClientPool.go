package httppool

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
)

var Pool = NewHTTPClientPool()

var Client = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

type HTTPClientPool struct {
	pool     chan *http.Client
	maxSize  int
	currSize int
	*sync.Mutex
	once sync.Once
}

func NewHTTPClientPool() *HTTPClientPool {
	return &HTTPClientPool{
		currSize: 0,
		Mutex:    &sync.Mutex{},
	}
}

func (p *HTTPClientPool) Init(maxSize int) {
	p.once.Do(func() {
		fmt.Println("Use HTTP Pool. Pool size:", maxSize)
		p.maxSize = maxSize
		p.pool = make(chan *http.Client, maxSize)
		for i := 0; i < maxSize; i++ {
			p.pool <- &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}
		}
	})
}

func (p *HTTPClientPool) Get() *http.Client {
	return <-p.pool

}

func (p *HTTPClientPool) Put(client *http.Client) {
	p.pool <- client
}
