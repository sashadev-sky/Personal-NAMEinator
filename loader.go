package main

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
)

// load nameservers
func loadNameserver(ip string, name string) {
	nsStoreAddNS(ip, name, "LOCAL")
}

// load nameservers
func readNameserversFromFile(filename string) {
	csvFile, _ := os.Open(filename)
	nameserverReader := csv.NewReader(bufio.NewReader(csvFile))
	for {
		line, err := nameserverReader.Read()
		if err == io.EOF {
			break
		}
		// fmt.Println(line)
		nsStoreAddNS(line[0], line[1], line[2])
		_ = err
	}
}

// readDomainsFromFile reads a whole file into memory
// and returns a slice of its lines.
func readDomainsFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
