package errors

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Errors Suite")
}

var (
	ErrNotFound            = New("not found")
	ErrNotFoundWithOptions = &Error{
		Title: "not found",
		opts: opts{
			Identifier: []uint32{1},
			Details:    []string{"File not found in the session"},
			Properties: map[string]any{
				"File": "test.txt",
				"User": "john.doe",
			},
		},
	}
	ErrForbidden = New("forbidden")
	ErrInternal  = New("internal error")

	errTest = errors.New("test error")
	errPerm = errors.New("permission denied")

	e, e1, e2 error
)

var _ = Describe("Errors", func() {
	Context("When creating an error with New()", func() {
		It("should return an Error with the given title and default options", func() {
			e := New("my error")
			var err *Error
			Expect(errors.As(e, &err)).To(BeTrue())
			Expect(err.Title).To(Equal("my error"))
			Expect(err.Identifier).To(BeNil())
			Expect(err.Details).To(BeEmpty())
			Expect(err.Properties).To(BeEmpty())
			Expect(err.Cause).To(BeNil())
		})
	})

	Context("When creating a new error from a standard one", func() {
		It("should return an Unknown Error error, when no options are provided", func() {
			e = Wrap(errTest)
			var err *Error
			if ok := errors.As(e, &err); ok {
				Expect(err.Title).To(Equal("unknown error"))
				Expect(err.Identifier).To(BeZero())
				Expect(err.Details).To(BeEmpty())
				Expect(err.Properties).To(BeEmpty())
				Expect(err.Cause).To(Equal(&causeError{error: errTest}))
				Expect(err.stack[0].File).To(ContainSubstring("errors_test.go"))
				Expect(err.stack[0].Line).To(BeNumerically(">", 0))
				Expect(err.stack[0].Function).NotTo(BeEmpty())
			}
		})
		It("should support WithDetailf for formatted details", func() {
			e = Wrap(errTest,
				WithDetailf("failed at %s:%d", "step", 42),
			)
			var err *Error
			Expect(errors.As(e, &err)).To(BeTrue())
			Expect(err.Details).To(Equal([]string{"failed at step:42"}))
		})
		It("should return the error with the correct identifier, details, properties, when options are provided", func() {
			e = Wrap(errTest,
				WithIdentifier(1001),
				WithDetail("This is a test error"),
				WithProperty("type", "fake"),
			)
			var err *Error
			if ok := errors.As(e, &err); ok {
				Expect(err.Title).To(Equal("unknown error"))
				Expect(err.Identifier).To(Equal([]uint32{1001}))
				Expect(err.Details).To(Equal([]string{"This is a test error"}))
				Expect(err.Properties).To(Equal(map[string]any{"type": "fake"}))
				Expect(err.Cause).To(Equal(&causeError{error: errTest}))
				Expect(err.stack[0].File).To(ContainSubstring("errors_test.go"))
				Expect(err.stack[0].Line).To(BeNumerically(">", 0))
				Expect(err.stack[0].Function).NotTo(BeEmpty())
			}
		})
	})

	Context("When creating a new error from a custom one", func() {
		It("should return the custom error, when no options are provided", func() {
			e = Wrap(ErrNotFound)
			var err *Error
			if ok := errors.As(e, &err); ok {
				Expect(err.Title).To(Equal("not found"))
				Expect(err.Identifier).To(BeZero())
				Expect(err.Details).To(BeEmpty())
				Expect(err.Properties).To(BeEmpty())
				Expect(err.Cause).To(BeNil())
				Expect(err.stack[0].File).To(ContainSubstring("errors_test.go"))
				Expect(err.stack[0].Line).To(BeNumerically(">", 0))
				Expect(err.stack[0].Function).NotTo(BeEmpty())
			}
		})
		It("should return the error with the correct identifier, details, properties, when error is enriched", func() {
			e = Wrap(ErrNotFound,
				WithDetail("File not found in the session"),
				WithProperty("File", "test.txt"),
				WithIdentifier(1001),
			)
			var err *Error
			if ok := errors.As(e, &err); ok {
				Expect(err.Title).To(Equal("not found"))
				Expect(err.Identifier).To(Equal([]uint32{1001}))
				Expect(err.Details).To(Equal([]string{"File not found in the session"}))
				Expect(err.Properties).To(Equal(map[string]any{"File": "test.txt"}))
				Expect(err.Cause).To(BeNil())
				Expect(err.stack[0].File).To(ContainSubstring("errors_test.go"))
			}
		})
	})

	Context("When creating a new error from a custom one with options", func() {
		It("should return the custom error with same options", func() {
			e = Wrap(ErrNotFoundWithOptions)
			var err *Error
			if ok := errors.As(e, &err); ok {
				Expect(err.Title).To(Equal("not found"))
				Expect(err.Identifier).To(Equal([]uint32{1}))
				Expect(err.Details).To(Equal([]string{"File not found in the session"}))
				Expect(err.Properties).To(Equal(map[string]any{"File": "test.txt", "User": "john.doe"}))
				Expect(err.Cause).To(BeNil())
				Expect(err.stack[0].File).To(ContainSubstring("errors_test.go"))
				Expect(err.stack[0].Line).To(BeNumerically(">", 0))
				Expect(err.stack[0].Function).NotTo(BeEmpty())
			}
		})
		It("should return the error with the correct identifier, details, properties, when error is enriched", func() {
			e = Wrap(ErrNotFoundWithOptions,
				WithIdentifier(1001),
				WithDetail("custom client role is 'Reader'"),
				WithProperty("ClientID", "1234567890"),
			)
			var err *Error
			if ok := errors.As(e, &err); ok {
				Expect(err.Title).To(Equal("not found"))
				Expect(err.Identifier).To(Equal([]uint32{1, 1001}))
				Expect(err.Details).To(Equal([]string{"File not found in the session", "custom client role is 'Reader'"}))
				Expect(err.Properties).To(Equal(map[string]any{"File": "test.txt", "User": "john.doe", "ClientID": "1234567890"}))
				Expect(err.Cause).To(BeNil())
				Expect(err.stack[0].File).To(ContainSubstring("errors_test.go"))
				Expect(err.stack[0].Line).To(BeNumerically(">", 0))
				Expect(err.stack[0].Function).NotTo(BeEmpty())
			}
		})

		It("should return the error without adding an empty property", func() {
			e = Wrap(ErrNotFoundWithOptions,
				WithIdentifier(1001),
				WithDetail("custom client role is 'Reader'"),
				WithProperty("", "test"),
			)
			var err *Error
			if ok := errors.As(e, &err); ok {
				Expect(err.Title).To(Equal("not found"))
				Expect(err.Identifier).To(Equal([]uint32{1, 1001}))
				Expect(err.Details).To(Equal([]string{"File not found in the session", "custom client role is 'Reader'"}))
				Expect(err.Properties).To(Equal(map[string]any{"File": "test.txt", "User": "john.doe"}))
				Expect(err.Cause).To(BeNil())
				Expect(err.stack[0].File).To(ContainSubstring("errors_test.go"))
				Expect(err.stack[0].Line).To(BeNumerically(">", 0))
				Expect(err.stack[0].Function).NotTo(BeEmpty())
			}
		})
	})

	Context("When creating a new error from a custom one caused by a standard error", func() {
		It("should return the custom error", func() {
			e = Wrap(ErrForbidden,
				WithIdentifier(403001),
				WithDetail("missing role 'admin' for the user"),
				WithProperty("User", "john.doe"),
				CausedBy(errPerm),
			)
			var err *Error
			if ok := errors.As(e, &err); ok {
				Expect(err.Title).To(Equal("forbidden"))
				Expect(err.Identifier).To(Equal([]uint32{403001}))
				Expect(err.Details).To(Equal([]string{"missing role 'admin' for the user"}))
				Expect(err.Properties).To(Equal(map[string]any{"User": "john.doe"}))
				Expect(err.Cause).To(Equal(&causeError{error: errPerm}))
				Expect(len(err.stack)).To(Equal(1))
				Expect(err.stack[0].File).To(ContainSubstring("errors_test.go"))
				Expect(err.stack[0].Line).To(BeNumerically(">", 0))
				Expect(err.stack[0].Function).NotTo(BeEmpty())
			}
		})
	})

	Context("When wrapping an error", func() {
		It("should return the wrapped error, when not enriched", func() {
			e1 = Wrap(ErrInternal,
				WithIdentifier(500001),
				WithDetail("unexpected error occurred while processing the request"),
				WithProperty("RequestID", "1234567890"),
				CausedBy(errPerm),
			)
			e = Wrap(e1)
			var err *Error
			if ok := errors.As(e, &err); ok {
				Expect(err.Title).To(Equal("internal error"))
				Expect(err.Identifier).To(Equal([]uint32{500001}))
				Expect(err.Details).To(Equal([]string{"unexpected error occurred while processing the request"}))
				Expect(err.Properties).To(Equal(map[string]any{"RequestID": "1234567890"}))
				Expect(err.Cause).To(Equal(&causeError{error: errPerm}))
				Expect(len(err.stack)).To(Equal(2))
				Expect(err.stack[0].File).To(ContainSubstring("errors_test.go"))
				Expect(err.stack[0].Line).To(BeNumerically(">", 0))
				Expect(err.stack[0].Function).NotTo(BeEmpty())
				Expect(err.stack[1].File).To(ContainSubstring("errors_test.go"))
				Expect(err.stack[1].Line).To(BeNumerically(">", 0))
				Expect(err.stack[1].Function).NotTo(BeEmpty())
			}
		})
		It("should return the wrapped error with additional information, when enriched", func() {
			e1 = Wrap(ErrInternal,
				WithIdentifier(500002),
				WithDetail("database connection failed"),
				WithProperty("RequestID", "1234567890"),
				CausedBy(errPerm),
			)

			e = Wrap(e1,
				WithDetail("wrong url"),
				WithProperty("url", "https://bdd.fake.com"),
			)
			var err *Error
			if ok := errors.As(e, &err); ok {
				Expect(err.Title).To(Equal("internal error"))
				Expect(err.Identifier).To(Equal([]uint32{500002}))
				Expect(err.Details).To(Equal([]string{"database connection failed", "wrong url"}))
				Expect(err.Properties).To(Equal(map[string]any{"RequestID": "1234567890", "url": "https://bdd.fake.com"}))
				Expect(err.Cause).To(Equal(&causeError{error: errPerm}))
				Expect(len(err.stack)).To(Equal(2))
				Expect(err.stack[0].File).To(ContainSubstring("errors_test.go"))
				Expect(err.stack[0].Line).To(BeNumerically(">", 0))
				Expect(err.stack[0].Function).NotTo(BeEmpty())
				Expect(err.stack[1].File).To(ContainSubstring("errors_test.go"))
				Expect(err.stack[1].Line).To(BeNumerically(">", 0))
				Expect(err.stack[1].Function).NotTo(BeEmpty())
			}
		})
	})

	Context("When printing an error", func() {
		It("should return the error as a string", func() {
			e1 = Wrap(ErrForbidden,
				WithIdentifier(128),
				WithDetail("missing write permission on the file"),
				WithProperty("File", "test.txt"),
				CausedBy(errPerm),
			)
			e = Wrap(e1,
				WithIdentifier(932),
				WithDetail("custom client role is 'Reader'"),
				WithProperty("ClientID", "1234567890"),
			)
			result := e.Error()

			// Replace line number and function references
			result = regexp.MustCompile(`line='\d+'`).ReplaceAllString(result, "line=''")
			result = regexp.MustCompile(`func='[a-z0-9\/\.\-]*'`).ReplaceAllString(result, "func=''")
			Expect(result).To(ContainSubstring("forbidden (932-128):"))
			Expect(result).To(ContainSubstring("custom client role is 'Reader':"))
			Expect(result).To(ContainSubstring("missing write permission on the file:"))
			Expect(result).To(ContainSubstring("File='test.txt'"))
			Expect(result).To(ContainSubstring("ClientID='1234567890'"))
			Expect(result).To(ContainSubstring("at=[(func='', file='errors_test.go', line=''), (func='', file='errors_test.go', line='')]"))
			Expect(result).To(ContainSubstring("caused by: permission denied"))
		})
	})

	Context("When getting JSON representation of an error", func() {
		It("should support the %v marker to print the error as simple JSON string", func() {
			e1 = Wrap(ErrForbidden,
				WithIdentifier(127),
				WithDetail("missing write permission on the file"),
				WithProperty("File", "test.txt"),
				CausedBy(errPerm),
			)
			e = Wrap(e1,
				WithIdentifier(432),
				WithDetail("custom client role is 'Reader'"),
				WithProperty("ClientID", "1234567890"),
			)
			result := fmt.Sprintf("%v", e)

			// Replace newline character
			result = strings.TrimSuffix(result, "\n")
			expected := "{" +
				"\"title\":\"forbidden\"," +
				"\"identifier\":[127,432]," +
				"\"details\":[\"missing write permission on the file\",\"custom client role is 'Reader'\"]," +
				"\"properties\":{\"ClientID\":\"1234567890\",\"File\":\"test.txt\"}," +
				"\"cause\":\"permission denied\"" +
				"}"
			Expect(result).To(Equal(expected))
		})

		It("should support the %+v marker to print the error as extended JSON string", func() {
			e1 = Wrap(ErrForbidden,
				WithIdentifier(127),
				WithDetail("missing write permission on the file"),
				WithProperty("File", "test.txt"),
				CausedBy(errPerm),
			)
			e = Wrap(e1,
				WithIdentifier(432),
				WithDetail("custom client role is 'Reader'"),
				WithProperty("ClientID", "1234567890"),
			)
			result := fmt.Sprintf("%+v", e)

			// Replace newline character, line numbers and function references
			result = strings.TrimSuffix(result, "\n")
			result = regexp.MustCompile(`"line":\d+`).ReplaceAllString(result, "\"line\":0")
			result = regexp.MustCompile(`"function":"[a-z0-9\/\.\-]*"`).ReplaceAllString(result, "\"function\":\"\"")
			result = regexp.MustCompile(`"file":"[a-zA-Z0-9\/\.\-_]*"`).ReplaceAllString(result, "\"file\":\"\"")
			expected := "{" +
				"\"title\":\"forbidden\"," +
				"\"identifier\":[127,432]," +
				"\"details\":[\"missing write permission on the file\",\"custom client role is 'Reader'\"]," +
				"\"properties\":{\"ClientID\":\"1234567890\",\"File\":\"test.txt\"}," +
				"\"cause\":\"permission denied\"," +
				"\"stack\":[{\"function\":\"\",\"file\":\"\",\"line\":0},{\"function\":\"\",\"file\":\"\",\"line\":0}]" +
				"}"
			Expect(result).To(Equal(expected))
		})
	})

	Context("When calling Error() on nil", func() {
		It("should return empty string", func() {
			var e *Error
			Expect(e.Error()).To(Equal(""))
		})
	})

	Context("When wrapping nil", func() {
		It("should return unknown error with nil cause", func() {
			e = Wrap(nil)
			var err *Error
			Expect(errors.As(e, &err)).To(BeTrue())
			Expect(err.Title).To(Equal("unknown error"))
			Expect(err.Cause).To(BeNil())
		})
	})

	Context("When comparing errors with Is()", func() {
		It("should return true for errors with same title and identifier", func() {
			e1 = Wrap(ErrForbidden, WithIdentifier(403001))
			e2 = Wrap(ErrForbidden, WithIdentifier(403001))
			Expect(Is(e1, e2)).To(BeTrue())
		})

		It("should return false for errors with different titles", func() {
			e1 = Wrap(ErrForbidden, WithIdentifier(403001))
			e2 = Wrap(ErrNotFound, WithIdentifier(403001))
			Expect(Is(e1, e2)).To(BeFalse())
		})

		It("should return false for errors with different identifiers", func() {
			e1 = Wrap(ErrForbidden, WithIdentifier(403001))
			e2 = Wrap(ErrForbidden, WithIdentifier(403002))
			Expect(Is(e1, e2)).To(BeFalse())
		})

		It("should return false for standard errors", func() {
			e := Wrap(ErrForbidden, WithIdentifier(403001))
			Expect(Is(e, errTest)).To(BeFalse())
		})

		It("should work with errors.Is() from standard library", func() {
			e1 = Wrap(ErrForbidden, WithIdentifier(403001))
			e2 = Wrap(ErrForbidden, WithIdentifier(403001))
			Expect(errors.Is(e1, e2)).To(BeTrue())
		})

		It("should return false when target is nil", func() {
			e := Wrap(ErrForbidden, WithIdentifier(403001))
			Expect(Is(e, nil)).To(BeFalse())
		})

		It("should return false when error is nil", func() {
			e := Wrap(ErrForbidden, WithIdentifier(403001))
			Expect(Is(nil, e)).To(BeFalse())
		})

		It("should return true for wrapped errors with same title and identifier", func() {
			e1_1 := Wrap(ErrForbidden, WithIdentifier(403001))
			e1_2 := Wrap(e1_1, WithIdentifier(403002))
			e2_1 := Wrap(ErrForbidden, WithIdentifier(403001))
			e2_2 := Wrap(e2_1, WithIdentifier(403002))
			Expect(Is(e1_2, e2_2)).To(BeTrue())
		})

		It("should return false for wrapped errors with different identifier", func() {
			e1_1 := Wrap(ErrForbidden, WithIdentifier(403001))
			e1_2 := Wrap(e1_1, WithIdentifier(403002))
			e2_1 := Wrap(ErrForbidden, WithIdentifier(403001))
			e2_2 := Wrap(e2_1, WithIdentifier(403003))
			Expect(Is(e1_2, e2_2)).To(BeFalse())
		})

		It("should return true for error and its child", func() {
			e1_1 := Wrap(ErrForbidden, WithIdentifier(1))
			e1_2 := Wrap(e1_1, WithIdentifier(2))
			e2_1 := Wrap(ErrForbidden, WithIdentifier(1))
			e2_2 := Wrap(e2_1, WithIdentifier(2))
			e2_3 := Wrap(e2_2, WithIdentifier(3))
			Expect(Is(e2_3, e1_2)).To(BeTrue())
			Expect(Is(e1_2, e2_3)).To(BeFalse())
		})

		It("should return false for error and its parent", func() {
			e1_1 := Wrap(ErrForbidden, WithIdentifier(2))
			e1_2 := Wrap(e1_1, WithIdentifier(3))
			e2_1 := Wrap(ErrForbidden, WithIdentifier(1))
			e2_2 := Wrap(e2_1, WithIdentifier(2))
			e2_3 := Wrap(e2_2, WithIdentifier(3))
			Expect(Is(e2_3, e1_2)).To(BeFalse())
		})
	})

	Context("When unwrapping errors", func() {
		It("should return the cause when present", func() {
			e := Wrap(ErrForbidden, CausedBy(errPerm))
			Expect(Unwrap(e)).To(Equal(errPerm))
		})

		It("should return the cause when error is a standard one", func() {
			err := errors.New("test error")
			e := Wrap(err)
			Expect(Unwrap(e)).To(Equal(err))
		})

		It("should return nil when no cause", func() {
			e := Wrap(ErrForbidden)
			Expect(Unwrap(e)).To(BeNil())
		})

		It("should return nil for nil error", func() {
			var e *Error
			Expect(Unwrap(e)).To(BeNil())
		})

		It("should return nil when error has no cause", func() {
			e := Wrap(ErrNotFoundWithOptions)
			Expect(Unwrap(e)).To(BeNil())
		})

		It("should work with errors.Unwrap() from standard library", func() {
			e := Wrap(ErrForbidden, CausedBy(errPerm))
			Expect(errors.Unwrap(e)).To(Equal(errPerm))
		})
	})

	Context("When using As()", func() {
		It("should return true when error is *Error", func() {
			e := Wrap(ErrForbidden, WithIdentifier(403001))
			var target *Error
			ok := As(e, &target)
			Expect(ok).To(BeTrue())
		})

		It("should return true when error wraps *Error", func() {
			inner := Wrap(ErrNotFound, WithIdentifier(404001))
			e := Wrap(inner, WithDetail("file missing"))
			var target *Error
			ok := As(e, &target)
			Expect(ok).To(BeTrue())
		})

		It("should return false for nil error", func() {
			var target *Error
			ok := As(nil, &target)
			Expect(ok).To(BeFalse())
		})

		It("should return false when error is not an *Error", func() {
			e := errTest
			var target *Error
			ok := As(e, &target)
			Expect(ok).To(BeFalse())
		})

		It("should match standard library errors.As behavior", func() {
			e := Wrap(ErrForbidden, CausedBy(errPerm))
			var pkgTarget *Error
			var stdTarget *Error
			pkgOk := As(e, &pkgTarget)
			stdOk := errors.As(e, &stdTarget)
			Expect(pkgOk).To(BeTrue())
			Expect(stdOk).To(BeTrue())
		})
	})
})
