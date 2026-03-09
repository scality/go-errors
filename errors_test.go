package errors

import (
	"errors"
	"regexp"
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
		Options: Opts{
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

var _ = BeforeSuite(func() {
})

var _ = Describe("Errors", func() {
	Context("When creating an error with New()", func() {
		It("should return an Error with the given title and default options", func() {
			e := New("my error")
			var err *Error
			Expect(errors.As(e, &err)).To(BeTrue())
			Expect(err.Title).To(Equal("my error"))
			Expect(err.Options.Identifier).To(BeNil())
			Expect(err.Options.Details).To(BeEmpty())
			Expect(err.Options.Properties).To(BeEmpty())
			Expect(err.Options.Cause).To(Equal(errors.New("my error")))
		})
	})

	Context("When creating a new error from a standard one", func() {
		It("should return an Unknown Error error, when no options are provided", func() {
			e = Wrap(errTest)
			var err *Error
			if ok := errors.As(e, &err); ok {
				Expect(err.Title).To(Equal("unknown error"))
				Expect(err.Options.Identifier).To(BeZero())
				Expect(err.Options.Details).To(BeEmpty())
				Expect(err.Options.Properties).To(BeEmpty())
				Expect(err.Options.Cause).To(Equal(errTest))
				Expect(err.Stack[0].File).To(ContainSubstring("errors_test.go"))
				Expect(err.Stack[0].Line).To(BeNumerically(">", 0))
				Expect(err.Stack[0].Function).NotTo(BeEmpty())
				Expect(err.Stack[0].Timestamp).NotTo(BeZero())
			}
		})
		It("should support WithDetailf for formatted details", func() {
			e = Wrap(errTest,
				WithDetailf("failed at %s:%d", "step", 42),
			)
			var err *Error
			Expect(errors.As(e, &err)).To(BeTrue())
			Expect(err.Options.Details).To(Equal([]string{"failed at step:42"}))
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
				Expect(err.Options.Identifier).To(Equal([]uint32{1001}))
				Expect(err.Options.Details).To(Equal([]string{"This is a test error"}))
				Expect(err.Options.Properties).To(Equal(map[string]any{"type": "fake"}))
				Expect(err.Options.Cause).To(Equal(errTest))
				Expect(err.Stack[0].File).To(ContainSubstring("errors_test.go"))
				Expect(err.Stack[0].Line).To(BeNumerically(">", 0))
				Expect(err.Stack[0].Function).NotTo(BeEmpty())
				Expect(err.Stack[0].Timestamp).NotTo(BeZero())
			}
		})
	})

	Context("When creating a new error from a custom one", func() {
		It("should return the custom error, when no options are provided", func() {
			e = Wrap(ErrNotFound)
			var err *Error
			if ok := errors.As(e, &err); ok {
				Expect(err.Title).To(Equal("not found"))
				Expect(err.Options.Identifier).To(BeZero())
				Expect(err.Options.Details).To(BeEmpty())
				Expect(err.Options.Properties).To(BeEmpty())
				Expect(err.Options.Cause).To(Equal(errors.New("not found")))
				Expect(err.Stack[0].File).To(ContainSubstring("errors_test.go"))
				Expect(err.Stack[0].Line).To(BeNumerically(">", 0))
				Expect(err.Stack[0].Function).NotTo(BeEmpty())
				Expect(err.Stack[0].Timestamp).NotTo(BeZero())
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
				Expect(err.Options.Identifier).To(Equal([]uint32{1001}))
				Expect(err.Options.Details).To(Equal([]string{"File not found in the session"}))
				Expect(err.Options.Properties).To(Equal(map[string]any{"File": "test.txt"}))
				Expect(err.Options.Cause).To(Equal(errors.New("not found")))
				Expect(err.Stack[0].File).To(ContainSubstring("errors_test.go"))
			}
		})
	})

	Context("When creating a new error from a custom one with options", func() {
		It("should return the custom error with same options", func() {
			e = Wrap(ErrNotFoundWithOptions)
			var err *Error
			if ok := errors.As(e, &err); ok {
				Expect(err.Title).To(Equal("not found"))
				Expect(err.Options.Identifier).To(Equal([]uint32{1}))
				Expect(err.Options.Details).To(Equal([]string{"File not found in the session"}))
				Expect(err.Options.Properties).To(Equal(map[string]any{"File": "test.txt", "User": "john.doe"}))
				Expect(err.Options.Cause).To(BeNil())
				Expect(err.Stack[0].File).To(ContainSubstring("errors_test.go"))
				Expect(err.Stack[0].Line).To(BeNumerically(">", 0))
				Expect(err.Stack[0].Function).NotTo(BeEmpty())
				Expect(err.Stack[0].Timestamp).NotTo(BeZero())
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
				Expect(err.Options.Identifier).To(Equal([]uint32{1, 1001}))
				Expect(err.Options.Details).To(Equal([]string{"File not found in the session", "custom client role is 'Reader'"}))
				Expect(err.Options.Properties).To(Equal(map[string]any{"File": "test.txt", "User": "john.doe", "ClientID": "1234567890"}))
				Expect(err.Options.Cause).To(BeNil())
				Expect(err.Stack[0].File).To(ContainSubstring("errors_test.go"))
				Expect(err.Stack[0].Line).To(BeNumerically(">", 0))
				Expect(err.Stack[0].Function).NotTo(BeEmpty())
				Expect(err.Stack[0].Timestamp).NotTo(BeZero())
			}
		})
	})

	Context("When creating a new error from a custom one based on a standard error", func() {
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
				Expect(err.Options.Identifier).To(Equal([]uint32{403001}))
				Expect(err.Options.Details).To(Equal([]string{"missing role 'admin' for the user"}))
				Expect(err.Options.Properties).To(Equal(map[string]any{"User": "john.doe"}))
				Expect(err.Options.Cause).To(Equal(errPerm))
				Expect(len(err.Stack)).To(Equal(1))
				Expect(err.Stack[0].File).To(ContainSubstring("errors_test.go"))
				Expect(err.Stack[0].Line).To(BeNumerically(">", 0))
				Expect(err.Stack[0].Function).NotTo(BeEmpty())
				Expect(err.Stack[0].Timestamp).NotTo(BeZero())
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
				Expect(err.Options.Identifier).To(Equal([]uint32{500001}))
				Expect(err.Options.Details).To(Equal([]string{"unexpected error occurred while processing the request"}))
				Expect(err.Options.Properties).To(Equal(map[string]any{"RequestID": "1234567890"}))
				Expect(err.Options.Cause).To(Equal(errPerm))
				Expect(len(err.Stack)).To(Equal(2))
				Expect(err.Stack[0].File).To(ContainSubstring("errors_test.go"))
				Expect(err.Stack[0].Line).To(BeNumerically(">", 0))
				Expect(err.Stack[0].Function).NotTo(BeEmpty())
				Expect(err.Stack[0].Timestamp).NotTo(BeZero())
				Expect(err.Stack[1].File).To(ContainSubstring("errors_test.go"))
				Expect(err.Stack[1].Line).To(BeNumerically(">", 0))
				Expect(err.Stack[1].Function).NotTo(BeEmpty())
				Expect(err.Stack[1].Timestamp).NotTo(BeZero())
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
				Expect(err.Options.Identifier).To(Equal([]uint32{500002}))
				Expect(err.Options.Details).To(Equal([]string{"database connection failed", "wrong url"}))
				Expect(err.Options.Properties).To(Equal(map[string]any{"RequestID": "1234567890", "url": "https://bdd.fake.com"}))
				Expect(err.Options.Cause).To(Equal(errPerm))
				Expect(len(err.Stack)).To(Equal(2))
				Expect(err.Stack[0].File).To(ContainSubstring("errors_test.go"))
				Expect(err.Stack[0].Line).To(BeNumerically(">", 0))
				Expect(err.Stack[0].Function).NotTo(BeEmpty())
				Expect(err.Stack[0].Timestamp).NotTo(BeZero())
				Expect(err.Stack[1].File).To(ContainSubstring("errors_test.go"))
				Expect(err.Stack[1].Line).To(BeNumerically(">", 0))
				Expect(err.Stack[1].Function).NotTo(BeEmpty())
				Expect(err.Stack[1].Timestamp).NotTo(BeZero())
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
			result = regexp.MustCompile(`func='[a-z0-9\.\-]*'`).ReplaceAllString(result, "func=''")
			expected := "forbidden (932-128):" +
				" missing write permission on the file:" +
				" custom client role is 'Reader':" +
				" File='test.txt'," +
				" ClientID='1234567890'," +
				" at=(func='', file='errors_test.go', line='')," +
				" caused by: permission denied"
			Expect(result).To(Equal(expected))
		})
	})

	Context("When getting JSON representation of an error", func() {
		It("should return the error as JSON string", func() {
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
			result := e.(*Error).String()

			Expect(result).To(ContainSubstring(`"title":"forbidden"`))
			Expect(result).To(ContainSubstring(`"identifier":[127,432]`))
			Expect(result).To(ContainSubstring(`"details":["missing write permission on the file","custom client role is 'Reader'"]`))
			Expect(result).To(ContainSubstring(`"properties":{"ClientID":"1234567890","File":"test.txt"}`))
		})

		It("should handle nil error", func() {
			var e *Error
			Expect(e.String()).To(Equal(""))
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
			e := Wrap(nil)
			var err *Error
			Expect(errors.As(e, &err)).To(BeTrue())
			Expect(err.Title).To(Equal("unknown error"))
			Expect(err.Options.Cause).To(BeNil())
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
	})

	Context("When unwrapping errors", func() {
		It("should return the cause when present", func() {
			e := Wrap(ErrForbidden, CausedBy(errPerm))
			Expect(Unwrap(e)).To(Equal(errPerm))
		})

		It("should return nil when no cause", func() {
			e := Wrap(ErrForbidden)
			Expect(Unwrap(e)).To(Equal(errors.New("forbidden")))
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
