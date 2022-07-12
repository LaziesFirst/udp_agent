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
func (cp *CsvProcess) writeRawData(rawData RawData) error {
	file, err := os.OpenFile(cp.RawWriterInfo.FileName,
		os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	wr := csv.NewWriter(bufio.NewWriter(file))
	params := ""
	if rawData.Parameters != nil {
		bytes, _ := json.Marshal(rawData.Parameters)
		params = string(bytes)
	}
	if err = wr.Write([]string{
		fmt.Sprintf("%d", rawData.EndTime),
		fmt.Sprintf("%d", rawData.StartTime),
		fmt.Sprintf("%d", rawData.Latency),
		rawData.Method,
		rawData.Url,
		rawData.RemoteIp,
		rawData.Header,
		params,
	}); err != nil {
		return err
	}
	wr.Flush()
	if err = file.Close(); err != nil {
		return err
	}
	return nil
}

// raw 데이터를 csv 파일에 작성하는 과정을 담당하는 고루틴
func (cp *CsvProcess) RawDataWriter() {
	defer func() {
		log.Println("Stop RawDataWriter")
	}()
	for rawData := range cp.RawWriterInfo.WriteRequest {
		if err := cp.writeRawData(rawData); err != nil {
			log.Printf("RawDataWriter ERROR : %s", err.Error())
			continue
		}
	}
}

// csv 컬럼 구조체로 변환
func (cp *CsvProcess) rawColumnsToStruct(columns []string) error {
	etInt, err := strconv.Atoi(columns[0])
	if err != nil {
		return err
	}
	stInt, err := strconv.Atoi(columns[1])
	if err != nil {
		return err
	}
	ltInt, err := strconv.Atoi(columns[2])
	if err != nil {
		return err
	}
	rd := RawData{
		EndTime:   int64(etInt),
		StartTime: int64(stInt),
		Latency:   int64(ltInt),
		Method:    columns[3],
		Url:       columns[4],
		RemoteIp:  columns[5],
		Header:    columns[6],
	}
	if columns[7] != "" { // params 가 존재하면
		params := make(map[string]interface{})
		if err = json.Unmarshal([]byte(columns[7]), &params); err != nil {
			return err
		}
		rd.Parameters = params
	}
	cp.RawReaderInfo.RawDatas = append(cp.RawReaderInfo.RawDatas, rd)
	return nil
}

// 파일 읽어서 멤버변수에 가지고 있기
// TODO 특정 범위만 가져와야 한다면 시간을 이용할 것
func (cp *CsvProcess) readRawData(now time.Time) (err error) {
	file, err := ioutil.ReadFile(cp.RawWriterInfo.FileName)
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
		if err = cp.rawColumnsToStruct(line); err != nil {
			return err
		}
	}
	return nil
}

// 매 틱마다 파일 읽어서 가지고 있기
func (cp *CsvProcess) RawDataReader() {
	defer func() {
		log.Println("Stop RawDataReader")
	}()
	var err error
	ticker := time.NewTicker(1 * time.Second) // 1초마다 파일 읽기
	for t := range ticker.C {
		cp.RawReaderInfo.RawDatas = []RawData{}
		if err = cp.readRawData(t); err != nil {
			log.Printf("readRawData ERROR : %s", err.Error())
			continue
		}
		if cp.RawReaderInfo.RawDataBytes, err =
			json.Marshal(cp.RawReaderInfo.RawDatas); err != nil {
			log.Printf("RawDataMarshal ERROR : %s", err.Error())
			continue
		}
	}
}
