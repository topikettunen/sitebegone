// Copyright 2021 Topi Kettunen.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
)

const (
	sectionBegin = "# Added by sitebegone"
	sectionEnd   = "# End of sitebegone section"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: sitebegone URL\n")
		os.Exit(1)
	}
	path := hostsPath()
	blockedHosts := newBlockedHosts(path)
	blockedHosts.add(os.Args[1])
	blockedHosts.write()
}

// Return path to hosts file.
func hostsPath() string {
	// TODO: add OS specific paths
	return "/etc/hosts"
}

// Structure for representing blocked sites.
type blockedHosts struct {
	path string

	// TODO: Defaults to 127.0.0.1 for now
	addr net.IP
	// TODO: Currently these are only used for creating a new section and
	// getting the last two (new) hosts but most likely there are some
	// other uses for this too
	hosts []string
}

// Create new structure for blocked hosts.
func newBlockedHosts(path string) *blockedHosts {
	return &blockedHosts{
		path:  path,
		addr:  net.IPv4(127, 0, 0, 1),
		hosts: getHosts(path),
	}
}

// Get currently blocked hosts from the hosts file.
func getHosts(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	sectionStartFound := false
	hosts := []string{}
	for scanner.Scan() {
		if scanner.Text() == sectionBegin {
			sectionStartFound = true
		} else if scanner.Text() == sectionEnd {
			break
		}
		if sectionStartFound {
			if strings.Contains(scanner.Text(), "127.0.0.1") {
				split := strings.Split(scanner.Text(), "\t")
				url, err := url.Parse(split[1])
				if err != nil {
					log.Fatal(err)
				}
				hosts = append(hosts, url.String())
			}
		}
	}
	return hosts
}

// Find section from hosts file and return the current line and the
// cursor pointing to it. This finds the section from bottom up since
// it's more likely that the section is found faster that way.
func findSection(file *os.File, start int64) (string, int64) {
	line := ""
	cursor := start
	stat, _ := file.Stat()
	filesize := stat.Size()
	for {
		cursor--
		if _, err := file.Seek(cursor, io.SeekEnd); err != nil {
			log.Fatal(err)
		}
		char := make([]byte, 1)
		if _, err := file.Read(char); err != nil {
			log.Fatal(err)
		}
		if cursor != -1 && (char[0] == 10 || char[0] == 13) {
			break
		}
		line = fmt.Sprintf("%s%s", string(char), line)
		if cursor == -filesize {
			break
		}
	}
	return line, cursor
}

// Add site to blocked hosts.
func (bh *blockedHosts) add(host string) {
	for _, e := range bh.hosts {
		if e == host {
			fmt.Fprintf(os.Stderr, "%s already blocked\n", host)
			os.Exit(0)
		}
	}
	bh.hosts = append(bh.hosts, host, fmt.Sprintf("www.%s", string(host)))
}

// Write blocked hosts to hosts file.
func (bh *blockedHosts) write() {
	if sectionFound(bh.path) {
		bh.appendToSection()
	} else {
		bh.newSection()
	}
}

// Find sitebegone section from hosts file.
func sectionFound(file string) bool {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	sectionStartFound := false
	sectionEndFound := false
	for scanner.Scan() {
		if scanner.Text() == sectionBegin {
			sectionStartFound = true
		} else if scanner.Text() == sectionEnd {
			sectionEndFound = true
		}
	}
	if sectionStartFound && !sectionEndFound {
		log.Fatal(errors.New("sitebegone: Section found with no end."))
	} else if !sectionStartFound && sectionEndFound {
		log.Fatal(errors.New("sitebegone: Section end found with no start."))
	}
	return sectionStartFound && sectionEndFound
}

// Create a new section for sitebegone.
func (bh *blockedHosts) newSection() {
	f, err := os.OpenFile(bh.path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	section := []byte(fmt.Sprintf("\n%s\n\n", sectionBegin))
	if _, err := f.Write(section); err != nil {
		f.Close()
		log.Fatal(err)
	}
	for _, e := range bh.hosts {
		entry := fmt.Sprintf("%s\t%s\n", bh.addr, e)
		if _, err := f.Write([]byte(entry)); err != nil {
			f.Close()
			log.Fatal(err)
		}
	}
	section = []byte(fmt.Sprintf("\n%s\n\n", sectionEnd))
	if _, err := f.Write(section); err != nil {
		f.Close()
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

// Append a new site to the existing section in hosts file.
func (bh *blockedHosts) appendToSection() {
	f, err := os.OpenFile(bh.path, os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	line := ""
	cursor := int64(0)
	for {
		line, cursor = findSection(f, cursor)
		if strings.Contains(line, sectionEnd) {
			break
		}
	}
	if _, err := f.Seek(cursor-1, io.SeekEnd); err != nil {
		log.Fatal(err)
	}
	remainderOfFile, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Seek(cursor, io.SeekEnd); err != nil {
		log.Fatal(err)
	}
	newHosts := bh.hosts[len(bh.hosts)-2:]
	for i, e := range newHosts {
		entry := fmt.Sprintf("%s\t%s\n", bh.addr, e)
		if i == len(newHosts)-1 {
			entry = fmt.Sprintf("%s\t%s", bh.addr, e)
		}
		if _, err := f.Write([]byte(entry)); err != nil {
			f.Close()
			log.Fatal(err)
		}
	}
	if _, err := f.Write([]byte(remainderOfFile)); err != nil {
		f.Close()
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
