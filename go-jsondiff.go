// A pure GO port of https://github.com/Simperium/jsondiff
package jsondiff

import(
	"github.com/sergi/go-diff/diffmatchpatch"
)

type Json map[string] interface{}

/*

Note to self for a little bit later when I get to actually dealing with the deltas... here's how to do it:

	dmp := diffmatchpatch.New()
	delta := dmp.DiffToDelta(dmp.DiffMain("blah blah", "blah foo blah", true))
	log.Printf("From: '%s' to '%s' :: %q", "blah blah", "blah foo blah", delta)
	diff, err := dmp.DiffFromDelta("blah blah", delta)
	if err != nil {
	log.Printf("error making diff from delta: %s", err.Error())
	}
	patch := dmp.PatchMake(diff)
	log.Printf("diff: %#v", diff)
	log.Printf("patch: %#v", patch)
	r, bools := dmp.PatchApply(patch, "blah blah")
	log.Printf("%#v, %s", bools, r)

The above code produces the following output:

	2014/01/31 15:14:01 From: 'blah blah' to 'blah foo blah' :: "=5\t+foo \t=4"
	2014/01/31 15:14:01 diff: []diffmatchpatch.Diff{diffmatchpatch.Diff{Type:0, Text:"blah "}, diffmatchpatch.Diff{Type:1, Text:"foo "}, diffmatchpatch.Diff{Type:0, Text:"blah"}}
	2014/01/31 15:14:01 patch: []diffmatchpatch.Patch{diffmatchpatch.Patch{diffs:[]diffmatchpatch.Diff{diffmatchpatch.Diff{Type:0, Text:"blah "}, diffmatchpatch.Diff{Type:1, Text:"foo "}, diffmatchpatch.Diff{Type:0, Text:"blah"}}, start1:0, start2:0, length1:9, length2:13}}
	2014/01/31 15:14:01 []bool{true}, blah foo blah
*/

type Diff struct {
	Operation string `json:"o"`
	Value string `json:"v"`
}

type DocumentChange struct {
	SourceRevision int `json:"sv"`
	ClientId string `json:"clientid"`
	Operation string `json:"o"`
	Changes map[string] Diff `json:"v"`
	Resutrevision int `json:"ev"`
	CurrentVersion string `json:"cv"`
	ChangesetIds []string `json:"ccds"`
}

type Changes []DocumentChange

type JsonDiff struct {
	dmp *diffmatchpatch.DiffMatchPatch
}

func (j *JsonDiff) Diff(from, to Json) Changes {
	changes := make([]DocumentChange,0)
	return changes
}

func (j *JsonDiff) Patch(from Json, diff Diff) Json {
	return Json{}
}

func New() *JsonDiff {
	j := new(JsonDiff)
	j.dmp = diffmatchpatch.New()
	return j
}
