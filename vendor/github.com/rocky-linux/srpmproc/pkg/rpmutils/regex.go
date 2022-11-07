package rpmutils

import "regexp"

// Nvr is a regular expression that matches a NVR.
var Nvr = regexp.MustCompile("^(\\S+)-([\\w~%.+]+)-(\\w+(?:\\.[\\w+]+)+?)(?:\\.rpm)?$")
