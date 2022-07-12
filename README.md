# udp_agent

## 사용

### 1) config 파일 수정
`udp_agent/configs/udp_agent_conf.json` 파일을 수정합니다.
UDP 서버는 `ports.udp_server` 에 위치한 포트에서 실행됩니다.
UDP 클라이언트는 `ports.web_app` 에 위치한 포트에서 실행됩니다. 
현재는 환경변수로 처리되어 서버 시작시 로드되지만 추후 실행옵션으로 변경될 예정입니다.(TODO)
`csv_data` 에 위치한 `path`에 csv 파일들이 저장됩니다. 
UDP 서버로 전달되는 데이터는 `csv_data.raw_filename` 으로, 
해당 데이터를 TPS 데이터로 변경한 이후에는 `csv_data.tps_filename` 로 관리됩니다.
### 2) udp 서버 실행
```
go run udp_agent/cmd/udp_server/main.go
```
UDP 서버는 실행되어 데이터를 지속적으로 받습니다.
UDP 서버는 csv 파일을 처리할 4개의 고루틴과 데이터 조회를 위한 웹 핸들러를 갖습니다.
csv 파일을 처리하는 4개의 고루틴은 각각 다음과 같습니다.

1. `RawDataWriter()`
   1. UDP 서버가 수신한 데이터를 raw 데이터로 저장합니다.
2. `RawDataReader()`
   1. 1초 틱마다 raw 데이터가 저장된 csv 파일을 읽어서 가지고 있습니다.
   2. 웹 핸들러를 통해 조회 요청이 오면 해당 데이터를 바로 리턴합니다.
3. `TpsDataWriter()`
   1. UDP 서버가 수신한 데이터를 계산해서 tps 데이터로 변환한 이후 저장합니다.
4. `TpsDataReader()`
   1. 5초 틱마다 tps 데이터가 저장된 csv 파일을 읽어서 가지고 있습니다.
   2. 웹 핸들러를 통해 조회 요청이 오면 해당 데이터를 바로 리턴합니다.

### 3) udp 클라이언트 실행
```
go run udp_agent/cmd/udp_server/main.go
```
(TODO) 클라이언트는 현재 서버로 가짜 데이터를 보내고 있습니다. 
