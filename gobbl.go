package main

// Go Bluetooth Battery Life (gobbl) -
// Simple Go utility to get connected bluetooth device battery levels via
// the Bluez dbus-interface

import (
	"encoding/xml"
	"fmt"
	"os"

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

type Device struct {
	name       string
	percentage int
	icon       string
	connected  bool
	paired     bool
}

func getDevice(bus *dbus.Conn, obj dbus.ObjectPath) Device {

	var info map[string]dbus.Variant
	bus.Object("org.bluez", obj).Call("org.freedesktop.DBus.Properties.GetAll", 0, "org.bluez.Device1").Store(&info)

	var connec, pair bool
	info["Connected"].Store(&connec)
	info["Paired"].Store(&pair)

	perc := -1
	if pair {
		bat, _ := bus.Object("org.bluez", obj).GetProperty("org.bluez.Battery1.Percentage")

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

func printOutput(dl []Device) {
	for _, d := range dl {
		fmt.Printf("%v: %d%% ", d.name, d.percentage)
	}
	fmt.Println()
}

func main() {

	// Get dbus connection
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		fmt.Println("Failed to connect to dbus ", err)
		os.Exit(1)
	}

	objl := searchAll(conn)

	var devl []Device
	for _, o := range objl {
		d := getDevice(conn, o)
		if d.paired {
			devl = append(devl, d)
		}

	}

	printOutput(devl)

}
