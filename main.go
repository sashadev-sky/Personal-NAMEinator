package main

import (
	"flag"
	"fmt"
	"github.com/miekg/dns"
	"log"
	"math/rand"
	"os/exec"
	"regexp"
	"runtime"
	"time"
)

// PROGRAM
var VERSION = "0.2 alpha"

// GLOBALS
var nsStore = nsInfoMap{ns: make(map[string]NSinfo)}
var dStore = dInfoMap{d: make(map[string]Dinfo)}
var appConfiguration APPconfig

type APPconfig struct {
	numberOfDomains int
	debug           bool
	contest         bool
	nameserver      string
}

// process flags
func processFlags() {
	var appConfigstruct APPconfig
	flagNumberOfDomains := flag.Int("domains", 100, "number of domains to be tested")
	flagNameserver := flag.String("nameserver", "", "specify a nameserver instead of using defaults")
	flagContest := flag.Bool("contest", true, "enable or disable a contest against your locally configured DNS server")
	flagDebug := flag.Bool("debug", false, "enable or disable debugging")
	flag.Parse()
	appConfigstruct.numberOfDomains = *flagNumberOfDomains
	appConfigstruct.debug = *flagDebug
	appConfigstruct.contest = *flagContest
	appConfigstruct.nameserver = *flagNameserver
	appConfiguration = appConfigstruct
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

// prints welcome messages
func printWelcome() {
	fmt.Println("starting NAMEinator - version " + VERSION)
	fmt.Printf("understood the following configuration: %+v\n", appConfiguration)
	fmt.Println("-------------")
	fmt.Println("NOTE: as this is an alpha - we rely on feedback - please report bugs and featurerequests to https://github.com/mwiora/NAMEinator/issues and provide this output")
	fmt.Println("OS: " + runtime.GOOS + " ARCH: " + runtime.GOARCH)
	fmt.Println("-------------")
}

func processResults() {
	nsStore.mutex.Lock()
	defer nsStore.mutex.Unlock()
	for _, entry := range nsStore.ns {
		nsResults := nsStoreGetMeasurement(entry.IPAddr)
		entry.rttAvg = nsResults.rttAvg
		entry.rttMin = nsResults.rttMin
		entry.rttMax = nsResults.rttMax
		nsStore.ns[entry.IPAddr] = entry
	}
}

// prints results
func printResults() {
	fmt.Println("")
	fmt.Println("finished - presenting results: ") // TODO: Colorful representation in a table PLEASE
	for _, nameserver := range nsStore.ns {
		fmt.Println("")
		fmt.Println(nameserver.IPAddr + ": ")
		fmt.Printf("Avg. [%v], Min. [%v], Max. [%v]", nameserver.rttAvg, nameserver.rttMin, nameserver.rttMax)
		if appConfiguration.debug {
			fmt.Println(nsStoreGetRecord(nameserver.IPAddr))
		}
		fmt.Println("")
	}
}

// prints bye messages
func printBye() {
	fmt.Println("")
	fmt.Println("Au revoir!")
}

func prepareBenchmark() {
	var domains []string

	if appConfiguration.contest {
		// we need to know who we are testing
		var localdns = getOSdns()
		loadNameserver(localdns, "localhost")
	}

	if appConfiguration.nameserver == "" {
		// read global nameservers from given file
		fmt.Println("trying to load nameservers from datasrc/nameserver-globals.csv")
		readNameserversFromFile("datasrc/nameserver-globals.csv")
	} else {
		loadNameserver(appConfiguration.nameserver, "givenByParameter")
	}

	// read domains from given file
	fmt.Println("trying to load domains from datasrc/alexa-top-2000-domains.txt")
	alldomains, err := readDomainsFromFile("datasrc/alexa-top-2000-domains.txt")
	_ = err // TODO: Exception handling in case that the files do not exist
	// randomize domains from file to avoid cached results
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(alldomains), func(i, j int) { alldomains[i], alldomains[j] = alldomains[j], alldomains[i] })
	// take care only for the domain-tests we were looking for
	domains = alldomains[0:appConfiguration.numberOfDomains]
	dStoreAddFQDN(domains)

}

func performBenchmark() {
	// initialize DNS client
	c := new(dns.Client)
	// to avoid overload against one server we will test all defined nameservers with one domain before proceeding
	for _, domain := range dStore.d {

		m1 := new(dns.Msg)
		m1.Id = dns.Id()
		m1.RecursionDesired = true
		m1.Question = make([]dns.Question, 1)
		m1.Question[0] = dns.Question{domain.FQDN, dns.TypeA, dns.ClassINET}

		// iterate through all given nameservers
		for _, nameserver := range nsStore.ns {
			in, rtt, err := c.Exchange(m1, "["+nameserver.IPAddr+"]"+":53")
			_ = in
			nsStoreSetRTT(nameserver.IPAddr, rtt)
			_ = err // TODO: Take care about errors during queries against the DNS - we will accept X fails
		}
		fmt.Print(".")
	}
}

func main() {
	// process startup parameters and welcome
	processFlags()
	printWelcome()

	prepareBenchmark()

	// lets go benchmark - iterate through all domains
	fmt.Println("LETS GO - each dot is a completed domain request against all nameservers")
	performBenchmark()

	processResults()
	printResults()
	printBye()
}
