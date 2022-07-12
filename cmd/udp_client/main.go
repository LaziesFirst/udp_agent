package main

import (
	"agent/internal/config_loader"
	"agent/internal/csv_process"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
)

// TODO 여기서 가짜 데이터를 서버로 보내서 csv 파일을 업데이트 하는지 확인
func main() {
	// 환경 설정 읽어오기
	cl, err := config_loader.NewConfigLoader("../../../configs/udp_agent_conf.json")
	if err != nil {
		log.Fatalf("NewConfigLoader Error : %s", err.Error())
	}
	log.Println("Load configs done")
	laddr, err := net.ResolveUDPAddr("udp",
		fmt.Sprintf(":%s", cl.Configs["ports"]["udp_server"]))
	if err != nil {
		log.Fatal(err)
	}
	client, _ := net.DialUDP("udp", nil, laddr)
	defer client.Close()
	testMethods := []string{"POST", "GET", "PUT", "DELETE"}
	remoteIp := "123.123.123.123"
	url := "path_1/path_2/resource"
	for k := 0; k < 1000; k++ {
		simultaneousRequestCount := rand.Intn(300) + 1
		endTime := time.Now().Unix()
		for i := 0; i < simultaneousRequestCount; i++ {
			latency := int64(rand.Intn(21))
			startTime := endTime - latency
			bytes, _ := json.Marshal(csv_process.RawData{
				StartTime: startTime,
				EndTime:   endTime,
				Latency:   latency,
				Method:    testMethods[rand.Intn(4)],
				Url:       url,
				RemoteIp:  remoteIp,
			})
			client.Write(bytes)
			buf := make([]byte, 1024)
			n, addr, _ := client.ReadFrom(buf)
			fmt.Println("client received: ", string(buf[:n]), addr)
		}
		time.Sleep(1 * time.Second)
	}
}
