package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"golang.org/x/sys/unix"
)

func main() {
	// open the Netlink socket
	sock, err := unix.Socket(
		unix.AF_NETLINK,             // domain
		unix.SOCK_DGRAM,             // type, datagram socket
		unix.NETLINK_KOBJECT_UEVENT, // proto/subsystem, listening to kernel uevents
	)
	if err != nil {
		log.Fatalf("Error creating socket: %v", err)
		return
	}
	defer unix.Close(sock)

	// Bind to the socket to receive all uevents (pid = 0, groups = 1)
	sa := &unix.SockaddrNetlink{
		Family: unix.AF_NETLINK,
		Groups: 1, // receive broadcast messages (UEVENT group)
	}
	if err := unix.Bind(sock, sa); err != nil {
		log.Fatalf("Error binding socket: %v", err)
	}

	fmt.Println("Listening for USB uevents... (Ctrl+C to exit)")

	buf := make([]byte, 4096)
	for {
		nr, _, err := unix.Recvfrom(sock, buf, 0)
		if err != nil {
			log.Fatalf("Error receiving message: %v", err)
		}

		msg := buf[:nr]
		fields := bytes.Split(msg, []byte{0})

		isUSB := false
		isAdd := false

		for _, f := range fields {
			if bytes.HasPrefix(f, []byte("SUBSYSTEM=usb")) {
				isUSB = true
			}
			if bytes.HasPrefix(f, []byte("ACTION=add")) {
				isAdd = true
			}
		}

		if isUSB && isAdd {
			eventData := make(map[string]string)
			for _, f := range fields {
				if len(f) == 0 {
					continue
				}
				pair := strings.SplitN(string(f), "=", 2)
				if len(pair) == 2 {
					eventData[pair[0]] = pair[1]
				}
			}

			fmt.Println("---- USB Add Event ----")
			fmt.Printf("Devpath: %s\n", eventData["DEVPATH"])
			if eventData["DEVTYPE"] == "usb_device" {
				fmt.Printf("Type: %s\n", eventData["TYPE"])
			} else if eventData["DEVTYPE"] == "usb_interface" {
				fmt.Printf("Interface: %s\n", eventData["INTERFACE"])
			}
		}
	}
}
