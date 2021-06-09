# solarmax-metrics

This small piece of code contacts a SolarMAX inverter and provides its metrics in JSON.

## Usage

Simple metric query only using the inverters ip and port:
(Default inverter id 1 will be used)

```bash
pi@raspberrypi:~ $ solarmax-metrics -host=192.168.188.19 -port=26126
{"IDC":0.34,"IL1":0.38,"KDY":15.8,"KMT":129,"KT0":45685,"KYR":1721,"PAC":83,"PRL":1,"SYS":"008-Mains operation","TKK":32,"TNF":50.01,"UDC":3228,"UL1":229.8}
```

Combined metric query with jq only using the inverters ip and port:
(Default inverter id 1 will be used)

```bash
pi@raspberrypi:~ $ solarmax-metrics -host=192.168.188.19 -port=26126 | jq
{
  "IDC": 0.29,
  "IL1": 0.31,
  "KDY": 15.8,
  "KMT": 129,
  "KT0": 45685,
  "KYR": 1721,
  "PAC": 66,
  "PRL": 1,
  "SYS": "008-Mains operation",
  "TKK": 32,
  "TNF": 50.03,
  "UDC": 3255,
  "UL1": 229.9
}
```

Usage (help option):

```bash
pi@raspberrypi:~ $ solarmax-metrics -h
Usage of ./solarmax-metrics:
  -host string
        host/inverter ip address (default "127.0.0.1")
  -inverter int
        inverter id (default 1)
  -loglevel string
        info, warn, debug, trace (default "info")
  -metrics string
        list of metric codes (comma separated) (default "KDY,KMT,KYR,KT0,TNF,TKK,PAC,PRL,IL1,IDC,UL1,UDC,SYS")
  -port int
        port number (default 80)
 ```
