// Copyright 2021 The golang.design Initiative Authors.
// All rights reserved. Use of this source code is governed
// by a MIT license that can be found in the LICENSE file.
//
// Written by Changkun Ou <changkun.de>

//go:build windows

package hotkey

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ZeronoFreya/go-hotkey/win"
)

type platformHotkey struct {
	mu         sync.Mutex
	hotkeyId   uint64
	registered bool
	funcs      chan func()
	canceled   chan struct{}
}

var hotkeyId uint64 // atomic

// register registers a system hotkey. It returns an error if
// the registration is failed. This could be that the hotkey is
// conflict with other hotkeys.
func (hk *Hotkey) register() error {
	hk.mu.Lock()
	if hk.registered {
		hk.mu.Unlock()
		return errors.New("hotkey already registered")
	}

	mod := uint8(0)
	for _, m := range hk.mods {
		mod = mod | uint8(m)
	}

	hk.hotkeyId = atomic.AddUint64(&hotkeyId, 1)
	hk.funcs = make(chan func())
	hk.canceled = make(chan struct{})
	go hk.handle()

	var (
		ok   bool
		err  error
		done = make(chan struct{})
	)
	hk.funcs <- func() {
		ok, err = win.RegisterHotKey(0, uintptr(hk.hotkeyId), uintptr(mod), uintptr(hk.key))
		done <- struct{}{}
	}
	<-done
	if !ok {
		close(hk.canceled)
		hk.mu.Unlock()
		return err
	}
	hk.registered = true
	hk.mu.Unlock()
	return nil
}

// unregister deregisteres a system hotkey.
func (hk *Hotkey) unregister() error {
	hk.mu.Lock()
	defer hk.mu.Unlock()
	if !hk.registered {
		return errors.New("hotkey is not registered")
	}

	done := make(chan struct{})
	hk.funcs <- func() {
		win.UnregisterHotKey(0, uintptr(hk.hotkeyId))
		done <- struct{}{}
		close(hk.canceled)
	}
	<-done

	<-hk.canceled
	hk.registered = false
	return nil
}

const (
	// wmHotkey represents hotkey message
	wmHotkey uint32 = 0x0312
	wmQuit   uint32 = 0x0012
)

// handle handles the hotkey event loop.
func (hk *Hotkey) handle() {
	// We could optimize this. So far each hotkey is served in an
	// individual thread. If we have too many hotkeys, then a program
	// have to create too many threads to serve them.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	tk := time.NewTicker(time.Second / 100)
	for range tk.C {
		msg := win.MSG{}
		if !win.PeekMessage(&msg, 0, 0, 0) {
			select {
			case f := <-hk.funcs:
				f()
			case <-hk.canceled:
				return
			default:
			}
			continue
		}
		if !win.GetMessage(&msg, 0, 0, 0) {
			return
		}

		switch msg.Message {
		case wmHotkey:
			if hk.Signal == "down" {
				hk.Callback()
			} else if hk.Signal == "up" {
				tk := time.NewTicker(time.Second / 100)
				for range tk.C {
					if win.GetAsyncKeyState(int(hk.key)) == 0 {
						hk.Callback()
						break
					}
				}
			}

		case wmQuit:
			return
		}
	}
}

// Modifier represents a modifier.
// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-registerhotkey
type Modifier uint8

// All kinds of Modifiers
const (
	ModAlt   Modifier = 0x1
	ModCtrl  Modifier = 0x2
	ModShift Modifier = 0x4
	ModWin   Modifier = 0x8
)

// Key represents a key.
// https://docs.microsoft.com/en-us/windows/win32/inputdev/virtual-key-codes
type Key uint16

// All kinds of Keys
var keyList = map[string]Key{
	"space": 0x20,
	"0":     0x30,
	"1":     0x31,
	"2":     0x32,
	"3":     0x33,
	"4":     0x34,
	"5":     0x35,
	"6":     0x36,
	"7":     0x37,
	"8":     0x38,
	"9":     0x39,
	"a":     0x41,
	"b":     0x42,
	"c":     0x43,
	"d":     0x44,
	"e":     0x45,
	"f":     0x46,
	"g":     0x47,
	"h":     0x48,
	"i":     0x49,
	"j":     0x4A,
	"k":     0x4B,
	"l":     0x4C,
	"m":     0x4D,
	"n":     0x4E,
	"o":     0x4F,
	"p":     0x50,
	"q":     0x51,
	"r":     0x52,
	"s":     0x53,
	"t":     0x54,
	"u":     0x55,
	"v":     0x56,
	"w":     0x57,
	"x":     0x58,
	"y":     0x59,
	"z":     0x5A,

	"return": 0x0D,
	"escape": 0x1B,
	"esc":    0x1B,
	"delete": 0x2E,
	"tab":    0x09,

	"left":  0x25,
	"right": 0x27,
	"up":    0x26,
	"down":  0x28,

	"f1":  0x70,
	"f2":  0x71,
	"f3":  0x72,
	"f4":  0x73,
	"f5":  0x74,
	"f6":  0x75,
	"f7":  0x76,
	"f8":  0x77,
	"f9":  0x78,
	"f10": 0x79,
	"f11": 0x7A,
	"f12": 0x7B,
	"f13": 0x7C,
	"f14": 0x7D,
	"f15": 0x7E,
	"f16": 0x7F,
	"f17": 0x80,
	"f18": 0x81,
	"f19": 0x82,
	"f20": 0x83,
}
