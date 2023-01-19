// Copyright 2021 The golang.design Initiative Authors.
// All rights reserved. Use of this source code is governed
// by a MIT license that can be found in the LICENSE file.
//
// Written by Changkun Ou <changkun.de>

//go:build windows

package hotkey

import (
	"errors"
	"fmt"
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

	keyDown := false
	keyUp := false
	keyPress := false

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
			if hk.Signal == "down" || hk.Signal == "press" {
				if !keyDown {
					keyDown = true
					if hk.Signal == "press" {
						if len(hk.Callbacks) > 0 {
							hk.Callbacks[0]()
						} else {
							fmt.Println(hk.String() + " start")
						}
						keyPress = true
					} else {
						if len(hk.Callbacks) > 0 {
							hk.Callbacks[0]()
						} else {
							fmt.Println(hk.String())
						}
					}

					go func() {
						tk := time.NewTicker(time.Second / 100)
						for range tk.C {
							if win.GetAsyncKeyState(int(hk.key)) == 0 {
								keyDown = false
								if keyPress {
									if len(hk.Callbacks) > 1 {
										hk.Callbacks[1]()
									} else {
										fmt.Println(hk.String() + " end")
									}
									keyPress = false
								}
								break
							}
						}
					}()
				}
			} else if hk.Signal == "up" {
				if !keyUp {
					keyUp = true
					go func() {
						tk := time.NewTicker(time.Second / 100)
						for range tk.C {
							if win.GetAsyncKeyState(int(hk.key)) == 0 {
								if len(hk.Callbacks) > 0 {
									hk.Callbacks[0]()
								} else {
									fmt.Println(hk.String())
								}
								keyUp = false
								break
							}
						}
					}()
				}
			}
		case wmQuit:
			return
		}
	}
}
