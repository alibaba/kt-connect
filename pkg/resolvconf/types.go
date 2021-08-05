package resolvconf

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
	"strings"
)

const (
	KTADDED   = " # ktctl added nameserver"
	KTCOMMENT = " # ktctl comment"
)

type ConfInterface interface {
	AddNameserver(nameserver string) error
	RestoreConfig() error
}

type Conf struct{}

func (c *Conf) AddNameserver(nameserver string) error {
	fileName := c.getResolvConf()
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	var buf bytes.Buffer

	prefix := fmt.Sprintf("nameserver %s", nameserver)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "nameserver") {
			buf.WriteString("#")
			buf.WriteString(line)
			buf.WriteString(KTCOMMENT)
			buf.WriteString("\n")
		} else if strings.HasPrefix(line, prefix) {
			return nil
		} else {
			buf.WriteString(line)
			buf.WriteString("\n")
		}
	}

	// Add nameserver and comment to resolv.conf
	buf.WriteString(fmt.Sprintf("%s%s", prefix, KTADDED))
	buf.WriteString("\n")

	stat, _ := f.Stat()
	err = ioutil.WriteFile(fileName, buf.Bytes(), stat.Mode())
	return err
}

// RestoreConfig remove the nameserver which is added by ktctl.
func (c *Conf) RestoreConfig() error {
	fileName := c.getResolvConf()
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	var buf bytes.Buffer

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasSuffix(line, KTCOMMENT) {
			line = strings.TrimSuffix(line, KTCOMMENT)
			line = strings.TrimPrefix(line, "#")
			buf.WriteString(line)
			buf.WriteString("\n")
		} else if strings.HasSuffix(line, KTADDED) {
			log.Info().Msgf("remove line: %s ", line)
		} else {
			buf.WriteString(line)
			buf.WriteString("\n")
		}
	}

	stat, _ := f.Stat()
	err = ioutil.WriteFile(fileName, buf.Bytes(), stat.Mode())
	return err
}

func (c *Conf) getResolvConf() string {
	return "/etc/resolv.conf"
}
