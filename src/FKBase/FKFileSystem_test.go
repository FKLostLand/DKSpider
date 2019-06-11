package FKBase

import (
	"testing"
)

func TestIsFileExists(t *testing.T) {
	b, err := IsFileExists("C:\\Windows\\System32\\cmd.exe")
	if err != nil {
		t.Error(err)
	} else {
		if b {
			t.Log("cmd is exist")
		} else {
			t.Log("cmd is not exist")
		}
	}

	b, err = IsFileExists("C:\\Windows\\System32\\cmd2.exe")
	if err != nil {
		t.Error(err)
	} else {
		if b {
			t.Log("cmd2 is exist")
		} else {
			t.Log("cmd2 is not exist")
		}
	}
}

func TestWalkDir(t *testing.T) {
	d := WalkDir("C:\\Windows")
	t.Log("files and dirs number = ", len(d))

	d = WalkDir("C:\\Windows\\")
	t.Log("files and dirs number = ", len(d))

	var l []string
	l = append(l, "dll")
	d = WalkDir("C:\\Windows", l...)
	t.Log("dll number = ", len(d))

	l = append(l, "exe")
	d = WalkDir("C:\\Windows", l...)
	t.Log("exe + dll number = ", len(d))

	d = WalkDir("C:\\Windows\\System32")
	t.Log("files and dirs number = ", len(d))

	d = WalkDir("C:\\Windows\\System32\\de-DE")
	t.Log("files and dirs number = ", len(d))

	d = WalkDir("C:\\Windows\\System32\\de-DE\\Licenses")
	t.Log("files and dirs number = ", len(d))

	d = WalkDir("C:\\Windows\\System32\\de-DE\\Licenses\\eval")
	t.Log("files and dirs number = ", len(d))

	d = WalkDir("C:\\Windows\\System32\\de-DE\\Licenses\\eval\\Enterprise")
	t.Log("files and dirs number = ", len(d))
}
