package color

import "fmt"

func wrapEscape(code string, x string) string {
	return fmt.Sprintf("%s%s%s", escape("31"), x, escape("0"))
}

func escape(code string) string {
	return fmt.Sprintf("\033[%sm", code)
}

func Red(x string) string {
	return fmt.Sprintf("%s%s%s", escape("31"), x, escape("0"))
}
