package csv_process

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

func openAndWrite(writeSig chan string) {
	for sig := range writeSig {
		file, err := os.OpenFile("/home/lazies/GolandProjects"+
			"/udp_agent/internal/csv_process/test.csv",
			os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Println(err.Error())
		}
		wr := csv.NewWriter(bufio.NewWriter(file))
		if err = wr.Write([]string{sig}); err != nil {
			log.Println(err.Error())
		}
		fmt.Println("write done")
		time.Sleep(1 * time.Second)
		wr.Flush()
		if err = file.Close(); err != nil {
			log.Println(err.Error())
		}
	}
}

func openAndRead(readSig, readResult chan []string) {
	for sig := range readSig {
		file, err := os.Open("/home/lazies/GolandProjects" +
			"/udp_agent/internal/csv_process/test.csv")
		if err != nil {
			log.Println(err.Error())
		}
		fmt.Print(sig[0])
		rdr := csv.NewReader(bufio.NewReader(file))
		rows, _ := rdr.ReadAll()
		result := []string{}
		if rows != nil {
			result = append(result, rows[len(rows)-1][0])
		}
		readResult <- result
	}
}

func TestOpenFileAtTheSameTime(t *testing.T) {
	write := make(chan string)
	read := make(chan []string)
	result := make(chan []string)
	go openAndWrite(write)
	go openAndRead(read, result)
	go func() {
		for i := 0; i < 10000; i++ {
			write <- fmt.Sprintf("%d string data", i)
		}
	}()
	for i := 0; i < 1000; i++ {
		read <- []string{""}
		res := <-result
		fmt.Println(res)
		time.Sleep(500 * time.Millisecond)
	}
}

func TestReadFromLastLine(t *testing.T) {
	filename := "/home/lazies/GolandProjects" +
		"/udp_agent/files/v_1_2022-07-11-19:21:32/raw_data.csv"
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	fileInfo, err := os.Stat(filename)
	if err != nil {
		panic(err)
	}
	fileSize := fileInfo.Size()
	for i := int64(1); i < fileSize; i++ {
		buffer := make([]byte, 1)
		offset := fileSize - i
		numRead, _ := file.ReadAt(buffer, offset)
		fmt.Println(numRead, buffer, string(buffer))
		time.Sleep(500 * time.Millisecond)
	}
}
