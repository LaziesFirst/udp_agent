package udp_handler

import (
	"agent/internal/config_loader"
	"agent/internal/csv_process"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
)

type Handler struct {
	Server net.Addr
	Cp     *csv_process.CsvProcess
}

func NewUdpHandler(configs config_loader.ConfigMap) (handler *Handler, err error) {
	handler = new(Handler)
	handler.Cp = new(csv_process.CsvProcess)
	if err = handler.Cp.Init(configs); err != nil { // csv 조작용 고루틴 실행
		return nil, err
	}
	laddr, err := net.ResolveUDPAddr("udp",
		fmt.Sprintf(":%s", configs["ports"]["udp_server"]))
	if err != nil {
		return nil, err
	}
	serve, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, err
	}
	go handler.serveUdp(serve, configs) // UDP 고루틴 실행
	handler.Server = serve.LocalAddr()
	return handler, nil
}

/*
udp_server 데이터 수신 후 csv writer 에게 데이터를 쓰도록 요청
*/
func (h *Handler) serveUdp(serve *net.UDPConn, configs config_loader.ConfigMap) {
	defer func() {
		log.Println("Stop udp_server handler")
	}()
	bSize, _ := strconv.Atoi(configs["udp"]["byte_size"].(string))
	for {
		buf := make([]byte, bSize)
		n, client, err := serve.ReadFromUDP(buf) // 데이터 수신
		if err != nil {
			log.Printf("ReadFrom Error : %s\n", err.Error())
			continue
		}
		if n == 0 {
			continue
		}
		// TODO 변환해서 보내주는 것과 고루틴에서 변환하는 것 속도 확인
		rd := csv_process.RawData{}
		if err = json.Unmarshal(buf[:n], &rd); err != nil {
			log.Printf("Unmarshal Error : %s\n", err.Error())
			continue
		}
		// 요청 처리
		h.Cp.RawWriterInfo.WriteRequest <- rd
		h.Cp.TpsWriterInfo.WriteRequest <- rd
		// TODO temp
		fmt.Println("server received: ", string(buf[:n]), "from", client)
		_, err = serve.WriteTo([]byte("Received"), client)
		if err != nil {
			log.Printf("WriteTo Error : %s\n", err.Error())
			continue
		}
	}
}

func (h *Handler) GetRawData(w http.ResponseWriter, r *http.Request) {
	log.Println("API OK : GetRawData")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).
		Encode(h.Cp.RawReaderInfo.RawDataBytes); err != nil {
		log.Printf("GetRawData Response Fail Error : %s", err.Error())
	}
}

func (h *Handler) GetTpsData(w http.ResponseWriter, r *http.Request) {
	log.Println("API OK : GetTpsData")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).
		Encode(h.Cp.TpsReaderInfo.TpsDataBytes); err != nil {
		log.Printf("GetTpsData Response Fail Error : %s", err.Error())
	}
}
