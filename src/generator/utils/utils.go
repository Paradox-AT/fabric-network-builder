package utils

import (
	"strings"
	"text/template"

	"network-builder/src/config"
)

// GetFuncMap returns a template.FuncMap containing all shared helper functions for templates.
func GetFuncMap() template.FuncMap {
	currentID := 0
	return template.FuncMap{
		"until": func(count int) []int {
			var r []int
			for i := 0; i < count; i++ {
				r = append(r, i)
			}
			return r
		},
		"add": func(a, b int) int {
			return a + b
		},
		"multiply": func(a, b int) int {
			return a * b
		},
		"inc": func(i int) int {
			return i + 1
		},
		"toLower": strings.ToLower,
		"nextID": func() int {
			currentID++
			return currentID
		},
		"calculatePeerPort": func(orgIndex, peerIndex, offset int) int {
			return ((orgIndex + 1) * 10000) + 4000 + (peerIndex * 100) + (offset % 100)
		},
		"calculateOrdererPort": func(orgIndex, ordererIndex, offset int) int {
			return ((orgIndex + 1) * 10000) + 3000 + (ordererIndex * 100) + (offset % 100)
		},
		"calculateCAPort": func(orgIndex, offset int) int {
			return ((orgIndex + 1) * 10000) + 2000 + (offset % 100)
		},
		"calculateDatabasePort": func(orgIndex, offset int) int {
			return ((orgIndex + 1) * 10000) + 1000 + (offset % 100)
		},
		"orgHasCouchDB": func(org config.OrgConfig) bool {
			for p := 0; p < org.PeerCount; p++ {
				if org.PeerStateDatabase(p) == "couchdb" {
					return true
				}
			}
			return false
		},
	}
}
