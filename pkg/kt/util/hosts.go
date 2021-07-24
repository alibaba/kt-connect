package util

import (
	"bufio"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
)

const ktHostsEscapeBegin = "# Kt Hosts Begin"
const ktHostsEscapeEnd = "# Kt Hosts End"

// DropHosts ...
func DropHosts() {
	lines, err := loadHostsFile()
	if err != nil {
		log.Error().Msgf("Failed to load hosts file: %s", err.Error())
		return
	}
	linesAfterDrop, err := dropHosts(lines)
	if err != nil {
		log.Error().Msgf("Failed to parse hosts file: %s", err.Error())
		return
	}
	if len(linesAfterDrop) < len(lines) {
		err = updateHostsFile(linesAfterDrop)
		if err != nil {
			log.Error().Msgf("Failed to update hosts file, you may require %s permission: %s",
				getAdminUserName(), err.Error())
			return
		}
		log.Info().Msgf("Drop hosts successful")
	}
}

// DumpHosts DumpToHosts
func DumpHosts(hostsMap map[string]string) {
	lines, err := loadHostsFile()
	if err != nil {
		log.Error().Msgf("Failed to load hosts file: %s", err.Error())
		return
	}
	linesBeforeDump, err := dropHosts(lines)
	if err != nil {
		log.Error().Msgf("Failed to parse hosts file: %s", err.Error())
		return
	}
	err = updateHostsFile(mergeLines(linesBeforeDump, dumpHosts(hostsMap)))
	if err != nil {
		log.Error().Msgf("Failed to update hosts file, you may require %s permission: %s",
			getAdminUserName(), err.Error())
		return
	}
	log.Info().Msg("Dump hosts successful")
}

func dropHosts(rawLines []string) ([]string, error) {
	escapeBegin := -1
	escapeEnd := -1
	for i, l := range rawLines {
		if l == ktHostsEscapeBegin {
			escapeBegin = i
		} else if l == ktHostsEscapeEnd {
			escapeEnd = i
		}
	}
	if escapeEnd < escapeBegin {
		return nil, fmt.Errorf("invalid hosts file: escapeBegin=%d, escapeEnd=%d", escapeBegin, escapeEnd)
	}

	if escapeBegin >= 0 && escapeEnd > 0 {
		linesAfterDrop := make([]string, len(rawLines)-(escapeEnd-escapeBegin+1))
		if escapeBegin > 0 {
			copy(linesAfterDrop[0:escapeBegin], rawLines[0:escapeBegin])
		}
		if escapeEnd < len(rawLines)-1 {
			copy(linesAfterDrop[escapeBegin:], rawLines[escapeEnd+1:])
		}
		return linesAfterDrop, nil
	} else {
		return rawLines, nil
	}
}

func dumpHosts(hostsMap map[string]string) []string {
	var lines []string
	lines = append(lines, ktHostsEscapeBegin)
	for host, ip := range hostsMap {
		lines = append(lines, fmt.Sprintf("%s %s", ip, host))
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
	file, err := os.Create(getHostsPath())
	if err != nil {
		return err
	}

	w := bufio.NewWriter(file)
	continualEmptyLine := false
	for _, l := range lines {
		if continualEmptyLine && l == "" {
			continue
		}
		continualEmptyLine = l == ""
		fmt.Fprintf(w, "%s%s", l, eol)
	}

	err = w.Flush()
	if err != nil {
		return err
	}
	return nil
}

func getHostsPath() string {
	if os.Getenv("HOSTS_PATH") == "" {
		return os.ExpandEnv(filepath.FromSlash(hostsFilePath))
	} else {
		return os.Getenv("HOSTS_PATH")
	}
}

func getAdminUserName() string {
	if IsWindows() {
		return "administrator"
	}
	return "root"
}
