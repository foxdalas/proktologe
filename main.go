package main

import (
	"github.com/anvie/port-scanner"
	"github.com/labstack/echo"
	"github.com/patrickmn/go-cache"
	"net/http"
	"sync"
	"time"
)

type Echo struct {
	*echo.Echo
}

type Hosts []*Ports

type Ports struct {
	Address string `json:"address"`
	Open    []int  `json:"open"`
}

type PortScan struct {
	cache *cache.Cache
	mx    *sync.Mutex
}

func (p *PortScan) scan(host string) *Ports {
	data, found := p.cache.Get(host)
	if found {
		return data.(*Ports)
	}
	p.mx.Lock()
	defer p.mx.Unlock()
	ps := portscanner.NewPortScanner(host, 200*time.Millisecond, 2000)
	openPorts := &Ports{
		Address: host,
		Open:    ps.GetOpenedPort(1, 65535),
	}
	p.cache.Set(host, openPorts, 30*time.Minute)
	return openPorts
}

func (p *PortScan) historyData() Hosts {
	var data Hosts

	for _, value := range p.cache.Items() {
		data = append(data, value.Object.(*Ports))
	}

	return data
}

func (p *PortScan) getOpenPorts(c echo.Context) error {
	return c.JSON(http.StatusOK, p.scan(c.Param("host")))
}

func (p *PortScan) getScanList(c echo.Context) error {
	return c.JSON(http.StatusOK, p.historyData())
}

func main() {
	p := &PortScan{
		cache: cache.New(20*time.Minute, 25*time.Minute),
		mx:    new(sync.Mutex),
	}

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.GET("/scan/:host", p.getOpenPorts)
	e.GET("/scan/list", p.getScanList)

	e.Logger.Fatal(e.Start(":8080"))
}
