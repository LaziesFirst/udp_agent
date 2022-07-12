package csv_process

import (
	"agent/internal/config_loader"
	"agent/internal/utils"
	"fmt"
	"path/filepath"
	"time"
)

type RawData struct {
	EndTime    int64                  `json:"end_time"`
	StartTime  int64                  `json:"start_time"`
	Latency    int64                  `json:"latency"`
	Method     string                 `json:"method"`
	Url        string                 `json:"url"`
	RemoteIp   string                 `json:"remote_ip"`
	Header     string                 `json:"header,omitempty"`     // additional
	Parameters map[string]interface{} `json:"parameters,omitempty"` // additional
}

type TpsData struct {
	EndTime          int64   // 종료시간
	TransactionCount float64 // 시간 내 트랜젝션 개수
	AverageLatency   float64 // 레이턴시
}

type WriterInfo struct {
	FileName     string       // 파일명
	WriteRequest chan RawData // csv 에 쓸 데이터
}

type RawReaderInfo struct {
	RawDatas     []RawData
	RawDataBytes []byte
}

type TpsReaderInfo struct {
	TpsDatas     []TpsData
	TpsDataBytes []byte
}

type CsvProcess struct {
	RawWriterInfo    WriterInfo
	TpsWriterInfo    WriterInfo
	RawReaderInfo    RawReaderInfo
	TpsReaderInfo    TpsReaderInfo
	tpsCalculateData TpsData
}

// set init configs
func (cp *CsvProcess) Init(configs config_loader.ConfigMap) (err error) {
	path := configs["csv_data"]["path"].(string)
	version := configs["csv_data"]["version"].(string)
	execTime := time.Now()
	path = filepath.Join(path,
		fmt.Sprintf("%s_%s", version, execTime.Format("2006-01-02-15:04:05")))
	cp.RawWriterInfo = WriterInfo{
		FileName:     filepath.Join(path, configs["csv_data"]["raw_filename"].(string)),
		WriteRequest: make(chan RawData),
	}
	cp.RawReaderInfo = RawReaderInfo{
		RawDatas:     []RawData{},
		RawDataBytes: []byte{},
	}
	cp.TpsWriterInfo = WriterInfo{
		FileName:     filepath.Join(path, configs["csv_data"]["tps_filename"].(string)),
		WriteRequest: make(chan RawData),
	}
	cp.TpsReaderInfo = TpsReaderInfo{
		TpsDatas:     []TpsData{},
		TpsDataBytes: []byte{},
	}
	cp.tpsCalculateData = TpsData{}

	// csv 파일 디렉토리 없으면 생성
	if err = utils.MakeDirIfNotExist(path); err != nil {
		return err
	}
	// raw 데이터 csv 없으면 생성
	if err = utils.MakeFileIfNotExists(cp.RawWriterInfo.FileName,
		configs["file_headers"]["raw"]); err != nil {
		return err
	}
	// tps 데이터 csv 없으면 생성
	if err = utils.MakeFileIfNotExists(cp.TpsWriterInfo.FileName,
		configs["file_headers"]["tps"]); err != nil {
		return err
	}
	// start go routines
	go cp.RawDataWriter()
	go cp.RawDataReader()
	go cp.TpsDataWriter()
	go cp.TpsDataReader()
	return nil
}
