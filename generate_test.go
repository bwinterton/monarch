package main

import "testing"

func TestRemoveNamespace(t *testing.T) {

	cases := []struct {
		in, want string
	}{
		{"blah/testing", "blah-testing"},
		{"blah/blah", "blah"},
		{"blah/blah/blah", "blah"},
		{"who/are/you", "who-are-you"},
		{"back/in-black", "back-in-black"},
		{"sugar/sugar/yes-please", "sugar-yes-please"},
	}

	for _, c := range cases {
		got := removeNamespace(c.in)
		if got != c.want {
			t.Errorf("Failed! Wanted: %s Got: %s", c.want, got)
		}
	}
}
