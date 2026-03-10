package main

import (
	"fmt"
	"godshell/config"
	"os"
)

func main() {
	fmt.Printf("UID: %d, EUID: %d\n", os.Getuid(), os.Geteuid())
	fmt.Printf("SUDO_USER: %s, SUDO_UID: %s\n", os.Getenv("SUDO_USER"), os.Getenv("SUDO_UID"))
	fmt.Printf("DBUS_SESSION_BUS_ADDRESS (before): %s\n", os.Getenv("DBUS_SESSION_BUS_ADDRESS"))

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Config LOAD failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("DBUS_SESSION_BUS_ADDRESS (after): %s\n", os.Getenv("DBUS_SESSION_BUS_ADDRESS"))
	fmt.Printf("Config Path: %s\n", config.ConfigPath())

	if cfg.APIKey == "" {
		fmt.Println("API Key: NOT FOUND")
	} else {
		fmt.Printf("API Key FOUND! Length: %d\n", len(cfg.APIKey))
	}
}
