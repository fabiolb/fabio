// +build !windows

package main

import (
	"log"
	"os"
	"sync"
	"time"
)

const interval = time.Hour

const warnInsecure = `

	************************************************************
	You are running fabio as root with the '-insecure' flag
	Please check https://fabiolb.net/faq/binding-to-low-ports/
	for alternatives.
	************************************************************

`

const warn17behavior = `

	************************************************************
	You are running fabio as root without the '-insecure' flag
	This will stop working with fabio 1.7!
	************************************************************

`

var once sync.Once

func WarnIfRunAsRoot(allowRoot bool) {
	// todo(fs): should we emit the same warning when running inside a container?
	// todo(fs): check for existence of `/.dockerenv` to determine Docker environment.
	isRoot := os.Getuid() == 0
	if !isRoot {
		return
	}
	doWarn(allowRoot)
	once.Do(func() { go remind(allowRoot) })
}

func doWarn(allowRoot bool) {
	warn := warnInsecure
	if !allowRoot {
		warn = warn17behavior
	}
	log.Printf("[INFO] Running fabio as UID=%d EUID=%d GID=%d", os.Getuid(), os.Geteuid(), os.Getgid())
	log.Print("[WARN] ", warn)
}

func remind(allowRoot bool) {
	for {
		doWarn(allowRoot)
		time.Sleep(interval)
	}
}
