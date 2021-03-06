package elastic

import (
	"encoding/json"
	"testing"
)

func TestRangeQuery(t *testing.T) {
	q := NewRangeQuery("postDate").From("2010-03-01").To("2010-04-01")
	q = q.QueryName("my_query")
	data, err := json.Marshal(q.Source())
	if err != nil {
		t.Fatalf("marshaling to JSON failed: %v", err)
	}
	got := string(data)
	expected := `{"range":{"_name":"my_query","postDate":{"from":"2010-03-01","include_lower":true,"include_upper":true,"to":"2010-04-01"}}}`
	if got != expected {
		t.Errorf("expected\n%s\n,got:\n%s", expected, got)
	}
}

/*
func TestRangeQueryGte(t *testing.T) {
	q := NewRangeQuery("postDate").Gte("2010-03-01")
	data, err := json.Marshal(q.Source())
	if err != nil {
		t.Fatalf("marshaling to JSON failed: %v", err)
	}
	got := string(data)
	expected := `{"range":{"postDate":{"gte":"2010-03-01"}}}`
	if got != expected {
		t.Errorf("expected\n%s\n,got:\n%s", expected, got)
	}
}
*/
