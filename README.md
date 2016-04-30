# convert gfwlist to dnsmasq

## install

### Download binnary

Download binnary from [Release](https://github.com/yinheli/gfwlist2dnsmasq/releases)

## usage

```bash

The gfwlist to dnsmasq converter by yinheli @version 1.0.1 

Usage of ./gfwlist2dnsmasq
  -c	try clear old ipset list if exists. (default true)
  -d	download latest gfwlist from: https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt . if false will load gfwlist file from current dir.
  -h string
    	upstream dns host (default "127.0.0.1")
  -i	convent with ipset, dnsmasq will add the dns result ip to ipset automaticly. (default true)
  -n string
    	the ipset list name which you want. (default "gfwlist")
  -o string
    	output the convent result to file location. (default "/etc/dnsmasq.d/gfwlist.conf")
  -p int
    	upstream dns host port (default 5353)
  -r	after convert try to restart dnsmasq service. (default true)
```
