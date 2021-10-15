package main

// Go Bluetooth Battery Life (gobbl) -
// Simple Go utility to get connected bluetooth device battery levels via
// the Bluez dbus-interface

import (
	"encoding/xml"
	"fmt"

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

	perc, _ := bus.Object("org.bluez", obj).GetProperty("org.bluez.Battery1.Percentage")
	_ = bus.Object("org.bluez", obj).Call("org.freedesktop.DBus.Properties.GetAll", 0, "org.bluez.Device1").Store(&info)

	var con, pair bool
	info["Connected"].Store(&con)
	info["Paired"].Store(&pair)

	var per int
	if perc.Value() == nil {
		per = 0
	} else {
		perc.Store(&per)
	}

	return Device{info["Name"].String(), per, info["Icon"].String(), con, pair}

}

func getAllPaired(bus *dbus.Conn) []dbus.ObjectPath {

	var introspect string
	_ = bus.Object("org.bluez", bluezPath).Call("org.freedesktop.DBus.Introspectable.Introspect", 0).Store(&introspect)

	var devices Node
	_ = xml.Unmarshal([]byte(introspect), &devices)

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
	conn, _ := dbus.ConnectSystemBus()

	objl := getAllPaired(conn)

	var devl []Device
	for _, o := range objl {
		devl = append(devl, getDevice(conn, o))
	}

	printOutput(devl)

}
