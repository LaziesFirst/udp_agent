package main

import (
	"agent/internal/config_loader"
	"agent/internal/udp_handler"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

/*
udp_server 서버 다음과 같은 3가지 기능을 제공함
1. raw_data 쓰기 -> udp_server 서버로 전달되는 데이터를 csv에 씀
2. raw_data 읽기 -> csv 에 쓰여진 raw_data 를 읽어서 웹 리스폰스 리턴
3. tps_data 읽기 -> csv 에 쓰여진 tps_data 를 읽어서 웹 리스폰스 리턴
(내부적으로) tps_data 쓰기 -> (raw_data -> tps_data) 변경 과정이 고루틴으로 진행됨
*/
func main() {
	// 환경 설정 읽어오기
	fmt.Println(os.Getwd())
	cl, err := config_loader.NewConfigLoader("../../../configs/udp_agent_conf.json")
	if err != nil {
		log.Fatalf("NewConfigLoader Error : %s", err.Error())
	}
	log.Println("Load configs done")
	// TODO 주소 및 포트는 실행 옵션으로 처리 (현재는 환경변수)
	listenPort := cl.Configs["ports"]["udp_server"]
	// UDP 로 동작할 것
	log.Printf("Start UDP Server : %s\n", listenPort)
	handler, err := udp_handler.NewUdpHandler(cl.Configs)
	if err != nil {
		log.Fatalf("NewUdpHandler Error : %s", err.Error())
	}
	// 웹 핸들러
	router := mux.NewRouter()
	router.HandleFunc(cl.Configs["urls"]["get_raw_data"].(string),
		handler.GetRawData).Methods("GET")
	router.HandleFunc(cl.Configs["urls"]["get_tps_data"].(string),
		handler.GetTpsData).Methods("GET")
	log.Printf("Start Web Handler : %s\n", listenPort)
	if err = http.ListenAndServe(fmt.Sprintf(":%s", listenPort), nil); err != nil {
		log.Fatalf("ListenAndServe Error : %s", err.Error())
	}
}
