package main

import (
	"sync"
	"time"

	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/miekg/dns"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"runtime"
)

var NUMBER_OF_DOMAINS int = 50 // TODO: Implement Switch
var VERSION = "0.1 alpha"
var nsStore = nsInfoMap{ns: make(map[string]NSinfo)}

type NSmeasurement struct {
	rttAvg time.Duration
	rttMin time.Duration
	rttMax time.Duration
}

type NSinfo struct {
	IPAddr           string
	Name             string
	Country          string
	Count            int
	ErrorsConnection int
	ErrorsValidation int
	rtt              []time.Duration
}

type nsInfoMap struct {
	ns    map[string]NSinfo
	mutex sync.RWMutex
}

//// Get IP address entry // DEBUG // TODO: Implement debug switch
//func nsStoreGetRecord(ipAddr string) NSinfo {
//	nsStore.mutex.RLock()
//	defer nsStore.mutex.RUnlock()
//	entry, found := nsStore.ns[ipAddr]
//	if !found {
//		entry.IPAddr = ipAddr
//	}
//	return entry
//}

// Get nameserver average time
func nsStoreGetMeasurement(ipAddr string) NSmeasurement {
	var nsMeasurement = NSmeasurement{}
	nsStore.mutex.RLock()
	defer nsStore.mutex.RUnlock()
	entry, found := nsStore.ns[ipAddr]
	if !found {
		entry.IPAddr = ipAddr
	}
	var total time.Duration = 0
	var min time.Duration = 10000000
	var max time.Duration = 0
	for _, value := range entry.rtt {
		// check for new min record
		if value < min {
			min = value
		}
		// check for new max record
		if value > max {
			max = value
		}
		// add for total time
		total += value
	}
	nsMeasurement.rttAvg = total / time.Duration(NUMBER_OF_DOMAINS)
	nsMeasurement.rttMin = min
	nsMeasurement.rttMax = max
	return nsMeasurement
}

// add rtt to the nameserver slice
func nsStoreSetRTT(ipAddr string, rtt time.Duration) {
	nsStore.mutex.Lock()
	defer nsStore.mutex.Unlock()
	entry, found := nsStore.ns[ipAddr]
	if !found {
		entry.IPAddr = ipAddr
	}
	entry.rtt = append(entry.rtt, rtt)
	entry.Count++
	nsStore.ns[ipAddr] = entry
}

// add rtt to the nameserver slice
func nsStoreAddNS(ipAddr string, name string, country string) {
	nsStore.mutex.Lock()
	defer nsStore.mutex.Unlock()
	entry, found := nsStore.ns[ipAddr]
	if !found {
		entry.IPAddr = ipAddr
	}
	entry.Name = name
	entry.Country = country
	nsStore.ns[ipAddr] = entry
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
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

// return the IP of the DNS used by the operating system
func getOSdns() string {
	// get local dns ip
	out, err := exec.Command("nslookup", ".").Output()
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("%s\n", out)
	re := regexp.MustCompile("\\b\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\b") // TODO: Make IPv6 compatible
	// fmt.Printf("%q\n", re.FindString(string(out)))
	var localdns = re.FindString(string(out))
	return localdns
}

func main() {
	fmt.Println("starting NAMEinator - version " + VERSION + " with configuration:")
	fmt.Printf("- Domains to be requested: %d\n", NUMBER_OF_DOMAINS)
	fmt.Println("-------------")
	fmt.Println("NOTE: as this is an alpha - we rely on feedback - please report bugs and featurerequests to https://github.com/mwiora/NAMEinator/issues and provide this output")
	fmt.Println("OS: " + runtime.GOOS + " ARCH: " + runtime.GOARCH)
	fmt.Println("-------------")

	// we need to know who we are testing
	var localdns = getOSdns()

	// initialize DNS client
	c := new(dns.Client)

	// read domains from given
	fmt.Println("trying to load domains from datasrc/alexa-top-2000-domains.txt")
	alldomains, err := readLines("datasrc/alexa-top-2000-domains.txt")
	_ = err // TODO: Exception handling in case that the files do not exist
	// read global nameservers from given file
	fmt.Println("trying to load nameservers from datasrc/nameserver-globals.csv")
	csvFile, _ := os.Open("datasrc/nameserver-globals.csv")
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

	// randomize domains from file to avoid cached results
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(alldomains), func(i, j int) { alldomains[i], alldomains[j] = alldomains[j], alldomains[i] })

	// take care only for the domain-tests we were looking for
	domains := alldomains[0:NUMBER_OF_DOMAINS]
	// fill list of nameservers to be tested
	nameservers := []string{localdns}
	for _, nameserver := range nsStore.ns {
		// fmt.Println(nameserver.IPAddr)
		nameservers = append(nameservers, nameserver.IPAddr)
	}

	fmt.Println("LETS GO - each dot is a completed domain request against all nameservers")
	// lets go benchmark - iterate through all domains
	// to avoid overload against one server we will test all defined nameservers with one domain before proceeding
	for _, domain := range domains {

		m1 := new(dns.Msg)
		m1.Id = dns.Id()
		m1.RecursionDesired = true
		m1.Question = make([]dns.Question, 1)
		m1.Question[0] = dns.Question{domain, dns.TypeA, dns.ClassINET}

		// iterate through all given nameservers
		for _, nameserver := range nameservers {
			in, rtt, err := c.Exchange(m1, "["+nameserver+"]"+":53")
			_ = in
			nsStoreSetRTT(nameserver, rtt)
			_ = err // TODO: Take care about errors during queries against the DNS - we will accept X fails
		}
		fmt.Print(".")
	}

	fmt.Println("")
	fmt.Println("finished - presenting results: ") // TODO: Colorful representation in a table PLEASE
	for _, nameserver := range nameservers {
		// fmt.Println(nsStoreGetRecord(nameserver)) // DEBUG
		nsStoreEntry := nsStoreGetMeasurement(nameserver)
		fmt.Println("")
		fmt.Println(nameserver + ": ")
		fmt.Printf("Avg. [%v], Min. [%v], Max. [%v]", nsStoreEntry.rttAvg, nsStoreEntry.rttMin, nsStoreEntry.rttMax)
		fmt.Println("")
	}

}
