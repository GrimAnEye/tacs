package server

import "regexp"

func loginValidation(login string) string {
	reg := regexp.MustCompile(`^\s*(([a-z0-9]+[\.\-_]?)+[a-z0-9])\s*$`)
	name := reg.FindAllStringSubmatch(login, -1)
	if len(name) == 0 {
		return ""
	}
	return name[0][1]
}
