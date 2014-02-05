package jsondiff

import "testing"
import "io/ioutil"
import "encoding/json"
import "log"

type TestDiff struct {
	Name string
	From map[string]interface{}
	To   map[string]interface{}
	Diff string
}

var jd *JsonDiff = New()
var asserts []interface{} = make([]interface{}, 0)
var diffs map[string]TestDiff = make(map[string]TestDiff)

func TestSetup(t *testing.T) {
	data, err := ioutil.ReadFile("assertions.json")
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(data, &asserts)
	if err != nil {
		log.Fatal(err)
	}
	for _, vv := range asserts {
		v := vv.([]interface{})
		if len(v[2].([]interface{})) > 2 {
			continue // not handling otypes yet
		}
		t := v[0].(string)
		n := v[1].(string)
		switch t {
		case "diff":
			b, _ := json.Marshal(v[3].(map[string]interface{}))
			if string(b[:7]) == "{\"o\":\"O" {
				b[6] = []byte("M")[0]
			}
			diffs[n] = TestDiff{
				Name: n,
				From: v[2].([]interface{})[0].(map[string]interface{}),
				To:   v[2].([]interface{})[1].(map[string]interface{}),
				Diff: string(b)}
		}
	}
}

func TestAssertedDiffs(t *testing.T) {
	for _, v := range diffs {
		works, err := doesDiffApply(v.From, v.To)
		if err != nil {
			t.Errorf("Got error: %s", err.Error())
		}
		if false == works {
			t.Errorf("Failed Transform\nFrom: %#v\nTo: %#v\nDelta: %s\nGot: %#v", v.From, v.To, getDeltaString(v.From, v.To), getResult(v.From, v.To))
		}
		if v.Diff != getDeltaString(v.From, v.To) {
			t.Errorf("Failed Diff\nFrom: %#v\nTo: %#v\nExpected: %s\nGot: %s", v.From, v.To, v.Diff, getDeltaString(v.From, v.To))
		}
	}
}

func TestNilNil(t *testing.T) {
	d, e := jd.Diff(nil, nil)
	if d != nil || e != nil {
		t.Errorf("Expected (nil) (nil), got %#v, %#v", d, e)
	}
}

func TestSameSame(t *testing.T) {
	d, e := jd.Diff(
		map[string]interface{}{"foo": "bar"},
		map[string]interface{}{"foo": "bar"})
	if d != nil || e != nil {
		t.Errorf("Expected (nil) (nil), got %#v, %#v", d, e)
	}
}

func TestDelete(t *testing.T) {
	d, e := jd.Diff(
		map[string]interface{}{},
		nil)
	if e != nil {
		t.Errorf("Got error: %s", e.Error())
	}
	s, _ := d.String()
	if s != "{\"o\":\"-\"}" {
		t.Errorf("Got incorrect diff: '%s'", s)
	}
}

func TestCreate(t *testing.T) {
	d, e := jd.Diff(
		nil,
		map[string]interface{}{"foo": "bar"})
	if e != nil {
		t.Errorf("Got error: %s", e.Error())
	}
	s, _ := d.String()
	if s != "{\"o\":\"M\",\"v\":{\"foo\":{\"o\":\"+\",\"v\":\"bar\"}}}" {
		t.Errorf("Got incorrect diff: '%s'", s)
	}
}

func TestEmptyToFullDiff(t *testing.T) {
	from := map[string]interface{}{}
	to := map[string]interface{}{"foo": "bar"}
	works, err := doesDiffApply(from, to)
	if err != nil {
		t.Errorf("Got error: %s", err.Error())
	}
	if false == works {
		t.Errorf("From: %#v\n\nTo: %#v\n\nDelta: %s\n\nGot: %#v", from, to, getDeltaString(from, to), getResult(from, to))
	}
}

func TestFullToDifferentDiff(t *testing.T) {
	from := map[string]interface{}{"foo": "bar"}
	to := map[string]interface{}{"foo": "baz"}
	works, err := doesDiffApply(from, to)
	if err != nil {
		t.Errorf("Got error: %s", err.Error())
	}
	if false == works {
		t.Errorf("From: %#v\n\nTo: %#v\n\nDelta: %s\n\nGot: %#v", from, to, getDeltaString(from, to), getResult(from, to))
	}
}
