# Go Bluetooth Battery Level (gobbl)

A simple Go utility to get and display connected bluetooth device battery levels via the Bluez's D-Bus interface. Intended for use with statusbars and tiling WMs, or other Desktop Environment extensions

![gobbl running in Waybar in Sway](docs/gobbl_waybar.png)

```
Usage of gobbl:
  -f string
        Formatting for output: 'Waybar', 'None' (default)
  -i    Replace device name with Font Awesome icons in output
```

Waybar example:
```json
"custom/btbattery": {
    "format": "{}",
    "exec": "gobbl -f waybar -i",
    "return-type": "json",
    "on-click": "gobbl -f waybar -i",
    "interval": 500
},
```

## Building

To locally, ensure you have `Go >=1.17` installed, and run the following in the source dir

```
$ go build . 
```

or 

```
$ go install .
```

to install to your `$GOPATH`

## Installing

gobbl is available in the [Arch User Repository](https://aur.archlinux.org/packages/gobbl)
