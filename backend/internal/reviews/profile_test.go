package reviews

import "testing"

func TestGradeProfileRejectsExecutableEscapes(t *testing.T) {
	for _, source := range []string{
		"package arena; import \"os\"; func grade(){os.Exit(0)}",
		"package arena; func init(){}",
		"package arena; type X int; func (x X) Run(){}",
		"package arena\n//go:linkname forbidden\nfunc grade(){}",
		"package other; func grade(){}",
	} {
		if err := validateGradeProfile(source); err == nil {
			t.Fatalf("unsafe profile accepted: %s", source)
		}
	}
	if err := validateGradeProfile("package arena; import \"math/big\"; func grade(){}"); err != nil {
		t.Fatal(err)
	}
}
