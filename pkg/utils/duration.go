package utils

import "regexp"

var reDuration = regexp.MustCompile(`(?P<duration>[1-9]\d*)(?P<unit>\w+)`)
