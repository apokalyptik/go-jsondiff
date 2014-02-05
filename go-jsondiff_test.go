package jsondiff

import "testing"

func rfirst(i... interface{}) interface{} {
	return i[0]
}

func TestNilNil(t *testing.T) {
	jd := New()
	d, e := jd.Diff(nil,nil)
	if d != nil || e != nil {
		t.Errorf("Expected (nil) (nil), got %#v, %#v", d, e)
	}
}

func TestSameSame(t *testing.T) {
	jd := New()
	d, e := jd.Diff(
		map[string]interface{}{ "foo": "bar" },
		map[string]interface{}{ "foo": "bar" })
	if d != nil || e != nil {
		t.Errorf("Expected (nil) (nil), got %#v, %#v", d, e)
	}
}

func TestDelete(t *testing.T) {
	jd := New()
	d, e := jd.Diff(
		map[string]interface{}{},
		nil)
	if e != nil { t.Errorf("Got error: %s", e.Error()) }
	s, _ := d.String()
	if s != "{\"o\":\"-\"}" {
		t.Errorf("Got incorrect diff: '%s'", s)
	}
}

func TestCreate(t *testing.T) {
	jd := New()
	d, e := jd.Diff(
		nil,
		map[string]interface{}{ "foo": "bar" })
	if e != nil { t.Errorf("Got error: %s", e.Error()) }
	s, _ := d.String()
	if s != "{\"o\":\"M\",\"v\":{\"foo\":{\"o\":\"+\",\"v\":\"bar\"}}}" {
		t.Errorf("Got incorrect diff: '%s'", s)
	}
}

func TestEmptyToFullDiff(t *testing.T) {
	from := map[string]interface{}{}
	to := map[string]interface{}{ "foo": "bar" }
	works, err := doesDiffApply(from, to)
	if err != nil { t.Errorf("Got error: %s", err.Error()) }
	if false == works {
		t.Errorf("From: %#v\n\nTo: %#v\n\nDelta: %s\n\nGot: %#v", from, to, getDeltaString(from, to), getResult(from, to))
	}
}

func TestFullToDifferentDiff(t *testing.T) {
	from := map[string]interface{}{ "foo": "bar" }
	to := map[string]interface{}{ "foo": "baz" }
	works, err := doesDiffApply(from, to)
	if err != nil { t.Errorf("Got error: %s", err.Error()) }
	if false == works {
		t.Errorf("From: %#v\n\nTo: %#v\n\nDelta: %s\n\nGot: %#v", from, to, getDeltaString(from, to), getResult(from, to))
	}
}


