package router

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

const pathKtConf = "/etc/kt.conf"

func ReadKtConf() (*KtConf, error) {
	ktConfFile, err := ioutil.ReadFile(pathKtConf)
	if err != nil {
		return nil, fmt.Errorf("failed to read kt configuration file: %s", err)
	}
	var ktConf KtConf
	err = json.Unmarshal(ktConfFile, &ktConf)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kt configuration file: %s", err)
	}
	return &ktConf, nil
}

func WriteKtConf(ktConf *KtConf) error {
	bytes, err := json.Marshal(ktConf)
	if err != nil {
		return fmt.Errorf("failed to parse setup parameters: %s", err)
	}
	err = ioutil.WriteFile(pathKtConf, bytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to create kt configuration: %s", err)
	}
	return nil
}
