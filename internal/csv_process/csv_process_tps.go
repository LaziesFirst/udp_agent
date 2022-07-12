package csv_process

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

// 실제 파일에 값 작성
func (cp *CsvProcess) writeTpsData() error {
	file, err := os.OpenFile(cp.TpsWriterInfo.FileName,
		os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	wr := csv.NewWriter(bufio.NewWriter(file))
	if err = wr.Write([]string{
		fmt.Sprintf("%d", cp.tpsCalculateData.EndTime),
		fmt.Sprintf("%f", cp.tpsCalculateData.TransactionCount/5.0),
		fmt.Sprintf("%f", cp.tpsCalculateData.AverageLatency/cp.tpsCalculateData.TransactionCount),
	}); err != nil {
		return err
	}
	wr.Flush()
	if err = file.Close(); err != nil {
		return err
	}
	return nil
}

/*
TODO 계산이 맞는지 확인이 필요함(그냥 임의로 계산해놓자)
1. endtime 기준으로 계산하는 것이 맞는가?
2. raw 데이터가 트랜젝션 하나하나 전달되는 것이 맞는가?
3. 5초 간격이라면 첫 시작이 언제인가?(00시00분00초 라든지)
*/
// 파일의 값으로 시계열 데이터 계산하기
func (cp *CsvProcess) calculateTimeSerialData(rawData RawData) error {
	if cp.tpsCalculateData.EndTime != 0 {
		if cp.tpsCalculateData.EndTime+5 > rawData.EndTime { // 누적 시키기
			cp.tpsCalculateData.AverageLatency += float64(rawData.Latency)
			cp.tpsCalculateData.TransactionCount += 1
		} else { // 쓰고 기준 데이터 변경하기
			if err := cp.writeTpsData(); err != nil {
				return err
			}
			cp.tpsCalculateData.EndTime = cp.tpsCalculateData.EndTime + 5
			cp.tpsCalculateData.AverageLatency = float64(rawData.Latency)
			cp.tpsCalculateData.TransactionCount = 1
		}
	} else {
		cp.tpsCalculateData.EndTime = rawData.EndTime
		cp.tpsCalculateData.AverageLatency = float64(rawData.Latency)
		cp.tpsCalculateData.TransactionCount = 1
	}
	return nil
}

// tps 데이터를 csv 파일에 작성하는 과정을 담당하는 고루틴
func (cp *CsvProcess) TpsDataWriter() {
	defer func() {
		log.Println("Stop TpsDataWriter")
	}()
	for rawData := range cp.TpsWriterInfo.WriteRequest {
		if err := cp.calculateTimeSerialData(rawData); err != nil {
			log.Printf("calculateTimeSerialData ERROR : %s", err.Error())
			continue
		}
	}
}

// csv 컬럼 구조체로 변환
func (cp *CsvProcess) tpsColumnsToStruct(columns []string) error {
	etInt, err := strconv.Atoi(columns[0])
	if err != nil {
		return err
	}
	tCnt, err := strconv.ParseFloat(columns[1], 64)
	if err != nil {
		return err
	}
	avgLat, err := strconv.ParseFloat(columns[2], 64)
	if err != nil {
		return err
	}
	td := TpsData{
		EndTime:          int64(etInt),
		TransactionCount: tCnt,
		AverageLatency:   avgLat,
	}
	cp.TpsReaderInfo.TpsDatas = append(cp.TpsReaderInfo.TpsDatas, td)
	return nil
}

// 파일 읽어서 멤버변수에 가지고 있기
// TODO 특정 범위만 가져와야 한다면 시간을 이용할 것
func (cp *CsvProcess) readTpsData(now time.Time) (err error) {
	file, err := ioutil.ReadFile(cp.TpsWriterInfo.FileName)
	if err != nil {
		return err
	}
	reader := csv.NewReader(bytes.NewBuffer(file))
	_, err = reader.Read() // header 건너뛰기
	if err != nil && err != io.EOF {
		return err
	}
	for {
		line, err := reader.Read()
		if err != nil && err != io.EOF {
			return err
		}
		if err == io.EOF {
			break
		}
		if err = cp.tpsColumnsToStruct(line); err != nil {
			return err
		}
	}
	return nil
}

// 매 틱마다 파일 읽어서 가지고 있기
func (cp *CsvProcess) TpsDataReader() {
	defer func() {
		log.Println("Stop TpsDataReader")
	}()
	var err error
	ticker := time.NewTicker(5 * time.Second) // 5초마다 파일 읽기
	for t := range ticker.C {
		cp.TpsReaderInfo.TpsDatas = []TpsData{}
		if err = cp.readTpsData(t); err != nil {
			log.Printf("readTpsData ERROR : %s", err.Error())
			continue
		}
		if cp.TpsReaderInfo.TpsDataBytes, err =
			json.Marshal(cp.TpsReaderInfo.TpsDatas); err != nil {
			log.Printf("TpsDataMarshal ERROR : %s", err.Error())
			continue
		}
	}
}
