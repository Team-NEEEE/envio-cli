package workspace

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type gitCall struct {
	cwd  string
	args []string
}

type fakeGitRunner struct {
	outputs []string
	errs    []error
	calls   []gitCall
}

func (f *fakeGitRunner) Run(_ context.Context, cwd string, args ...string) (string, error) {
	f.calls = append(f.calls, gitCall{cwd: cwd, args: append([]string(nil), args...)})
	index := len(f.calls) - 1
	var output string
	if index < len(f.outputs) {
		output = f.outputs[index]
	}
	var err error
	if index < len(f.errs) {
		err = f.errs[index]
	}
	return output, err
}

func TestGitInspectorInspectReturnsRootAndOrigin(t *testing.T) {
	t.Parallel()

	runner := &fakeGitRunner{
		outputs: []string{"C:/work/envio-cli\n", "git@github.com:Team-NEEEE/envio-cli.git\n"},
	}
	inspector := NewGitInspector(runner)

	got, err := inspector.Inspect(context.Background(), "C:/work/envio-cli/internal")
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if got.Root != "C:/work/envio-cli" || got.OriginURL != "git@github.com:Team-NEEEE/envio-cli.git" {
		t.Fatalf("Inspect() = %#v", got)
	}

	wantCalls := []gitCall{
		{cwd: "C:/work/envio-cli/internal", args: []string{"rev-parse", "--show-toplevel"}},
		{cwd: "C:/work/envio-cli", args: []string{"remote", "get-url", "origin"}},
	}
	if !reflect.DeepEqual(runner.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, wantCalls)
	}
}

func TestGitInspectorMapsMissingRepositoryToRequiredError(t *testing.T) {
	t.Parallel()

	runner := &fakeGitRunner{
		errs: []error{errors.New("not a git repository")},
	}
	inspector := NewGitInspector(runner)

	_, err := inspector.Inspect(context.Background(), "C:/work/not-git")
	if !errors.Is(err, ErrGitRepositoryRequired) {
		t.Fatalf("Inspect() error = %v, want ErrGitRepositoryRequired", err)
	}
	if len(runner.calls) != 1 {
		t.Fatalf("calls = %#v, want one rev-parse call", runner.calls)
	}
}

func TestGitInspectorMapsMissingOriginToRequiredError(t *testing.T) {
	t.Parallel()

	runner := &fakeGitRunner{
		outputs: []string{"C:/work/envio-cli\n"},
		errs:    []error{nil, errors.New("No such remote 'origin'")},
	}
	inspector := NewGitInspector(runner)

	_, err := inspector.Inspect(context.Background(), "C:/work/envio-cli")
	if !errors.Is(err, ErrGitRepositoryRequired) {
		t.Fatalf("Inspect() error = %v, want ErrGitRepositoryRequired", err)
	}
	if len(runner.calls) != 2 {
		t.Fatalf("calls = %#v, want rev-parse and remote lookup", runner.calls)
	}
}
