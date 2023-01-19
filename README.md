# go-hotkey
golang 全局热键

修改自 https://github.com/golang-design/hotkey

> 只适用于windows

|  参数   | 说明  |
|  ----  | ----  |
| w  | win |
| c  | ctrl |
| s  | shift |
| a  | alt |

使用方法: 
* **不接受大写**

* 第一个参数为热键的组合，`win`､`ctrl`､`shift`､`alt` 无顺序要求
    > 如 `csa_z` 或 `ctrl_shift_alt_z`, 如有必要, 可修改分隔符: `hotkey.SetSplitStr("+")`
* 第二个参数执行时机
    > 可选值: `down`､`up`､`press`, 默认为 `down`
    >
    > * `down` 立刻执行
    > * `up` 按键松开时执行
    > * `press` 可绑定 `按下` 与 `松开` 时的方法

```go
// 注册热键
hotkey.Register("cs_z", "up", func() {
	fmt.Println(time.Now(), "cs_z up")
})

// 解除热键
hotkey.Unregister("cs_z", "up")
```


示例:

```go
package main

import (
	"fmt"
	"time"

	"github.com/ZeronoFreya/go-hotkey"
)

func main() {

	done := make(chan bool)
    // 默认打印热键: ctrl shift z down
	hotkey.Register("cs_z", "")
    // 或
	hotkey.Register("cs_s", "up", func() {
		fmt.Println(time.Now(), "cs_s up")
	})
    // esc 注销热键绑定
	hotkey.Register("esc", "", func() {
		hotkey.Unregister("cs_z", "")
		done <- false
	})
    // 关于 press
    // 可传递第二个方法来接收 松开 事件
    // down 与 up 不会响应第二个方法
    hotkey.Register("cs_d", "press", func() {
		fmt.Println("cs_d 按下")
	}, func() {
		fmt.Println("cs_d 松开")
	})
	fmt.Println("Start!")
	<-done
}

```