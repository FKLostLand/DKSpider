package FKLog

import (
	"FKBase"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestFile(t *testing.T) {
	log := CreateDefaultLogger(10000)
	log.SetLogger("file", map[string]interface{}{"filename": "test.log"})

	log.Debug("debug")
	log.Informational("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("error")
	log.Alert("alert")
	log.Critical("critical")
	log.Emergency("emergency")

	time.Sleep(time.Second * 2)

	f, err := os.Open("test.log")
	if err != nil {
		t.Fatal(err)
	}
	b := bufio.NewReader(f)
	linenum := 0
	for {
		line, _, err := b.ReadLine()
		if err != nil {
			break
		}
		if len(line) > 0 {
			linenum++
		}
	}
	var expected = FKBase.LevelDebug
	if linenum != expected {
		t.Fatal(linenum, "not "+strconv.Itoa(expected)+" lines")
	}
	os.Remove("test.log")
}

func TestFileRotate(t *testing.T) {
	log := CreateDefaultLogger(10000)
	log.SetLogger("file", map[string]interface{}{"filename": "testRotate.log", "maxlines": 4})
	log.Debug("debug")
	log.Informational("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("error")
	log.Alert("alert")
	log.Critical("critical")
	log.Emergency("emergency")

	time.Sleep(time.Second * 2)

	rotatename := "testRotate.log" + fmt.Sprintf(".%s.%03d", time.Now().Format("2006-01-02"), 1)
	b, err := FKBase.IsFileExists(rotatename)
	if !b || err != nil {
		t.Fatal("rotate not generated: " + rotatename)
	}
	os.Remove(rotatename)
	time.Sleep(time.Second * 1)
	os.Remove("testRotate.log")
}

func BenchmarkFile(b *testing.B) {
	log := CreateDefaultLogger(100000)
	log.SetLogger("file", map[string]interface{}{"filename": "benchmark.log"})
	for i := 0; i < b.N; i++ {
		log.Debug("debug")
	}
	os.Remove("benchmark.log")
}
