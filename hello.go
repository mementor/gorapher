package main

import (
	"flag"
	"fmt"
	"github.com/mementor/gorapher/gobaser"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var listen string
var tsdataMap map[string]TSData

type Metric struct {
	Timestamp int64
	Type      int64
	Name      string
	Value     int64
}

type TSData struct {
	Name   string
	Values map[int64]int64
}

func init() {
	flag.StringVar(&listen, "listen", "0.0.0.0:6996", "Port on listening for incoming connections")
	flag.Parse()

	tsdataMap = make(map[string]TSData)
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
}

func handleConnection(conn net.Conn, outChan chan Metric) {
	data := make([]byte, 1024)
	// var metric Metric
	for {
		lenn, err := conn.Read(data)
		if err != nil {
			log.Printf("Error on client read, close connection (%v)", err)
			return
		}
		log.Printf("Got metric (len: %v)\n", lenn)
		metric, err := metricFromBytes(data, lenn)
		if err != nil {
			log.Printf("Error parsing metric: %v\n", err)
		} else {
			outChan <- metric
		}
	}
}

func metricFromBytes(data []byte, lenn int) (Metric, error) {
	var metric Metric
	var err error
	str := strings.Trim(string(data[:lenn]), "\n")
	metr := strings.Split(str, " ")
	metric.Timestamp, err = strconv.ParseInt(metr[0], 10, 64)
	if err != nil {
		return Metric{}, err
	}
	metric.Type, err = strconv.ParseInt(metr[1], 10, 64)
	if err != nil {
		return Metric{}, err
	}
	metric.Name = metr[2]
	metric.Value, err = strconv.ParseInt(metr[3], 10, 64)
	if err != nil {
		return Metric{}, err
	}
	return metric, nil
}

func serveRequests() chan Metric {
	metricChan := make(chan Metric)

	go func() {
		l, err := net.Listen("tcp", listen)
		if err != nil {
			fmt.Printf("Cant listen %v\n", listen)
			return
		}

		for {
			c, err := l.Accept()
			if err != nil {
				fmt.Printf("Cant accept: %v", err)
				continue
			}
			defer c.Close()

			go handleConnection(c, metricChan)
		}

	}()

	return metricChan
}

func processMetrics(metrics chan Metric, exitChan chan int) {
	for {
		select {
		case metric := <-metrics:
			fmt.Printf("New metric:\n'%v'\n", metric)
			if _, exists := tsdataMap[metric.Name]; !exists {
				var tsdat TSData
				values := make(map[int64]int64, 1)
				values[metric.Timestamp] = metric.Value
				tsdat.Name = metric.Name
				tsdat.Values = values
				tsdataMap[metric.Name] = tsdat
			} else {
				tsdataMap[metric.Name].Values[metric.Timestamp] += metric.Value
			}
		case <-exitChan:
			fmt.Printf("exiting processMetrics goroutine\n")
			return
		}
		fmt.Printf("Now tsdataMap looks like:\n%v\n", tsdataMap)
	}
}

func main() {
	gobaser.WriteToFile("srvs.http1.hit", time.Now(), 100500)
	exitChan := make(chan int)
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan
		log.Printf("Got signal")
		exitChan <- 1
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	metricsChan := serveRequests()

	isExit := false

	metrics := make(chan Metric)
	goExitChan := make(chan int)
	go processMetrics(metrics, goExitChan)

	for {
		if isExit {
			break
		}
		select {
		case metric := <-metricsChan:
			metrics <- metric
		case <-exitChan:
			isExit = true
			close(goExitChan)
		}
	}

}
