package main

/*
Go Bluetooth Battery Life (gobbl) -
Simple Go utility to get connected bluetooth device battery levels via the Bluez dbus-interface
*/

import (
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/godbus/dbus/v5"
)

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

// Construct a Device from the provided D-Bus Object
func getDevice(bus *dbus.Conn, obj dbus.ObjectPath) Device {

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
func searchAll(bus *dbus.Conn) []dbus.ObjectPath {

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
func output(dl []Device) {
	for _, d := range dl {
		fmt.Printf("%v: %d%%\n", d.name, d.percentage)
	}
}

// Print out a JSON formatted line, for Waybar's 'custom' module
func outputWaybar(dl []Device, uic bool) {
	var text, tooltip string = "", ""
	for _, d := range dl {

		var p, ic, n string
		if d.percentage == -1 {
			p = "?"
		} else {
			p = strconv.Itoa(d.percentage) + "%"
		}

		if len(d.name) > 25 {
			n = fmt.Sprintf("%-19s", d.name[0:19]+"...:")
		} else {
			n = fmt.Sprintf("%-25s", d.name+":")
		}

		if uic {
			if i, ok := iconMap[d.icon]; ok {
				ic = i
			} else {
				ic = iconMap["Default"]
			}
			text += fmt.Sprintf("%v %v  ", ic, p)
			tooltip += fmt.Sprintf("%v %v %v\\n", ic, n, p)
		} else {
			text += fmt.Sprintf("%v %v  ", d.name, p)
			tooltip += fmt.Sprintf("%v %v\\n", n, p)
		}
	}
	// if no paired devices are connected display "Disconnected"
	if text == "" {
		text = "Disconnected"
		tooltip = "Disconnected"
	}

	fmt.Printf("{\"text\": \"%v\", \"tooltip\": \"%v\", \"class\": \"$class\"}\n", strings.TrimSpace(text), strings.Trim(tooltip, "\\n"))
}

func main() {

	uic := flag.Bool("i", false, "Replace device name with Font Awesome icons in output")
	wb := flag.Bool("w", false, "Format output as JSON for Waybar's 'custom' module")
	flag.Parse()

	// Get dbus connection
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		fmt.Println("Failed to connect to dbus ", err)
		os.Exit(1)
	}

	// Gather list of all BT device objects currently viewable by bluez
	objl := searchAll(conn)

	// Build list of Devices of only paired and connected devices
	var devl []Device
	for _, o := range objl {
		d := getDevice(conn, o)
		if d.paired && d.connected {
			devl = append(devl, d)
		}

	}

	// Format battery level output
	if *wb {
		outputWaybar(devl, *uic)
	} else {
		output(devl)
	}

}
