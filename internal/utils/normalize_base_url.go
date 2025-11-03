package utils

import "strings"

func NormalizeBaseURL(u *string) {
	if strings.HasPrefix(*u, "http://") || strings.HasPrefix(*u, "https://") {
		return
	}

	*u = "http://" + *u
}
