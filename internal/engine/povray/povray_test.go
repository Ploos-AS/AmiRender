package povray

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Ploos-AS/AmiRender/internal/engine"
)

func TestRenderUsesArgumentVectorWithoutShell(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "args.txt")
	fake := filepath.Join(dir, "povray")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"" + log + "\"\n"
	if err := os.WriteFile(fake, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}

	e := New(fake)
	scene := "scene;touch SHOULD_NOT_EXIST.pov"
	_, err := e.Render(context.Background(), engine.Job{
		ID: "m1-test", Scene: scene, Output: "frame.png", Width: 320, Height: 256,
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	for _, want := range []string{"+I" + scene, "+Oframe.png", "-D", "+W320", "+H256"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing argument %q in %q", want, got)
		}
	}
	if _, err := os.Stat("SHOULD_NOT_EXIST.pov"); !os.IsNotExist(err) {
		t.Fatal("scene name was interpreted by a shell")
	}
}
