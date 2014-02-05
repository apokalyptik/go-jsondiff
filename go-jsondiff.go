// A pure GO port of https://github.com/Simperium/jsondiff
package jsondiff

import(
	"bytes"
	"sort"
	"strconv"
	"strings"
	"log"
	"github.com/sergi/go-diff/diffmatchpatch"
	"encoding/json"
	"reflect"
)

// A slice of document changes. Returned by Parse()
type Changes []DocumentChange

// This type is just here to save on placing "map[string] interface{}" all over the codebase.
type Json map[string] interface{}

// Parses a string list of jsondiff changes into an array of changes.  These changes are
// generated by Simperium as objects in a bucket for a user are modified by other connected
// clients. The string from Simperium looke like: %d:c:[...] where %d is the channel that
// You initialized when authorizing against the bucket.
func Parse(changes string) (Changes, error) {
	js := make(Changes,0)
	if err := json.Unmarshal([]byte(changes), &js); err != nil {
		return nil, err
	}
	return js, nil
}

// This struct represents an diff for a document. Sometimes this is a single operation and
// other times more diffs are recursively nested down inside. This only applies to keys 
// inside the main document dict. For top level changes see DocumentChange{}.
type Diff struct {
	Operation string `json:"o,omitempty"` // +, -, r, I, d, L, dL, O
	Value interface{} `json:"v,omitempty"`
}

func patchString(from, delta string) (string, error) {
	dmp := diffmatchpatch.New()
	diff, err := dmp.DiffFromDelta(from, delta)
	if err != nil {
		return from, err
	}
	rval, _ := dmp.PatchApply(dmp.PatchMake(diff), from) // TODO: error handling
	return rval, nil
}

// Apply a diff to a part of a document. This only applies to keys inside the main document dict. 
// Never to the entire document. For top level changes see DocumentChange. interface{} is used
// here because we could be looking at anything (bool, string, int, float, list, or dict)
func (d *Diff) apply(data interface{}) interface{} {
	// At this point data could be a number of things
	switch d.Operation {
		case "-": // Delete value at index
			return nil
		case "+": // Insert new value at index
			return d.Value
		case "r": // Replace value at index
			return d.Value
		case "I": // Integer, add the difference to current value
			switch data.(type) {
				case int, int8, int16, int32, int64:
					return (data.(int64)+d.Value.(int64))
				case uint, uint8, uint16, uint32, uint64:
					return (data.(uint64)+d.Value.(uint64))
				case float32, float64:
					return (data.(float64)+d.Value.(float64))
				case nil:
					return d.Value
			}
		case "d": // DiffMatchPatch string at index
			rval, _ := patchString(data.(string), d.Value.(string)) // TODO: error handling
			return rval
		case "dL": // List, apply the diff operations to the current array (w/dmp) (?)
			l := len(data.([]interface{}))
			list := make([]string, l)
			for k, v := range data.([]interface{}) {
				buf, _ := json.Marshal(v) // TODO: error handling
				list[k] = string(buf)
			}
			newList, _ := patchString(strings.Join(list, "\n") + "\n", d.Value.(string))
			newL := make([]interface{}, 0)
			for _, v := range strings.Split(newList, "\n") {
				if len(v) < 1 {
					continue
				}
				i := new(interface{})
				json.Unmarshal([]byte(v), i) // TODO: error handling
				newL = append(newL, *i)
			}
			return newL
		case "L":  // [recurse] List, apply the diff operations to the current array
			// TODO: This is relatively expensive (especially the copying/extending of slices)
		    list := data.([]interface{})
			newList := make([]interface{},len(list))
			for k, v := range list {
				newList[k] = v
			}
			changes := make(map[int]interface{})
			changeList := make([]int, 0);
			deleted := 0
			for key, value := range d.Value.(map[string]interface{}) {
				iKey, _ := strconv.Atoi(key)
				changeList = append(changeList, iKey)
				changes[iKey] = value
			}
			sort.Ints(changeList)
			for _, iKey := range changeList {
				vm := changes[iKey].(map[string]interface{})
				diff := new(Diff)
				diff.Operation = vm["o"].(string)
				if v, ok := vm["v"]; ok == true {
					diff.Value = v
				}
				affect := iKey - deleted
				if len(newList) <= affect {
					newNewList := make([]interface{}, iKey+1)
					for k, v := range newList {
						newNewList[k] = v
					}
					newList = newNewList
				}
				if newValue := diff.apply(newList[affect]); newValue != nil {
					newList[affect] = newValue
				} else {
					deleted++
					newList = append(newList[:affect], newList[affect+1:]...)
				}
			}
			return newList
			// Buggy, requires a schema option to allow simperium to return this op. Unimplimented for now
		case "O":  // [recurse] Object, apply the diff operations to the current object
			doc := data.(map[string]interface {})
			for k, v := range d.Value.(map[string]interface{}) {
				diff := new(Diff)
				diff.Operation = v.(map[string]interface{})["o"].(string)
				diff.Value = v.(map[string]interface{})["v"]
				if part, ok := doc[k]; true == ok {
					if newpart := diff.apply(part); nil != newpart {
						doc[k] = newpart
					} else {
						delete(doc, k)
					}
				} else {
					if newpart := diff.apply(nil); nil != newpart {
						doc[k] = newpart
					} else {
						// non operation delete already missing key from a dict
					}
				}
			}
			return doc
	}
	return data
}

// This represents all of the changes to a single document that are occuring at once. This data
// (excepting Changes which are processed on their own) never deals directly with individual
// parts of the document but rather the document as a whole.
type DocumentChange struct {
	Document string `json:"id,omitempty"`
	SourceRevision int `json:"sv,omitempty"`
	ClientId string `json:"clientid,omitempty"`
	Operation string `json:"o,omitempty"` // M, -
	Changes map[string] Diff `json:"v,omitempty"`
	Resultrevision int `json:"ev,omitempty"`
	CurrentVersion string `json:"cv,omitempty"`
	ChangesetIds []string `json:"ccids,omitempty"`
	ChangesetId string `json:"ccid,omitempty"`
}

func (d *DocumentChange) String() (string, error) {
	if b, e := json.Marshal(*d); e != nil {
		return "", e
	} else {
		return string(b), nil
	}
}

// This only applies to the *entire* document object. Never to keys inside the document dict
func (d *DocumentChange) apply(doc Json) Json {
	// { {dict-key}: { "o": {operation}, "v": {value} }
	switch d.Operation {
		case "-": // Delete Document
			return nil
		case "M": // Modify/Create Document
			newDocument := make(Json)
			for k, v := range doc {
				newDocument[k] = v
			}
			for k, change := range d.Changes {
				if v, ok := newDocument[k]; ok == true {
					if r := change.apply(v); r != nil {
						newDocument[k] = r
					} else {
						delete(newDocument, k)
					}
				} else {
					if r := change.apply(nil); r != nil {
						newDocument[k] = r
					} else {
						// non op, delete k, no k. possibly inconsistent?
					}
				}
			}
			return newDocument
		default:
			log.Fatal("jsondiff.DocumentChange.Operation was of unknown value")
			return doc
	}
}

// The main entrypoint for changing a document dict via a documentchange. I suspect that I'll
// reture this and just export the apply function on DocumentChange{}. The main reason this 
// is here is to keep a copy of DiffMatchPatch all to ourselves
type JsonDiff struct {
	dmp *diffmatchpatch.DiffMatchPatch
}

// Apply a DocumentChange{} to a document dict
func (j *JsonDiff) Apply(document Json, change DocumentChange) (Json, error) {
	newDocument := change.apply(document)
	return newDocument, nil
}

func (j *JsonDiff) diff_obj(from interface{}, to Json) *Diff {
	return &Diff{}
}

func (j *JsonDiff) diff_list(from interface{}, to []interface{}) *Diff {
	return &Diff{}
}

func (j *JsonDiff) diff_string(from interface{}, to string) *Diff {
	switch from.(type) {
		case string:
			if from.(string) != to {
				return &Diff{ Operation: "d", Value: j.dmp.DiffToDelta(j.dmp.DiffMain(from.(string), to, true)) }
			}
			return nil
		default:
			return j.diff_replace(from, to)
	}
}

func (j *JsonDiff) diff_replace(from interface{}, to interface{}) *Diff {
	return &Diff{Operation: "r", Value: to}
}

func (j *JsonDiff) diff(from interface{}, to interface{}) *Diff {
	if reflect.DeepEqual(from, to) {
		return nil
	}

	switch from.(type) {
		case nil:
			return &Diff{ Operation: "+", Value: to }
	}

	switch to.(type) {
		case nil:
			return &Diff{ Operation: "-" }
	}

	fKind := reflect.TypeOf(from).Kind().String()
	tKind := reflect.TypeOf(to).Kind().String()

	if fKind != tKind {
		return j.diff_replace(from, to)
	}

	switch to.(type) {
		case map[string] interface{}:
			return j.diff_obj(from, to.(map[string] interface{}))
		case []interface{}:
			return j.diff_list(from, to.([]interface{}))
		case string:
			return j.diff_string(from, to.(string))
		default:
			return j.diff_replace(from, to)
	}
}

func(j *JsonDiff) Diff(from, to Json) (*DocumentChange, error) {
	// nothing to nothing
	if from == nil && to == nil {
		return nil, nil
	}

	// nothing to something
	if from == nil && to != nil {
		rval := new(DocumentChange)
		rval.Changes =make(map[string] Diff)
		for k, v := range to {
			rval.Changes[k] = Diff{ Operation: "+", Value: v }
		}
		rval.Operation = "M"
		return rval, nil
	}

	// something to nothing
	if from != nil && to == nil {
		rval := new(DocumentChange)
		rval.Operation = "-"
		return rval, nil
	}

	if equal, err := j.equal(to, from); true == equal {
		return nil, nil
	} else {
		if err != nil {
			return nil, err
		}
	}

	rval := new(DocumentChange)
	rval.Operation = "M"
	rval.Changes = make(map[string] Diff)
	for k, tv := range to {
		if fv, ok := from[k]; ok {
			if d := j.diff(fv, tv); d != nil {
				rval.Changes[k] = *d
			}
		} else {
			if d := j.diff(fv, tv); d != nil {
				rval.Changes[k] = *d
			}
		}
	}
	for k, fv := range from {
		if _, ok := to[k]; ok {
			continue
		}
		if d := j.diff(fv, nil); d != nil {
			rval.Changes[k] = *d
		}
	}
	return rval, nil
}

func (j *JsonDiff) equal(from, to interface{}) (bool, error) {
	b1, e1 := json.Marshal(from)
	b2, e2 := json.Marshal(to)
	if e1 != nil { return false, e1 }
	if e2 != nil { return false, e2 }
	if 0 == bytes.Compare(b1, b2) {
		return true, nil
	}
	return false, nil
}

// Get a new JsonDiff
func New() *JsonDiff {
	j := new(JsonDiff)
	j.dmp = diffmatchpatch.New()
	return j
}

// Test helper
func getDiff(from, to map[string]interface{}) (*DocumentChange, error) {
	jd := New()
	return jd.Diff(from, to)
}

// Test helper
func getDelta(from, to map[string]interface{}) (string, error) {
	diff, e := getDiff(from, to)
	if e != nil {
		return "", e
	}
	return diff.String()
}

// Test helper
func getDeltaString(from, to map[string]interface{}) string {
	s, _ := getDelta(from, to)
	return s
}

// Test Helper
func getResult(from, to map[string]interface{}) map[string]interface{} {
	jd := New()
	diff, _ := getDiff(from, to)
	res, _ := jd.Apply(from, *diff)
	return res
}

// Test helper
func doesDiffApply(from, to map[string]interface{}) (bool, error) {
	jd := New()
	diff, e := getDiff(from, to)
	res, e := jd.Apply(from, *diff)
	if e != nil {
		return false, e
	}
	if equal, e := jd.equal(to, res); e != nil {
		return false, e
	} else {
		if false == equal {
			return false, nil
		}
		return true, nil
	}
}
