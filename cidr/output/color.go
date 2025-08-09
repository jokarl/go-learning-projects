package output

type color int

const (
	Reset color = iota
	Red
	Blue
	Green
	Yellow
)

var colorValue = map[color]string{
	Reset:  "\033[0m",
	Red:    "\033[31m",
	Blue:   "\033[34m",
	Green:  "\033[32m",
	Yellow: "\033[33m",
}

func (c color) String() string {
	return colorValue[c]
}
