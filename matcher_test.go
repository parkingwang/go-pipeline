package gopl

import "testing"

func TestDefaultURLMatcher_Match(t *testing.T) {
	pack := NewDataFrame()
	pack.setTopic("/gms/test/topic")

	matcher, err := NewDefaultURLMatcher("/gms/test/topic")
	if nil != err {
		t.Fatalf("Failed to create url matcher")
		panic(err)
	}

	if !matcher.Match(pack) {
		t.Fatalf("Not match")
	}

}

func TestDefaultURLMatcher_Match2(t *testing.T) {
	pack := NewDataFrame()
	pack.setTopic("gms://test.com/topic")

	matcher, err := NewDefaultURLMatcher("gms://test.com/topic")
	if nil != err {
		t.Fatalf("Failed to create url matcher")
		panic(err)
	}

	if !matcher.Match(pack) {
		t.Fatalf("Not match")
	}

}

func TestDefaultURLMatcher_MatchHeaders(t *testing.T) {
	pack := NewDataFrame()
	pack.setTopic("/gms/test/topic")
	pack.SetHeader("version", "2018")
	pack.SetDataFrameType("test.pack.type")

	matcher, err := NewDefaultURLMatcher("/gms/test/topic?version=2018&pack.type=test.pack.type")
	if nil != err {
		t.Fatalf("Failed to create url matcher")
		panic(err)
	}

	if !matcher.Match(pack) {
		t.Fatalf("Not match with headers")
	}

}

func BenchmarkNewDefaultURLMatcher(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewDefaultURLMatcher("/gms/test/topic?version=2018")
	}
}

func BenchmarkAnyMatcher_Match(b *testing.B) {
	pack := NewDataFrame()
	pack.setTopic("/gms/test/topic")
	pack.SetHeader("version", "2018")
	pack.SetDataFrameType("test.pack.type")

	matcher, err := NewDefaultURLMatcher("/gms/test/topic?version=2018&pack.type=test.pack.type")
	if nil != err {
		b.Fatalf("Failed to create url matcher")
		panic(err)
	}

	for i := 0; i < b.N; i++ {
		if !matcher.Match(pack) {
			b.Fatalf("Not match with headers")
		}
	}

}
