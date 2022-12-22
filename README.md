# crtsh

Small wrapper around [crt.sh](crt.sh). For domain passed, it extracts all values that are found at crt.sh.

# Installation

```bash
$ git clone https://github.com/lateralusd/crtsh && cd crtsh
$ go build
$ ./crtsh --help
Usage of ./crtsh:
  -a	print unresolvable (default true)
  -d string
    	dns server to use (default "8.8.8.8")
  -l	check whether host is live
  -p int
    	dns port number (default 53)
  -t	print as table
$ ./crtsh -a=false -l=true -t=true google.com # check for live hosts, dont show unresolvable and print output as table
+----------------------------------------------------------------------------------------------------------------------------------------------+
|                                        Live domains for "google.com" (129 resolvable; 56 unresolvable)                                       |
+-----+---------------------------------+------------------------------------------------------------------------------------------------------+
|   # | NAME                            | IP ADDRESSES                                                                                         |
+-----+---------------------------------+------------------------------------------------------------------------------------------------------+
|   1 | google.com.au                   | 142.250.186.99                                                                                       |
|   2 | google.com.br                   | 216.58.212.163                                                                                       |
|   3 | m.google.com                    | 142.250.74.203                                                                                       |
|   4 | google.com.tn                   | 172.217.18.4                                                                                         |
|   5 | google.com.gr                   | 142.250.184.195                                                                                      |
|  12 | google.com.qa                   | 172.217.16.131                                                                                       |
|  13 | google.com.np                   | 142.250.186.163                                                                                      |
|  14 | google.com.hk                   | 142.250.184.195                                                                                      |
|  15 | google.com.af                   | 172.217.18.4                                                                                         |
|  16 | google.com.ly                   | 172.217.18.3                                                                                         |
|  17 | alt1.gmail-smtp-in.l.google.com | 142.250.153.27                                                                                       |
|  18 | adwords.google.com              | 142.250.110.139, 142.250.110.100, 142.250.110.138, 142.250.110.101, 142.250.110.102, 142.250.110.113 |
```
