package panicmsg

import (
	"fmt"
	"reflect"

	. "github.com/onsi/gomega" // nolint
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

// PanicMsg is a wrapper for different types of panics.
func PanicMsg(expected interface{}) types.GomegaMatcher {
	switch e := expected.(type) {
	case string:
		return &Matcher{Expected: HavePrefix(e)}
	case types.GomegaMatcher:
		return &Matcher{Expected: e}
	default:
		panic("PanicMsg requires a string or matcher")
	}
}

// Matcher matches panic values against expectations
type Matcher struct {
	PanicValue interface{}
	Expected   types.GomegaMatcher
}

// Match handles the matching of panics
func (matcher *Matcher) Match(actual interface{}) (success bool, err error) {
	if actual == nil {
		return false, fmt.Errorf("Matcher expects a non-nil actual")
	}

	actualType := reflect.TypeOf(actual)
	if actualType.Kind() != reflect.Func {
		return false, fmt.Errorf("Matcher expects a function.  Got:\n%s", format.Object(actual, 1))
	}
	if !(actualType.NumIn() == 0 && actualType.NumOut() == 0) {
		return false, fmt.Errorf("Matcher expects a function with no arguments and no return value.  Got:\n%s", format.Object(actual, 1))
	}

	success = false
	defer func() {
		if e := recover(); e != nil {
			matcher.PanicValue = e
			switch me := e.(type) {
			case error:
				success, err = matcher.Expected.Match(me.Error())
			default:
				success, err = matcher.Expected.Match(e)
			}
		}
	}()

	reflect.ValueOf(actual).Call([]reflect.Value{})
	return
}

// FailureMessage is used when the match doesn't succeed.
func (matcher *Matcher) FailureMessage(actual interface{}) (message string) {
	if matcher.PanicValue == nil {
		return format.Message(actual, fmt.Sprintf("to panic with\n%s", format.Object(matcher.Expected, 1)))
	}
	if pv, ok := matcher.PanicValue.(error); ok {
		matcher.PanicValue = pv.Error()
	}
	return format.Message(actual, fmt.Sprintf("to panic with\n%s\ninstead of\n%s",
		format.Object(matcher.Expected, 1), format.Object(matcher.PanicValue, 1)))
}

// NegatedFailureMessage is used when an interse match doesn't succeed.
func (matcher *Matcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, fmt.Sprintf("not to panic, but panicked with\n%s", format.Object(matcher.PanicValue, 1)))
}
