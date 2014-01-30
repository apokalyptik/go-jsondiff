// A pure GO port of https://github.com/Simperium/jsondiff
package jsondiff

import(
	"github.com/sergi/go-diff/diffmatchpatch"
)

type Json map[string] interface{}

// Placeholder until I figure out what this'll look like...
type Diff string

type JsonDiff struct {
	dmp *diffmatchpatch.DiffMatchPatch
}

func (j *JsonDiff) Diff(from, to Json) Diff {
	return Diff("")
}

func (j *JsonDiff) Patch(from Json, diff Diff) Json {
	return Json{}
}

func New() *JsonDiff {
	j := new(JsonDiff)
	j.dmp = diffmatchpatch.New()
	return j
}
