# solarmax-metrics

This small piece of code contacts a SolarMax S series inverter and provides its metrics in JSON.

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
  -mode string
        query, loop, listmetrics (default "query")
  -port int
        port number (default 80)
 ```

List available metrics:

```bash
tpi@raspberrypi:~ $ solarmax-metrics -mode listmetrics
ADR: Address
BDN: Build number
CAC: Start Ups (?)
DDY: Date day
DIN: Date in integer format with offset 23.12.1510
DMT: Date month
DYR: Date year
EC00: Error Code 0
EC01: Error Code 1
EC02: Error Code 2
EC03: Error Code 3
EC04: Error Code 4
EC05: Error Code 5
EC06: Error Code 6
EC07: Error Code 7
EC08: Error Code 8
FDAT: datetime ?
F_AC: Grid Frequency
ID01: String 1 Current (A)
ID02: String 2 Current (A)
ID03: String 3 Current (A)
IDC: DC Current (A)
IL1: AC Current Phase 1 (A)
IL2: AC Current Phase 2 (A)
IL3: AC Current Phase 3 (A)
KDL: Energy yesterday (Wh)
KDY: Energy today (kWh)
KHR: Operating Hours
KLM: Energy last month (kWh)
KLY: Energy last year (kWh)
KMT: Energy this month (kWh)
KT0: Total Energy(kWh)
KYR: Energy this year (kWh)
LAN: Language
MAC: MAC Address
PAC: AC Power (W)
PDC: DC Power (W)
PIN: Installed Power (W)
PRL: Relative power (%)
SAL: System Alarms
SDAT: datetime ?
SE1: 
SWV: Software Version
SYS: System Status
THR: Time hours
TKK: Inverter Temperature (C)
TMI: Time minutes
TNF: Generated Frequency (Hz)
TNP: Grid period duration
TYP: Type
UD01: String 1 Voltage (V)
UD02: String 2 Voltage (V)
UD03: String 3 Voltage (V)
UDC: DC Voltage (V)
UL1: AC Voltage Phase 1 (V)
UL2: AC Voltage Phase 2 (V)
UL3: AC Voltage Phase 3 (V)
U_AC: ?
U_L1L2: Phase1 to Phase2 Voltage (V)
U_L2L3: Phase2 to Phase3 Voltage (V)
U_L3L1: Phase3 to Phase1 Voltage (V)
```