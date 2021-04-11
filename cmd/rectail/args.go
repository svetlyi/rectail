package main

import (
	"fmt"
	"strings"
)

type strSlice []string

func (s *strSlice) String() string {
	return fmt.Sprint(strings.Join(*s, ","))
}

func (s *strSlice) Set(value string) error {
	for _, strPart := range strings.Split(value, ",") {
		*s = append(*s, strPart)
	}
	return nil
}
