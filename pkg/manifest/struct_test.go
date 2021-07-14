package manifest_test

import (
	"bosh-compile/pkg/manifest"
	"reflect"
	"testing"
)

func TestCanCreateManifest(t *testing.T) {
	m := manifest.Manifest{
		Packages: []manifest.Package{
			{Name: "go-application", Dependencies: []string{"golang"}},
			{Name: "java-application", Dependencies: []string{"maven"}},
			{Name: "maven", Dependencies: []string{"java"}},
			{Name: "java", Dependencies: []string{}},
			{Name: "golang", Dependencies: []string{}},
		},
	}

	type test struct {
		node string
		want []string
	}

	tests := []test{
		{node: "go-application", want: []string{"golang", "go-application"}},
		{node: "java-application", want: []string{"java", "maven", "java-application"}},
		{node: "maven", want: []string{"java", "maven"}},
		{node: "java", want: []string{"java"}},
		{node: "golang", want: []string{"golang"}},
	}

	for _, tc := range tests {
		got, err := m.Dependencies(tc.node)
		if err != nil {
			t.Fail()
		}
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}
}

func TestDetermineTopLevelPackages(t *testing.T) {
	m := manifest.Manifest{
		Packages: []manifest.Package{
			{Name: "go-application", Dependencies: []string{"golang"}},
			{Name: "java-application", Dependencies: []string{"maven"}},
			{Name: "maven", Dependencies: []string{"java"}},
			{Name: "java", Dependencies: []string{}},
			{Name: "golang", Dependencies: []string{}},
		},
	}

	got, err := m.TopLevelPackages()
	if err != nil {
		t.Fail()
	}
	if len(got) != 3 {
		t.Fail()
	}
	expected := []string{"go-application", "java-application", "maven"}
	if !reflect.DeepEqual(expected, got) {
		t.Fatalf("expected: %v, got: %v", expected, got)
	}
}
