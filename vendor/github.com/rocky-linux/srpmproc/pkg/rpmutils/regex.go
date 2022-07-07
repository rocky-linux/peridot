package rpmutils

import "regexp"

var (
	Nvr        = regexp.MustCompile("^(\\S+)-([\\w~%.+]+)-(\\w+(?:\\.[\\w+]+)+?)(?:\\.(\\w+))?(?:\\.rpm)?$")
	epoch      = regexp.MustCompile("(\\d+):")
	module     = regexp.MustCompile("^(.+)-(.+)-([0-9]{19})\\.((?:.+){8})$")
	dist       = regexp.MustCompile("(\\.el\\d(?:_\\d|))")
	moduleDist = regexp.MustCompile("\\.module.+$")
)
