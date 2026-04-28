/*
 * (C) Copyright 2025
 *
 * SPDX-License-Identifier:     MIT
 */

package swupdate

import "C"

// CopyToCChars is an helper function to copy a Go string to a given slice of C chars.
// To copy a sized array:
// ```
// var arr [SIZE]C.char
// CopyToCChars(arr[:], str)
// ```
// Note that it aims to be used as the `copy(dst []byte, src string)` builtin, see `go doc builtin copy`.
func CopyToCChars(dst []C.char, src string) int {
	for index, b := range src {
		if index >= len(dst) {
			return len(dst)
		}
		dst[index] = C.char(b)
	}

	return len(src)
}

// GoStringFromSlice is an helper function get a Go String from a C.char slice.
func GoStringFromSlice(src []C.char) string {
	dst := make([]byte, len(src))
	for index, ch := range src {
		dst[index] = byte(ch)
	}

	return string(dst)
}
