package FKPinyin

import "testing"

func TestExamplePinyin_default(t *testing.T) {
	hans := "中国人"
	a := NewArgs()
	t.Log("default:", CreatePinyin(hans, a))
	// Output: default: [[zhong] [guo] [ren]]
}

func TestExamplePinyin_normal(t *testing.T) {
	hans := "中国人"
	a := NewArgs()
	a.Style = Normal
	t.Log("Normal:", CreatePinyin(hans, a))
	// Output: Normal: [[zhong] [guo] [ren]]
}

func TestExamplePinyin_tone(t *testing.T) {
	hans := "中国人"
	a := NewArgs()
	a.Style = Tone
	t.Log("Tone:", CreatePinyin(hans, a))
	// Output: Tone: [[zhōng] [guó] [rén]]
}

func TestExamplePinyin_tone2(t *testing.T) {
	hans := "中国人"
	a := NewArgs()
	a.Style = Tone2
	t.Log("Tone2:", CreatePinyin(hans, a))
	// Output: Tone2: [[zho1ng] [guo2] [re2n]]
}

func TestExamplePinyin_initials(t *testing.T) {
	hans := "中国人"
	a := NewArgs()
	a.Style = Initials
	t.Log("Initials:", CreatePinyin(hans, a))
	// Output: Initials: [[zh] [g] [r]]
}

func TestExamplePinyin_firstLetter(t *testing.T) {
	hans := "中国人"
	a := NewArgs()
	a.Style = FirstLetter
	t.Log(CreatePinyin(hans, a))
	// Output: [[z] [g] [r]]
}

func TestExamplePinyin_finals(t *testing.T) {
	hans := "中国人"
	a := NewArgs()
	a.Style = Finals
	t.Log(CreatePinyin(hans, a))
	// Output: [[ong] [uo] [en]]
}

func TestExamplePinyin_finalsTone(t *testing.T) {
	hans := "中国人"
	a := NewArgs()
	a.Style = FinalsTone
	t.Log(CreatePinyin(hans, a))
	// Output: [[ōng] [uó] [én]]
}

func TestExamplePinyin_finalsTone2(t *testing.T) {
	hans := "中国人"
	a := NewArgs()
	a.Style = FinalsTone2
	t.Log(CreatePinyin(hans, a))
	// Output: [[o1ng] [uo2] [e2n]]
}

func TestExamplePinyin_heteronym(t *testing.T) {
	hans := "中国人"
	a := NewArgs()
	a.Heteronym = true
	a.Style = Tone2
	t.Log(CreatePinyin(hans, a))
	// Output: [[zho1ng zho4ng] [guo2] [re2n]]
}

func TestExampleLazyPinyin(t *testing.T) {
	hans := "中国人"
	a := NewArgs()
	t.Log(CreateLazyPinyin(hans, a))
	// Output: [zhong guo ren]
}

func TestExampleSlug(t *testing.T) {
	hans := "中国人"
	a := NewArgs()
	t.Log(CreateSlugLazyPinyin(hans, a))
	// Output: zhong-guo-ren
}
