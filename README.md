# proxy-finder
Proxy-finder is a http proxy server discovering tool. 
1. Read an ip rage from file
2. Select some ips from the list randomly, and send tcp syn packets to them.
3. When receive the packets from the hosts above, try to judge whether they are proxies. 

## Requires
```
apt-get install libpcap-dev
apt-get install golang
git clone https://github.com/robertdavidgraham/masscan
cd masscan
make
```

## Build
```
go build proxy-finder.go
```

## Usage
proxy-finder speed ports
The result will record to pf.log.
The proxy-finder read destination ips from zone_include.zone, and exclude the ips in zone_exclude.zone.

## Usage flag:
- speed : transmit rate(packets/second)
- ports  : port to scan(like 3128,8080)

