NAMEinator [![Build Status](https://travis-ci.org/mwiora/NAMEinator.svg)](https://travis-ci.org/mwiora/NAMEinator.svg) [![Code Climate](https://codeclimate.com/github/mwiora/NAMEinator/badges/gpa.svg)](https://codeclimate.com/github/mwiora/NAMEinator)
=========

Are you a power-user with 5 minutes to spare? Do you want a faster internet experience?

Try out NAMEinator. It hunts down the fastest DNS servers available for your
computer to use. NAMEinator runs a fair and thorough benchmark using standardized
datasets in order to provide an individualized recommendation. NAMEinator is completely
free and does not modify your system in any way.
This project began as a 20% project at Google.

NAMEinator runs on Windows (10), Linux (tested on Ubuntu 16.04) and is available with a
a command-line interface and in the near future with a graphical user interface.

how2
---------------

* download and run the corresponding compiled files from releases.

or

* compile yourself (requirement: install go as described here https://golang.org/doc/install)
```
go get github.com/mwiora/NAMEinator
cd ~/go/src/github.com/mwiora/NAMEinator/
go build
./NAMEinator
```

continuation of this project
---------------

This project has been forked from google/namebench. While it seems that the initial Author wanted to move the application towards GO with another application use - I just wanted to continue and maintain the python variant as it did what it should.
After investigation the python code showed up some very frustrating complications, which were leading to my decision to reimplement the initial application idea - also in GO.

sample output of current version
---------------

```
starting NAMEinator - version 0.1 alpha with configuration:
- Domains to be requested: 50
-------------
NOTE: as this is an alpha - we rely on feedback - please report bugs and featurerequests to https://github.com/mwiora/NAMEinator/issues and provide this output
OS: windows ARCH: amd64
-------------
trying to load domains from datasrc/alexa-top-2000-domains.txt
trying to load nameservers from datasrc/nameserver-globals.csv
LETS GO - each dot is a completed domain request against all nameservers
..................................................
finished - presenting results:

172.31.0.2:
Avg. [18.056522ms], Min. [0s], Max. [209.4723ms]

8.8.8.8:
Avg. [10.38975ms], Min. [567.7µs], Max. [112.4137ms]

8.8.4.4:
Avg. [5.984448ms], Min. [191.6µs], Max. [38.1314ms]

208.67.222.222:
Avg. [42.736608ms], Min. [0s], Max. [482.7889ms]

2001:470:20::2:
Avg. [48.21425ms], Min. [209.1µs], Max. [1.5964583s]

156.154.71.1:
Avg. [52.636148ms], Min. [309.7µs], Max. [859.5941ms]

216.146.35.35:
Avg. [54.935102ms], Min. [332.8µs], Max. [279.1109ms]

Process finished with exit code 0
```

checklist
---------------
basics
- [x] perform DNS Requests
- [x] iterate through given nameservers (basic set) and measure time
- [x] randomly select domain names from alexa top 2000 list
- [x] produce cli report
- [x] test on windows and linux
- [x] release cli version

nice2have
= [ ] implement test driven development

to subsitute namebench 1.3.1
- [ ] support localization of execution
- [ ] select the best suitable DNS server
- [ ] provide basic GUI which has the CLI version as its base
- [ ] produce html/pdf report
- [ ] test on windows and linux
- [ ] release gui version

reimplement functions that were planned, but did not work with namebench 1.3.1
- [ ] detect censorship and manipulated dns entries
- [ ] optional upload of results

new features
- [ ] perform identification of best usable dns server not only based on location, but based also on network path traces
- [ ] test caching - disable cache if this option is selected and ask for the domains a second time (increasing number of domains)
