package main

import (
	"fmt"
	"strings"
)

// Used to handle channels as an array in the commandline arguments
type stringSet []string

type Value interface {
	String() string
	Set(string) error
}

func (i *stringSet) String() string {
	return fmt.Sprintf("%s", *i)
}

func (i *stringSet) Set(values string) error {
	// Split by comman, but only add new values
	for _, value := range strings.Split(values, ",") {
		for _, channel := range *i {
			if value == channel {
				return nil
			}
		}
		*i = append(*i, []string{value}...)
	}
	return nil
}
