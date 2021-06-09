package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// SolarMax inverter implementation
type SolarMax struct {
	uri      string
	inverter int
}

func main() {

	host := flag.String("host", "127.0.0.1", "host/inverter ip address")
	port := flag.Int("port", 80, "port number")
	inverter := flag.Int("inverter", 1, "inverter id")
	metrics := flag.String("metrics", "KDY,KMT,KYR,KT0,TNF,TKK,PAC,PRL,IL1,IDC,UL1,UDC,SYS", "list of metric codes (comma separated)")
	loglevel := flag.String("loglevel", "info", "info, warn, debug, trace")
	flag.Parse()

	s := &SolarMax{
		uri:      net.JoinHostPort(*host, fmt.Sprintf("%d", *port)),
		inverter: *inverter,
	}

	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	switch ll := strings.ToLower(*loglevel); ll {
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	res := s.execCmd(smQuery(*metrics, *inverter))

	resj, err := smDecode(res)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resj)
}

// execCmd executes an SolarMax command and provides the response
func (s *SolarMax) execCmd(cmd string) string {
	log.Debugf("send: %s", cmd)
	buf := bytes.NewBuffer([]byte(cmd))

	// Open connection to SolarMax inverter
	conn, err := net.DialTimeout("tcp", s.uri, 5*time.Second)
	if err != nil {
		// return "{01;FB;78|64:KDY=2E;KMT=62;KYR=699;KT0=B256;TNF=1388;TKK=20;PAC=10A;PRL=2;IL1=3A;IDC=31;UL1=908;UDC=D12;SYS=4E28,0|1C86}", nil
		return "{01;FB;00|64:SYS=752F|0000}"
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// Send command
	if _, err = buf.WriteTo(conn); err != nil {
		return "{01;FB;00|64:SYS=752E|0000}"
	}

	// Read response
	resp := make([]byte, 8192)
	len, err := conn.Read(resp)
	if err != nil {
		return "{01;FB;00|64:SYS=752D|0000}"
	}

	// Create response message
	for i := 0; i < len; i++ {
		_ = buf.WriteByte(resp[i])
	}
	log.Debugf("recv: %s\n", buf.String())

	return buf.String()
}

func smQuery(metrics string, inverter int) string {
	// Prepare metric query
	const qtype int = 100 // query type

	metricLst := strings.Split(strings.ToUpper(strings.TrimSpace(metrics)), ",")
	metricDesc := smMetricDesc()
	var cmdLst []string
	for i := 0; i < len(metricLst); i++ {
		if metricDesc[metricLst[i]] != "" {
			cmdLst = append(cmdLst, metricLst[i])
		} else {
			log.Warnf("unknown metric: <%s>\n", metricLst[i])
		}
	}

	// Assure always a multiple of 4 metrics
	cmdlen := len(cmdLst)
	cmdoptlen := 4 * math.Ceil((float64(cmdlen) / float64(4)))
	// Complete missing metrics by using the system status to get a multiple of 4
	for i := 0; i < int(cmdoptlen)-cmdlen; i++ {
		cmdLst = append(cmdLst, "SYS")
	}

	query := strings.Join(cmdLst, ";")
	query = fmt.Sprintf("|%02x:%s|", qtype, query)
	query = fmt.Sprintf("FB;%02x;%02x%s", inverter, 9+len(query)+5, query)
	query = strings.ToUpper(fmt.Sprintf("{%s%04x}", query, smChksum(query)))

	// query = "{FB;01;46|64:KDY;KMT;KYR;KT0;TNF;TKK;PAC;PRL;IL1;IDC;UL1;UDC;SYS|1199}"

	return query
}

// smDecode decodes SolarMax protocol response
func smDecode(raw string) (string, error) {

	invs := strings.Split(strings.Split(strings.Split(raw, "|")[0], ";")[0], "{")[1]
	inv, err := strconv.ParseInt(invs, 16, 64)
	if err != nil {
		return "", err
	}
	log.Debugf("inverter id: %02d\n", inv)

	data := strings.Split(raw, "|")[1]
	adata := strings.Split(strings.Split(data, ":")[1], ";")

	mdata := make(map[string]interface{})
	// increment map's value for every key from slice
	for i := 0; i < len(adata); i++ {
		keyval := strings.Split(adata[i], "=")
		key := keyval[0]
		val := strings.Split(keyval[1], ",")[0]
		valint, err := strconv.ParseInt(val, 16, 64)
		if err != nil {
			return "", err
		}
		// Default int 16 mapping of SolarMax response
		mdata[key] = valint

		// value type dependend conversions
		// power of frequency reading, then  divide by 2...for some reason
		statusdesc := smStatus()
		if key == "SYS" {
			mdata[key] = fmt.Sprintf("%03d-", valint-20000) + statusdesc[valint]
		}
		// power of frequency reading, then  divide by 2...for some reason
		valdiv2 := map[string]bool{
			"PAC": true,
			"PDC": true,
		}
		if valdiv2[key] {
			mdata[key] = float64(valint) / 2
		}
		// voltage reading or frequency, divide by 10 to get Volts
		// same for "energy today"
		valdiv10 := map[string]bool{
			"UL1":  true,
			"UL2":  true,
			"UL3":  true,
			"KDY":  true,
			"UD01": true,
			"UD02": true,
			"UD03": true,
		}
		if valdiv10[key] {
			mdata[key] = float64(valint) / 10
		}

		// current readings or frequency, divide by 100 to get Amps
		valdiv100 := map[string]bool{
			"IL1":  true,
			"IL2":  true,
			"IL3":  true,
			"IDC":  true,
			"TNF":  true,
			"ID01": true,
			"ID02": true,
			"ID03": true,
		}
		if valdiv100[key] {
			mdata[key] = float64(valint) / 100
		}
		keydesc := smMetricDesc()
		log.Tracef("%d. %s value: %v", i+1, key+"-"+keydesc[key], mdata[key])

		// stop decode loop when current key = next key (filled up to get a multiple of 4)
		if i+1 < len(adata) {
			if nextkey := strings.Split(adata[i+1], "=")[0]; key == nextkey {
				break
			}
		}

	}
	// Marshal sData into a JSON string.
	sData, err := json.Marshal(mdata)
	if err != nil {
		return "{}", err
	}

	return string(sData), err
}

// smChksum creates a simple checksum
func smChksum(data string) int {
	smChksum := 0
	for i := 0; i < len(data); i++ {
		smChksum = smChksum + int(data[i])
	}
	return smChksum
}

// smMetricDesc provides metric descriptions
func smMetricDesc() map[string]string {
	metricdesc := map[string]string{
		"ADR":    "Address",
		"BDN":    "Build number",
		"CAC":    "Start Ups (?)",
		"DDY":    "Date day",
		"DIN":    "Date in integer format with offset 23.12.1510",
		"DMT":    "Date month",
		"DYR":    "Date year",
		"EC00":   "Error Code 0",
		"EC01":   "Error Code 1",
		"EC02":   "Error Code 2",
		"EC03":   "Error Code 3",
		"EC04":   "Error Code 4",
		"EC05":   "Error Code 5",
		"EC06":   "Error Code 6",
		"EC07":   "Error Code 7",
		"EC08":   "Error Code 8",
		"FDAT":   "datetime ?",
		"F_AC":   "Grid Frequency",
		"ID01":   "String 1 Current (A)",
		"ID02":   "String 2 Current (A)",
		"ID03":   "String 3 Current (A)",
		"IDC":    "DC Current (A)",
		"IL1":    "AC Current Phase 1 (A)",
		"IL2":    "AC Current Phase 2 (A)",
		"IL3":    "AC Current Phase 3 (A)",
		"KDL":    "Energy yesterday (Wh)",
		"KDY":    "Energy today (kWh)",
		"KHR":    "Operating Hours",
		"KLM":    "Energy last month (kWh)",
		"KLY":    "Energy last year (kWh)",
		"KMT":    "Energy this month (kWh)",
		"KT0":    "Total Energy(kWh)",
		"KYR":    "Energy this year (kWh)",
		"LAN":    "Language",
		"MAC":    "MAC Address",
		"PAC":    "AC Power (W)",
		"PDC":    "DC Power (W)",
		"PIN":    "Installed Power (W)",
		"PRL":    "Relative power (%)",
		"SAL":    "System Alarms",
		"SDAT":   "datetime ?",
		"SE1":    "", // Response delivers only value but no key and will be ignored
		"SWV":    "Software Version",
		"SYS":    "System Status",
		"THR":    "Time hours",
		"TKK":    "Inverter Temperature (C)",
		"TMI":    "Time minutes",
		"TNF":    "Generated Frequency (Hz)",
		"TNP":    "Grid period duration",
		"TYP":    "Type",
		"UD01":   "String 1 Voltage (V)",
		"UD02":   "String 2 Voltage (V)",
		"UD03":   "String 3 Voltage (V)",
		"UDC":    "DC Voltage (V)",
		"UL1":    "AC Voltage Phase 1 (V)",
		"UL2":    "AC Voltage Phase 2 (V)",
		"UL3":    "AC Voltage Phase 3 (V)",
		"U_AC":   "?",
		"U_L1L2": "Phase1 to Phase2 Voltage (V)",
		"U_L2L3": "Phase2 to Phase3 Voltage (V)",
		"U_L3L1": "Phase3 to Phase1 Voltage (V)",
	}

	return metricdesc
}

// smStatus provides status descriptions
func smStatus() map[int64]string {
	statusdesc := map[int64]string{
		20001: "Running",
		20002: "Irradiance too low",
		20003: "Startup",
		20004: "MPP operation",
		20006: "Maximum power",
		20007: "Temperature limitation",
		20008: "Mains operation",
		20009: "Idc limitation",
		20010: "Iac limitation",
		20011: "Test mode",
		20012: "Remote controlled",
		20013: "Restart delay",
		20014: "External limitation",
		20015: "Frequency limitation",
		20016: "Restart limitation",
		20017: "Booting",
		20018: "Insufficient boot power",
		20019: "Insufficient power",
		20021: "Uninitialized",
		20022: "Disabled",
		20023: "Idle",
		20024: "Powerunit not ready",
		20050: "Program firmware",
		20101: "Device error 101",
		20102: "Device error 102",
		20103: "Device error 103",
		20104: "Device error 104",
		20105: "Insulation fault DC",
		20106: "Insulation fault DC",
		20107: "Device error 107",
		20108: "Device error 108",
		20109: "Vdc too high",
		20110: "Device error 110",
		20111: "Device error 111",
		20112: "Device error 112",
		20113: "Device error 113",
		20114: "Ierr too high",
		20115: "No mains",
		20116: "Frequency too high",
		20117: "Frequency too low",
		20118: "Mains error",
		20119: "Vac 10min too high",
		20120: "Device error 120",
		20121: "Device error 121",
		20122: "Vac too high",
		20123: "Vac too low",
		20124: "Device error 124",
		20125: "Device error 125",
		20126: "Error ext. input 1",
		20127: "Fault ext. input 2",
		20128: "Device error 128",
		20129: "Incorr. rotation dir.",
		20130: "Device error 130",
		20131: "Main switch off",
		20132: "Device error 132",
		20133: "Device error 133",
		20134: "Device error 134",
		20135: "Device error 135",
		20136: "Device error 136",
		20137: "Device error 137",
		20138: "Device error 138",
		20139: "Device error 139",
		20140: "Device error 140",
		20141: "Device error 141",
		20142: "Device error 142",
		20143: "Device error 143",
		20144: "Device error 144",
		20145: "df/dt too high",
		20146: "Device error 146",
		20147: "Device error 147",
		20148: "Device error 148",
		20150: "Ierr step too high",
		20151: "Ierr step too high",
		20153: "Device error 153",
		20154: "Shutdown 1",
		20155: "Shutdown 2",
		20156: "Device error 156",
		20157: "Insulation fault DC",
		20158: "Device error 158",
		20159: "Device error 159",
		20160: "Device error 160",
		20161: "Device error 161",
		20163: "Device error 163",
		20164: "Ierr too high",
		20165: "No mains",
		20166: "Frequency too high",
		20167: "Frequency too low",
		20168: "Mains error",
		20169: "Vac 10min too high",
		20170: "Device error 170",
		20171: "Device error 171",
		20172: "Vac too high",
		20173: "Vac too low",
		20174: "Device error 174",
		20175: "Device error 175",
		20176: "Error DC polarity",
		20177: "Device error 177",
		20178: "Device error 178",
		20179: "Device error 179",
		20180: "Vdc too low",
		20181: "Blocked external",
		20185: "Device error 185",
		20186: "Device error 186",
		20187: "Device error 187",
		20188: "Device error 188",
		20189: "L and N interchanged",
		20190: "Below-average yield",
		20191: "Limitation error",
		20198: "Device error 198",
		20199: "Device error 199",
		20999: "Device error 999",

		// Custom errors to handle go errors
		29997: "Inverter response read error",
		29998: "Inverter network send timeout",
		29999: "Inverter network i/o timeout",
	}

	return statusdesc
}
