package dns

import (
	"bufio"
	"context"
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/gofrs/flock"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const ktHostsEscapeBegin = "# Kt Hosts Begin"
const ktHostsEscapeEnd = "# Kt Hosts End"

// TODO: this is a temporary solution to avoid dumping after cleanup triggered
var doNotDump = false

// DropHosts remove hosts domain record added by kt
func DropHosts() {
	doNotDump = true
	lines, err := loadHostsFile()
	if err != nil {
		log.Error().Err(err).Msgf("Failed to load hosts file")
		return
	}
	linesAfterDrop, _, err := dropHosts(lines, "")
	if err != nil {
		log.Error().Err(err).Msgf("Failed to parse hosts file")
		return
	}
	if len(linesAfterDrop) < len(lines) {
		err = updateHostsFile(linesAfterDrop)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to drop hosts file")
			return
		}
		log.Info().Msgf("Drop hosts successful")
	}
}

// DumpHosts dump service domain to hosts file
func DumpHosts(hostsMap map[string]string, namespaceToDrop string) error {
	if doNotDump {
		return nil
	}
	lines, err := loadHostsFile()
	if err != nil {
		log.Error().Err(err).Msgf("Failed to load hosts file")
		return err
	}
	linesBeforeDump, linesToKeep, err := dropHosts(lines, namespaceToDrop)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to parse hosts file")
		return err
	}
	if err = updateHostsFile(mergeLines(linesBeforeDump, dumpHosts(hostsMap, linesToKeep))); err != nil {
		log.Warn().Msgf("Failed to dump hosts file")
		log.Debug().Msg(err.Error())
		return err
	}
	log.Debug().Msg("Dump hosts successful")
	return nil
}

func dropHosts(rawLines []string, namespaceToDrop string) ([]string, []string, error) {
	escapeBegin := -1
	escapeEnd := -1
	midDomain := fmt.Sprintf(".%s", namespaceToDrop)
	fullDomain := fmt.Sprintf(".%s.svc.%s", namespaceToDrop, opt.Get().Connect.ClusterDomain)
	keepShortDomain := namespaceToDrop != opt.Get().Global.Namespace
	recordsToKeep := make([]string, 0)
	for i, l := range rawLines {
		if l == ktHostsEscapeBegin {
			escapeBegin = i
		} else if l == ktHostsEscapeEnd {
			escapeEnd = i
		} else if escapeBegin >= 0 && escapeEnd < 0 && namespaceToDrop != "" {
			if ok, err := regexp.MatchString(".+ [^.]+$", l); ok && err == nil {
				if keepShortDomain {
					recordsToKeep = append(recordsToKeep, l)
				}
			} else if !strings.HasSuffix(l, midDomain) && !strings.HasSuffix(l, fullDomain) {
				recordsToKeep = append(recordsToKeep, l)
			}
		}
	}
	if escapeEnd < escapeBegin {
		return nil, nil, fmt.Errorf("invalid hosts file: recordBegin=%d, recordEnd=%d", escapeBegin, escapeEnd)
	}

	if escapeBegin >= 0 && escapeEnd > 0 {
		linesAfterDrop := make([]string, len(rawLines)-(escapeEnd-escapeBegin+1))
		if escapeBegin > 0 {
			copy(linesAfterDrop[0:escapeBegin], rawLines[0:escapeBegin])
		}
		if escapeEnd < len(rawLines)-1 {
			copy(linesAfterDrop[escapeBegin:], rawLines[escapeEnd+1:])
		}
		return linesAfterDrop, recordsToKeep, nil
	} else if escapeBegin >= 0 || escapeEnd > 0 {
		return nil, nil, fmt.Errorf("invalid hosts file: recordBegin=%d, recordEnd=%d", escapeBegin, escapeEnd)
	} else {
		return rawLines, []string{}, nil
	}
}

func dumpHosts(hostsMap map[string]string, linesToKeep []string) []string {
	var lines []string
	lines = append(lines, ktHostsEscapeBegin)
	for host, ip := range hostsMap {
		if ip != "" {
			lines = append(lines, fmt.Sprintf("%s %s", ip, host))
		}
	}
	for _, l := range linesToKeep {
		lines = append(lines, l)
	}
	lines = append(lines, ktHostsEscapeEnd)
	return lines
}

func mergeLines(linesBefore []string, linesAfter []string) []string {
	lines := make([]string, len(linesBefore)+len(linesAfter)+2)
	posBegin := len(linesBefore)
	if posBegin > 0 {
		copy(lines[0:posBegin], linesBefore[:])
	}
	if len(linesAfter) > 0 {
		copy(lines[posBegin+1:len(lines)-1], linesAfter[:])
	}
	return lines
}

func loadHostsFile() ([]string, error) {
	var lines []string
	file, err := os.Open(getHostsPath())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func updateHostsFile(lines []string) error {
	lock := flock.New(fmt.Sprintf("%s/hosts.lock", util.KtLockDir))
	timeoutContext, cancel := context.WithTimeout(context.TODO(), 2 * time.Second)
	defer cancel()
	if ok, err := lock.TryLockContext(timeoutContext, 100 * time.Millisecond); !ok {
		return fmt.Errorf("failed to require hosts lock")
	} else if err != nil {
		log.Error().Err(err).Msgf("require hosts file failed with error")
		return err
	}
	defer lock.Unlock()

	file, err := os.Create(getHostsPath())
	if err != nil {
		return err
	}

	// fix sometimes update windows system hosts file but not effective immediately on existed in linesBeforeDump
	// because hosts not closed immediately
	// for example 10.110.55.200 vip-apiserver.test.com, this domain not existed in dns
	defer file.Close()

	w := bufio.NewWriter(file)
	continualEmptyLine := false
	for _, l := range lines {
		if continualEmptyLine && l == "" {
			continue
		}
		continualEmptyLine = l == ""
		fmt.Fprintf(w, "%s%s", l, util.Eol)
	}

	err = w.Flush()
	if err != nil {
		return err
	}
	return nil
}

func getHostsPath() string {
	if os.Getenv("HOSTS_PATH") == "" {
		return os.ExpandEnv(filepath.FromSlash(util.HostsFilePath))
	} else {
		return os.Getenv("HOSTS_PATH")
	}
}
