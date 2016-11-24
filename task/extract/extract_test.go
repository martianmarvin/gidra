package extract

import (
	"strings"
	"testing"
)

var testText = []byte(`
<div>
  <div class="t2">
    <p class="t3">
	  content
	</p>
  </div>
</div>
`)

func TestExtractElement(t *testing.T) {
	sel := `.t2>.t3`
	matcher, err := elementMatcher(sel)
	if err != nil {
		t.Fatal(err)
	}
	res := extractByElement(matcher, testText)
	if !strings.Contains(res[0], "content") {
		t.Fatal("Extracted " + res[0] + ", expecting 'content'")
	}
}

func TestExtractRegex(t *testing.T) {
	re := `(?s)t3">(.*?)<`
	matcher, err := regexMatcher(re)
	if err != nil {
		t.Fatal(err)
	}
	res := extractByRegex(matcher, testText)
	if !strings.Contains(res[0], "content") {
		t.Fatal("Extracted " + res[0] + ", expecting 'content'")
	}
}
