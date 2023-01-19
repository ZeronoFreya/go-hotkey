package hotkey

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

var keyCodeWin = map[string]Key{
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
