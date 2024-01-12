package goerr_test

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/angel-one/goerr"
	"github.com/angel-one/goerr/samplesrc"
)

type testErrorType struct {
	err error
}

func (e *testErrorType) Error() string {
	return e.err.Error()
}

func TestBasic(t *testing.T) {
	err := samplesrc.Repository()
	if err == nil {
		t.Error("expecting an error")
	}

	want := "error from database"
	got := err.Error()

	if want != got {
		t.Errorf("Want: %s; Got: %s", want, got)
	}
}

func TestNestedErrors(t *testing.T) {
	err := samplesrc.Service()

	want := "service failed"
	got := err.Error()

	if want != got {
		t.Errorf("Want: %s; Got: %s", want, got)
	}
}

func TestStackDetails(t *testing.T) {
	err := samplesrc.Service()

	stacks := goerr.ListStacks(err)
	if len(stacks) != 2 {
		t.Errorf("Nr. of stack entries. Want: %d; Got: %d", 2, len(stacks))
	}

	first := stacks[0]
	if !strings.Contains(first, "service failed") {
		t.Errorf("stack do not contain right error")
	}
	if !strings.Contains(first, "/goerr/samplesrc/samples.go:20") {
		t.Errorf("stack do not contain right file/line number")
	}

	second := stacks[1]
	if !strings.Contains(second, "error from database") {
		t.Errorf("stack do not contain right error")
	}
	if !strings.Contains(second, "/goerr/samplesrc/samples.go:27") {
		t.Errorf("stack do not contain right file/line number")
	}
}

func TestStack(t *testing.T) {
	err := samplesrc.Controller()
	if err != nil {
		t.Logf("error in controller: %s", goerr.Stack(err))
	}

	stack := goerr.Stack(err)
	t.Log(stack)

	pattern := `controller failed \[.*/goerr/samplesrc/samples.go:12 \(samplesrc.Controller\)\]
\tservice failed \[.*/goerr/samplesrc/samples.go:20 \(samplesrc.Service\)\]
\t\terror from database.* \[.*/goerr/samplesrc/samples.go:27 \(samplesrc.Repository\)\]`
	match, _ := regexp.MatchString(pattern, stack)

	if !match {
		t.Errorf("stack is not matching the expectation")
	}
}

func TestStackNonGoErr(t *testing.T) {
	err := errors.New("some sample error")

	want := "some sample error"
	got := goerr.Stack(err)

	if want != got {
		t.Errorf("Want: %s, Got: %s", want, got)
	}
}

func TestStackWithNil(t *testing.T) {
	want := ""
	got := goerr.Stack(nil)

	if want != got {
		t.Errorf("Want: %v, Got: %v", want, got)
	}
}

func TestWithHTTPCode(t *testing.T) {
	repository := func() error {
		return goerr.New(errors.New("db key error"), http.StatusConflict, "repository error")
	}
	service := func() error {
		err := repository()
		if err != nil {
			return goerr.New(err, "service error")
		}
		return nil
	}
	controller := func() error {
		err := service()
		if err != nil {
			return goerr.New(err, "controller error")
		}
		return nil
	}

	want := http.StatusConflict
	got := goerr.Code(controller())

	if want != got {
		t.Errorf("Want: %d, Got: %d", want, got)
	}
}

func TestWithHTTPCodeChangedInMiddle(t *testing.T) {
	repository := func() error {
		return goerr.New(errors.New("db key error"), http.StatusConflict, "repository error")
	}
	service := func() error {
		err := repository()
		if err != nil {
			return goerr.New(err, http.StatusBadRequest, "service error")
		}
		return nil
	}
	controller := func() error {
		err := service()
		if err != nil {
			return goerr.New(err, "controller error")
		}
		return nil
	}

	want := http.StatusBadRequest
	got := goerr.Code(controller())

	if want != got {
		t.Errorf("Want: %d, Got: %d", want, got)
	}
}

func TestWithHTTPCodeStack(t *testing.T) {
	repository := func() error {
		return goerr.New(errors.New("db key error"), http.StatusConflict, "repository error")
	}
	service := func() error {
		err := repository()
		if err != nil {
			return goerr.New(err, "service error")
		}
		return nil
	}
	controller := func() error {
		err := service()
		if err != nil {
			return goerr.New(err, "controller error")
		}
		return nil
	}

	err := controller()

	got := goerr.Stack(err)
	if !strings.Contains(got, "repository error (409)") {
		t.Errorf("stack do not contain the error code. %s", got)
	}

}

func Test_Unwrap(t *testing.T) {
	testErr1 := &testErrorType{errors.New("foo err")}
	testErr2 := &testErrorType{errors.New("bar err")}

	t.Run("Unwrap should return correct underlying error", func(t *testing.T) {
		err := goerr.New(testErr1, "layer 1 failed")

		if unwrapped := errors.Unwrap(err); unwrapped != testErr1 {
			t.Fatalf("invalid unwrapped error returned. expected = %+v, got = %+v", testErr1, err)
		}
	})

	t.Run("wrapped error should support errors.Is across nestings", func(t *testing.T) {
		err := goerr.New(testErr1, "layer 1 failed")
		err = goerr.New(err, "layer 2 failed")

		if !errors.Is(err, testErr1) {
			t.Fatalf("expected errors.Is(testErr1, err) to return true, returned false")
		}

		if errors.Is(err, testErr2) {
			t.Fatalf("expected errors.Is(err, testErr2) to return false, returned true")
		}
	})

	t.Run("wrapped error should support errors.As across nestings", func(t *testing.T) {
		err := goerr.New(testErr1, "layer 1 failed")
		err = goerr.New(err, "layer 2 failed")

		target := &testErrorType{}

		if !errors.As(err, &target) {
			t.Fatalf("expected !errors.As(err, &target) to return true, returned false")
		}

		if target.err != testErr1.err {
			t.Fatalf("found target.err != testErr1.err")
		}
	})
}
