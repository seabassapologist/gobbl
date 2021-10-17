package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/seabassapologist/gobbl"

	"github.com/godbus/dbus/v5"
)

/*
Go Bluetooth Battery Life (gobbl) -
Simple Go utility to get connected bluetooth device battery levels via the Bluez dbus-interface
*/

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
	objl := gobbl.SearchAll(conn)

	// Build list of Devices of only paired and connected devices
	var devl []gobbl.Device
	for _, o := range objl {
		var d *gobbl.Device = gobbl.GetDevice(conn, o)
		if d.IsPaired() && d.IsConnected() {
			devl = append(devl, *d)
		}
	}

	// Format battery level output
	if *wb {
		gobbl.OutputWaybar(devl, *uic)
	} else {
		gobbl.Output(devl)
	}

}
