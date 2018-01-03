package config

import (
	"testing"
)

func TestHasColons(t *testing.T) {
	containsColon := hasColon(":8000")
	if containsColon == false {
		t.Errorf("hasColon returned false when string has colon. Want: true")
	}
	containsColon = hasColon("8000")
	if containsColon == true {
		t.Errorf("hasColon returned true when string does not have a colon. Want: false")
	}

}

func TestReadConfiguration(t *testing.T) {

}
