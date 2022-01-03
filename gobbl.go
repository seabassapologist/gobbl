package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/godbus/dbus/v5"
)

/*
Go Bluetooth Battery Life (gobbl) -
Simple Go utility to get connected bluetooth device battery levels via the Bluez dbus-interface
*/

type childNode struct {
	Name string `xml:"name,attr"`
}

type Node struct {
	XMLName xml.Name    `xml:"node"`
	Nodes   []childNode `xml:"node"`
}

const bluezPath = "/org/bluez/hci0"

// Map bluez icon names to Font Awesome glyphs
var iconMap = map[string]string{
	"input-keyboard":         "",
	"input-gaming":           "",
	"input-mouse":            "",
	"input-tablet":           "",
	"audio-input-microphone": "",
	"audio-speakers":         "",
	"audio-headphones":       "",
	"audio-headset":          "",
	"phone":                  "",
	"default":                "",
}

type Device struct {
	name       string
	percentage int
	icon       string
	connected  bool
	paired     bool
}

type Waybar struct {
	Text    string `json:"text"`
	Tooltip string `json:"tooltip"`
	Class   string `json:"class"`
}

// Construct a Device from the provided D-Bus Object
func GetDevice(bus *dbus.Conn, obj dbus.ObjectPath) Device {

	var info map[string]dbus.Variant
	bus.Object("org.bluez", obj).Call("org.freedesktop.DBus.Properties.GetAll", 0, "org.bluez.Device1").Store(&info)

	var connec, pair bool
	info["Connected"].Store(&connec)
	info["Paired"].Store(&pair)

	perc := -1
	if pair {
		bat, _ := bus.Object("org.bluez", obj).GetProperty("org.bluez.Battery1.Percentage")

		// if device is paired, but battery level can't be read, set percentage to -1
		if bat.Value() != nil {
			bat.Store(&perc)
		}
	}

	var name, icon string = "", ""
	if info["Name"].Value() != nil {
		info["Name"].Store(&name)
	}
	if info["Icon"].Value() != nil {
		info["Icon"].Store(&icon)
	}

	return Device{name, perc, icon, connec, pair}

}

// Get list of D-Bus object paths for all viewable BT devices.
// May include devices that aren't paired or connected
func SearchAll(bus *dbus.Conn) []dbus.ObjectPath {

	var introspect string
	bus.Object("org.bluez", bluezPath).Call("org.freedesktop.DBus.Introspectable.Introspect", 0).Store(&introspect)

	var devices Node
	err := xml.Unmarshal([]byte(introspect), &devices)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var l []dbus.ObjectPath
	for _, d := range devices.Nodes {
		l = append(l, dbus.ObjectPath(bluezPath+"/"+d.Name))
	}
	return l

}

// Output one device reading per line
func Output(dl []Device) {
	for _, d := range dl {
		fmt.Printf("%v: %d%%\n", d.name, d.percentage)
	}
}

// Print out a JSON formatted line, for Waybar's 'custom' module
func OutputWaybar(dl []Device, uic bool) {
	var text, tooltip string = "", ""
	var s int = 0

	// Find longest device name to calculate tooltip padding
	for _, d := range dl {
		if len(d.name) > s {
			s = len(d.name)
		}
	}

	// Build ouput and tooltip strings
	for _, d := range dl {
		var p, ic string

		if d.percentage == -1 {
			p = "?"
		} else {
			p = strconv.Itoa(d.percentage) + "%"
		}

		if uic {
			// Set corresponding icon glyph
			if i, ok := iconMap[d.icon]; ok {
				ic = i
			} else {
				ic = iconMap["Default"]
			}
			text += fmt.Sprintf("%v %v  ", ic, p)
			tooltip += fmt.Sprintf("%v %-*s %v\n", ic, s+1, d.name+":", p)
		} else {
			// Use device name
			text += fmt.Sprintf("%v %v  ", d.name, p)
			tooltip += fmt.Sprintf("%-*s %v\n", s+1, d.name+":", p)
		}
	}
	// if no paired devices are connected display "Disconnected"
	if text == "" {
		text = "Disconnected"
		tooltip = "So lonely..."
	}

	wb, _ := json.Marshal(Waybar{
		Text:    strings.TrimSpace(text),
		Tooltip: strings.TrimSpace(tooltip),
		Class:   "$class",
	})

	fmt.Println(string(wb))

}

func main() {

	uic := flag.Bool("i", false, "Replace device name with Font Awesome icons in output")
	fm := flag.String("f", "", "Formatting for output: 'Waybar', 'None' (default)")
	flag.Parse()

	// Get dbus connection
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		fmt.Println("Failed to connect to dbus ", err)
		os.Exit(1)
	}

	// Gather list of all BT device objects currently viewable by bluez
	objl := SearchAll(conn)

	// Build list of Devices of only paired and connected devices
	var dvl []Device
	for _, o := range objl {
		var d Device = GetDevice(conn, o)
		if d.paired && d.connected {
			dvl = append(dvl, d)
		}
	}

	// Format battery level output
	switch strings.ToLower(*fm) {
	case "waybar":
		OutputWaybar(dvl, *uic)
	case "none":
		Output(dvl)
	default:
		Output(dvl)
	}

}
