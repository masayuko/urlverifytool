package cmd

import (
	"fmt"
	"os"
	"unicode"
)

func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(-1)
}

var alphanumConv = unicode.SpecialCase{
	// numbers
	unicode.CaseRange{
		Lo: 0xff10, // '全角0'
		Hi: 0xff19, // '全角9'
		Delta: [unicode.MaxCase]rune{
			0,
			0x0030 - 0xff10, // '0' - '全角0'
			0,
		},
	},
	// uppercase Letters
	unicode.CaseRange{
		Lo: 0xff21, // '全角A'
		Hi: 0xff3a, // '全角Z'
		Delta: [unicode.MaxCase]rune{
			0,
			0x0041 - 0xff21, // 'A' - '全角A'
			0,
		},
	},
	// lowercase letters
	unicode.CaseRange{
		Lo: 0xff41, // '全角a'
		Hi: 0xff5a, // '全角z'
		Delta: [unicode.MaxCase]rune{
			0,
			0x0061 - 0xff41, // 'a' - '全角a'
			0,
		},
	},
}
