// Copyright 2021 The golang.design Initiative Authors.
// All rights reserved. Use of this source code is governed
// by a MIT license that can be found in the LICENSE file.
//
// Written by Changkun Ou <changkun.de>

// Package hotkey provides the basic facility to register a system-level
// global hotkey shortcut so that an application can be notified if a user
// triggers the desired hotkey. A hotkey must be a combination of modifiers
// and a single key.
//
// Note platform specific details:
//
//   - On macOS, due to the OS restriction (other platforms does not have
//     this restriction), hotkey events must be handled on the "main thread".
//     Therefore, in order to use this package properly, one must start an
//     OS main event loop on the main thread, For self-contained applications,
//     using [mainthread] package.
//     is possible. It is uncessary or applications based on other GUI frameworks,
//     such as fyne, ebiten, or Gio. See the "[examples]" for more examples.
//
//   - On Linux (X11), when AutoRepeat is enabled in the X server, the
//     Keyup is triggered automatically and continuously as Keydown continues.
//
//   - On Linux (X11), some keys may be mapped to multiple Mod keys. To
//     correctly register the key combination, one must use the correct
//     underlying keycode combination. For example, a regular Ctrl+Alt+S
//     might be registered as: Ctrl+Mod2+Mod4+S.
//
//   - If this package did not include a desired key, one can always provide
//     the keycode to the API. For example, if a key code is 0x15, then the
//     corresponding key is `hotkey.Key(0x15)`.
//
// THe following is a minimum example:
//
//	package main
//
//	import (
//		"log"
//
//		"golang.design/x/hotkey"
//		"golang.design/x/hotkey/mainthread"
//	)
//
//	func main() { mainthread.Init(fn) } // Not necessary when use in Fyne, Ebiten or Gio.
//	func fn() {
//		hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyS)
//		err := hk.Register()
//		if err != nil {
//			log.Fatalf("hotkey: failed to register hotkey: %v", err)
//		}
//
//		log.Printf("hotkey: %v is registered\n", hk)
//		<-hk.Keydown()
//		log.Printf("hotkey: %v is down\n", hk)
//		<-hk.Keyup()
//		log.Printf("hotkey: %v is up\n", hk)
//		hk.Unregister()
//		log.Printf("hotkey: %v is unregistered\n", hk)
//	}
//
// [mainthread]: https://pkg.go.dev/golang.design/x/hotkey/mainthread
// [examples]: https://github.com/golang-design/hotkey/tree/main/examples
package hotkey

import (
	"errors"
	"runtime"
	"strings"
)

// Event represents a hotkey event
type Event struct{}

// Hotkey is a combination of modifiers and key to trigger an event
type Hotkey struct {
	platformHotkey

	Signal    string
	Callbacks []func()

	mods []Modifier
	key  Key
}

var splitStr = "_"

func SetSplitStr(str string) {
	splitStr = str
}

var registeredHotkey = make(map[string]*Hotkey)

// New creates a new hotkey for the given modifiers and keycode.
func New(mods []Modifier, key Key) *Hotkey {
	hk := &Hotkey{
		mods: mods,
		key:  key,
	}

	// Make sure the hotkey is unregistered when the created
	// hotkey is garbage collected.
	runtime.SetFinalizer(hk, func(x interface{}) {
		hk := x.(*Hotkey)
		hk.unregister()
	})
	return hk
}

func getHkInfo(hkStr, signalStr string) (modifierSort, keyName, signal string) {
	hkList := strings.Split(hkStr, splitStr)
	hkLen := len(hkList)
	if hkLen == 0 {
		return modifierSort, keyName, signal
	}
	k0 := [4]string{}
	if hkLen == 2 {
		// 简写形式: csa_a
		keyName = hkList[1]
		for _, v := range strings.Split(hkList[0], "") {
			switch v {
			case "w":
				k0[0] = "w"
			case "c":
				k0[1] = "c"
			case "s":
				k0[2] = "s"
			case "a":
				k0[3] = "a"
			}
		}
	} else {
		for _, v := range hkList {
			switch v {
			case "win":
				k0[0] = "w"
			case "ctrl":
				k0[1] = "c"
			case "shift":
				k0[2] = "s"
			case "alt":
				k0[3] = "a"
			default:
				keyName = v
			}
		}
	}
	modifierSort = strings.Join(k0[:], "")

	signal = "down"
	if signalStr == "up" || signalStr == "press" {
		signal = signalStr
	}

	return modifierSort, keyName, signal
}

func getModifier(modifier string) (mod []Modifier) {
	if len(modifier) > 0 {
		for _, v := range strings.Split(modifier, "") {
			switch v {
			case "w":
				mod = append(mod, ModWin)
			case "c":
				mod = append(mod, ModCtrl)
			case "s":
				mod = append(mod, ModShift)
			case "a":
				mod = append(mod, ModAlt)
			}
		}
	}

	return mod
}

func Register(modifier, key string, callbacks ...func()) error {
	modifierSort, keyName, signal := getHkInfo(modifier, key)

	keyCode, ok := keyCodeWin[keyName]
	if !ok {
		return errors.New("key error")
	}

	mod := getModifier(modifierSort)

	hk := New(mod, keyCode)
	hk.Signal = signal
	hk.Callbacks = callbacks

	registeredHotkey[modifierSort+keyName+signal] = hk

	err := hk.register()
	if err != nil {
		return err
	}
	return nil
}

func Unregister(modifier, key string) error {
	modifierSort, keyName, signal := getHkInfo(modifier, key)
	k := modifierSort + keyName + signal

	hk, ok := registeredHotkey[k]
	if !ok {
		return nil
	}

	err := hk.unregister()
	if err != nil {
		return err
	}

	delete(registeredHotkey, k)
	return nil
}

// String returns a string representation of the hotkey.
func (hk *Hotkey) String() string {
	s := [6]string{}
	for _, mod := range hk.mods {
		if mod&ModWin != 0 {
			s[0] = "win"
		} else if mod&ModCtrl != 0 {
			s[1] = "ctrl"
		} else if mod&ModShift != 0 {
			s[2] = "shift"
		} else if mod&ModAlt != 0 {
			s[3] = "alt"
		}
	}

	for k, v := range keyCodeWin {
		if v == hk.key {
			s[4] = k
			break
		}
	}

	s[5] = hk.Signal

	return strings.Join(s[:], " ")
}
