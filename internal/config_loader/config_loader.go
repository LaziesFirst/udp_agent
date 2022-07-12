package config_loader

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
)

// 환경 설정 파일이 저장될 형태
type ConfigMap map[string]map[string]interface{}

// 환경 설정
type ConfigLoader struct {
	filePath string
	Configs  ConfigMap
}

// 환경 설정 불러오기
// "~/udp_agent/configs" 에 환경 설정 파일이 있어야 함
func NewConfigLoader(filename string) (*ConfigLoader, error) {
	cl := new(ConfigLoader)
	if filename == "" {
		return nil, errors.New("empty config filename")
	}
	cl.filePath =
		filepath.Join("configs", filename)
	if err := cl.loadConfigs(); err != nil {
		return nil, err
	}
	return cl, nil
}

// 이중 맵 형태로 환경파일 덮어씌우기
func (cl *ConfigLoader) loadConfigs() (err error) {
	file, err := ioutil.ReadFile(cl.filePath)
	if err != nil {
		return err
	}
	cl.Configs = make(map[string]map[string]interface{})
	if err = json.Unmarshal(file, &cl.Configs); err != nil {
		return err
	}
	return nil
}
