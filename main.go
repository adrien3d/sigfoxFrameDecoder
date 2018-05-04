package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type SigfoxMessage struct {
	Id          string  `json:"id" bson:"_id,omitempty" valid:"-"`
	SigfoxId    string  `json:"sigfoxId" bson:"sigfoxId" valid:"-"`
	FrameNumber uint    `json:"frameNumber" bson:"frameNumber" valid:"-"` //Device : (daily frames under 140)
	Timestamp   int64   `json:"timestamp" bson:"timestamp" valid:"-"`     //Sigfox : time
	Station     string  `json:"station" bson:"station" valid:"-"`         //Sigfox : station
	Snr         float64 `json:"snr" bson:"snr" valid:"-"`                 //Sigfox : snr
	AvgSnr      float64 `json:"avgSnr" bson:"avgSnr" valid:"-"`           //Sigfox : avgSnr
	Rssi        float64 `json:"rssi" bson:"rssi" valid:"-"`               //Sigfox : rssi
	MesType     uint8   `json:"mesType" bson:"mesType" valid:"-"`         //Sigfox : mesType
	Data        string  `json:"data" bson:"data" valid:"-"`               //Sigfox : data
	EventType   string  `json:"eventType" bson:"eventType" valid:"-"`     //Device : eventType
	SwRev       string  `json:"swRev" bson:"swRev" valid:"-"`             //Device : swRev
	Mode        string  `json:"mode" bson:"mode" valid:"-"`               //Device : mode
	Timeframe   string  `json:"timeframe" bson:"timeframe" valid:"-"`     //Device : timeframe
	Data1       float64 `json:"data1" bson:"data1" valid:"-"`             //Device : battery
	Data2       float64 `json:"data2" bson:"data2" valid:"-"`             //Device : temperature
	Data3       float64 `json:"data3" bson:"data3" valid:"-"`             //Device : humidity
	Data4       float64 `json:"data4" bson:"data4" valid:"-"`             //Device : light
	Data5       float64 `json:"data5" bson:"data5" valid:"-"`             //Device : custom
	Data6       float64 `json:"data6" bson:"data6" valid:"-"`             //Device : custom
	Alerts      int64   `json:"alerts" bson:"alerts" valid:"-"`           //Device : alerts
}

type Location struct {
	Id          string  `json:"id" bson:"_id,omitempty" valid:"-"`
	SigfoxId    string  `json:"sigfoxId" bson:"sigfoxId" valid:"-"`
	FrameNumber uint    `json:"frameNumber" bson:"frameNumber" valid:"-"` //Device : (daily frames under 140)
	Timestamp   int64   `json:"timestamp" bson:"timestamp" valid:"-"`
	Latitude    float64 `json:"latitude" bson:"latitude" valid:"-"`
	Longitude   float64 `json:"longitude" bson:"longitude" valid:"-"`
	Radius      float64 `json:"radius" bson:"radius" valid:"-"`
	SpotIt      bool    `json:"spotIt" bson:"spotIt" valid:"-"`
	GPS         bool    `json:"gps" bson:"gps" valid:"-"`
	WiFi        bool    `json:"wifi" bson:"wifi" valid:"-"`
}

func decodeWisolGPSFrame(msg SigfoxMessage) (Location, float64, bool) {
	fmt.Println("_____________________________________________________________________________________________________________________")
	fmt.Print("GPS frame: \t\t\t")
	var gpsLoc Location
	var temperature float64
	var status bool
	var latitude, longitude float64
	var latDeg, latMin, latSec float64
	var lngDeg, lngMin, lngSec float64

	isNorth, isEast := false, false
	if string(msg.Data[0:2]) == "4e" {
		isNorth = true
	}
	if string(msg.Data[10:12]) == "45" {
		isEast = true
	}

	if isNorth {
		fmt.Print("N:")
	} else {
		fmt.Print("S:")
	}

	valLatDeg, _ := strconv.ParseInt(msg.Data[2:4], 16, 8)
	latDeg = float64(valLatDeg)
	valLatMin, _ := strconv.ParseInt(msg.Data[4:6], 16, 8)
	latMin = float64(valLatMin)
	valLatSec, _ := strconv.ParseInt(msg.Data[6:8], 16, 8)
	latSec = float64(valLatSec)
	fmt.Print(latDeg, "° ", latMin, "m ", latSec, "s\t")

	latitude = float64(latDeg) + float64(latMin/60) + float64(latSec/3600)

	if isEast {
		fmt.Print("E:")
	} else {
		fmt.Print("W:")
	}
	valLngDeg, _ := strconv.ParseInt(msg.Data[10:12], 16, 8)
	lngDeg = float64(valLngDeg)
	valLngMin, _ := strconv.ParseInt(msg.Data[12:14], 16, 8)
	lngMin = float64(valLngMin)
	valLngSec, _ := strconv.ParseInt(msg.Data[14:16], 16, 8)
	lngSec = float64(valLngSec)
	fmt.Print(lngDeg, "° ", lngMin, "m ", lngSec, "s")

	longitude = float64(lngDeg) + float64(lngMin/60) + float64(lngSec/3600)

	fmt.Print("\t\t\t Lat: ", latitude, "\t Lng:", longitude)
	// Populating returned location
	gpsLoc.Latitude = latitude
	gpsLoc.Longitude = longitude
	gpsLoc.FrameNumber = msg.FrameNumber
	gpsLoc.SpotIt = false
	gpsLoc.GPS = true
	gpsLoc.WiFi = false

	if msg.Data[18:20] == "41" {
		status = true
	} else if msg.Data[18:20] == "56" {
		status = false
	}

	temperature, err := strconv.ParseFloat(msg.Data[20:22], 64)
	if err != nil {
		fmt.Println("Error while converting temperature main")
	}
	dec, err := strconv.ParseFloat(msg.Data[22:24], 64)
	if err != nil {
		fmt.Println("Error while converting temperature decimal")
	}

	temperature += dec * 0.01

	fmt.Println("\t\t", gpsLoc, "\t", temperature, '\t', status)
	return gpsLoc, temperature, status
}

func decodeSensitGPSFrame(msg SigfoxMessage) {
	fmt.Println(len(msg.Data))
	parsed, err := strconv.ParseUint(msg.Data, 16, 32)
	if err != nil {
		log.Fatal(err)
	}
	pars := fmt.Sprintf("%08b", parsed)
	fmt.Println(len(pars))

	if len(pars) == 25 {
		fmt.Println("Low battery")
	}

	byte1 := pars[0:8]
	byte2 := pars[8:16]
	byte3 := pars[16:24]
	byte4 := pars[24:32]

	//Byte 1
	mode, _ := strconv.ParseInt(pars[5:8], 2, 8)
	timeframe, _ := strconv.ParseInt(pars[3:5], 2, 8)
	eventType, _ := strconv.ParseInt(pars[1:3], 2, 8)
	batteryMsb := pars[0:1]

	//Byte 2
	temperatureMsb := pars[8:12]
	batteryLsb := pars[12:16]
	battData := []string{batteryMsb, batteryLsb}
	battery, _ := strconv.ParseInt(strings.Join(battData, ""), 2, 8)
	batVal := float32(battery) * 0.05 * 2.7

	//Byte 3
	temperature := int64(0)
	tempVal := float32(0)

	reedSwitch := false
	if mode == 0 || mode == 1 {
		temperatureLsb := pars[18:24]
		tempData := []string{temperatureMsb, temperatureLsb}
		temperature, _ := strconv.ParseInt(strings.Join(tempData, ""), 2, 16)
		tempVal = (float32(temperature) - 200) / 8
		if pars[17] == 1 {
			reedSwitch = true
		}
	} else {
		temperature, _ = strconv.ParseInt(temperatureMsb, 2, 16)
		tempVal = (float32(temperature) - 200) / 8
	}

	modeStr := ""
	swRev := ""
	humidity := 0.0
	light := float32(0.0)
	switch mode {
	case 0:
		modeStr = "Button"
		majorSwRev, _ := strconv.ParseInt(pars[24:28], 2, 8)
		minorSwRev, _ := strconv.ParseInt(pars[28:32], 2, 8)
		swRev = fmt.Sprintf("%d.%d", majorSwRev, minorSwRev)
	case 1:
		modeStr = "Temperature + Humidity"
		humi, _ := strconv.ParseInt(pars[24:32], 2, 16)
		humidity = float64(humi) * 0.5
	case 2:
		modeStr = "Light"
		lightVal, _ := strconv.ParseInt(pars[18:24], 2, 8)
		lightMulti, _ := strconv.ParseInt(pars[17:18], 2, 8)
		light = float32(lightVal) * 0.01
		if lightMulti == 1 {
			light = light * 8
		}
	case 3:
		modeStr = "Door"
	case 4:
		modeStr = "Move"
	case 5:
		modeStr = "Reed switch"
	default:
		modeStr = ""
	}

	timeStr := ""
	switch timeframe {
	case 0:
		timeStr = "10 mins"
	case 1:
		timeStr = "1 hour"
	case 2:
		timeStr = "6 hours"
	case 3:
		timeStr = "24 hours"
	default:
		timeStr = ""
	}

	typeStr := ""
	switch eventType {
	case 0:
		typeStr = "Regular, no alert"
	case 1:
		typeStr = "Button call"
	case 2:
		typeStr = "Alert"
	case 3:
		typeStr = "New mode"
	default:
		timeStr = ""
	}

	//fmt.Println(data)
	fmt.Println("_____________________________________________________________________________________________________________________")
	fmt.Println("Raw data :", byte1, byte2, byte3, byte4)
	fmt.Println("Mode", mode, ":", modeStr, "\t\t", "Event type", eventType, ":", typeStr, "\t\t", "Timeframe", timeframe, ":", timeStr)
	fmt.Println("Battery :", batVal, "V\t\t")
	switch mode {
	case 0:
		fmt.Println("v" + swRev)
		fmt.Println("Temperature :", tempVal, "°C")
	case 1:
		fmt.Println(humidity, "% RH")
		fmt.Println("Temperature :", tempVal, "°C")
	case 2:
		fmt.Println(light, "lux")
		alerts, _ := strconv.ParseInt(pars[24:32], 2, 16)
		fmt.Println("Number of alerts :", alerts)
	case 3, 4, 5:
		alerts, _ := strconv.ParseInt(pars[24:32], 2, 16)
		fmt.Println("Number of alerts :", alerts)
	}
	if reedSwitch {
		fmt.Println("Reed switch on")
	}
}

func main() {
	for i := 1; i <= len(os.Args[1:]); i++ {
		var msg SigfoxMessage
		msg.Data = os.Args[i]
		frameLength := len(msg.Data)
		if frameLength <= 8 {
			fmt.Println("Sensit message")
			decodeSensitGPSFrame(msg)
		} else {
			fmt.Println("Wisol message")
			decodeWisolGPSFrame(msg)
		}
	}
}
