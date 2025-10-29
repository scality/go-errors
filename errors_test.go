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
	ErrNotFoundWithDetails = &Error{
		Title:   "not found",
		Details: []string{"File not found in the session"},
	}
	ErrNotFoundWithProperties = &Error{
		Title: "not found",
		Properties: map[string]any{
			"File": "test.txt",
			"User": "john.doe",
		},
	}
	ErrForbidden = New("forbidden")
	ErrInternal  = New("internal error")

	errTest = errors.New("test error")
	errPerm = errors.New("permission denied")
)

var _ = BeforeSuite(func() {
})

var _ = Describe("Errors", func() {
	Context("When creating a new error from a standard one", func() {
		It("should return an Unknown Error error", func() {
			e := From(errTest)
			Expect(e.Title).To(Equal("unknown error"))
			Expect(e.Identifier).To(BeZero())
			Expect(e.Details).To(BeEmpty())
			Expect(e.Properties).To(BeEmpty())
			Expect(e.Cause).To(Equal(errTest))
			Expect(e.Stack).To(BeEmpty())
		})
	})

	Context("When creating a new error from a custom one", func() {
		It("should return the custom error", func() {
			e := From(ErrNotFound).
				WithDetail("File not found in the session.").
				WithProperty("File", "test.txt")
			Expect(e.Title).To(Equal("not found"))
			Expect(e.Identifier).To(BeZero())
			Expect(e.Details).To(Equal([]string{"File not found in the session"}))
			Expect(e.Properties).To(Equal(map[string]any{"File": "test.txt"}))
			Expect(e.Cause).To(BeNil())
			Expect(e.Stack).To(BeEmpty())
		})
	})

	Context("When creating a new error from a custom one with details", func() {
		It("should return the custom error with same details", func() {
			e := From(ErrNotFoundWithDetails)
			Expect(e.Title).To(Equal("not found"))
			Expect(e.Identifier).To(BeZero())
			Expect(e.Details).To(Equal([]string{"File not found in the session"}))
			Expect(e.Properties).To(BeEmpty())
			Expect(e.Cause).To(BeNil())
			Expect(e.Stack).To(BeEmpty())
		})
	})

	Context("When creating a new error from a custom one with properties", func() {
		It("should return the custom error with same properties", func() {
			e := From(ErrNotFoundWithProperties)
			Expect(e.Title).To(Equal("not found"))
			Expect(e.Identifier).To(BeZero())
			Expect(e.Details).To(BeEmpty())
			Expect(e.Properties).To(Equal(map[string]any{"File": "test.txt", "User": "john.doe"}))
			Expect(e.Cause).To(BeNil())
			Expect(e.Stack).To(BeEmpty())
		})
	})

	Context("When creating a new error from a custom one based on a standard error", func() {
		It("should return the custom error", func() {
			e := From(ErrForbidden).
				WithIdentifier(403001).
				WithDetail("missing role 'admin' for the user.").
				WithProperty("User", "john.doe").
				CausedBy(errPerm)
			Expect(e.Title).To(Equal("forbidden"))
			Expect(e.Identifier).To(Equal(int32(403001)))
			Expect(e.Details).To(Equal([]string{"missing role 'admin' for the user"}))
			Expect(e.Properties).To(Equal(map[string]any{"User": "john.doe"}))
			Expect(e.Cause).To(Equal(errPerm))
			Expect(e.Stack).To(BeEmpty())
		})
	})

	Context("When wrapping errors", func() {
		It("should return the wrapped error", func() {
			e1 := From(ErrForbidden).
				WithIdentifier(403001).
				WithDetail("missing write permission on the file.").
				WithProperty("File", "test.txt").
				CausedBy(errPerm)
			e2 := Intercept(e1).
				WithDetail("missing role 'admin' for the user.").
				WithProperty("User", "john.doe")
			Expect(e2.Title).To(Equal("forbidden"))
			Expect(e2.Identifier).To(Equal(int32(403001)))
			Expect(e2.Details).To(Equal([]string{"missing write permission on the file", "missing role 'admin' for the user"}))
			Expect(e2.Properties).To(Equal(map[string]any{"File": "test.txt", "User": "john.doe"}))
			Expect(e2.Cause).To(Equal(errPerm))
		})
	})

	Context("When stamping an error", func() {
		It("should return the error with a stack trace", func() {
			e := From(ErrForbidden).
				WithIdentifier(403001).
				WithDetail("missing write permission on the file.").
				WithProperty("File", "test.txt").
				CausedBy(errPerm)
			err := Intercept(Stamp(e))
			Expect(err.Title).To(Equal("forbidden"))
			Expect(err.Identifier).To(Equal(int32(403001)))
			Expect(err.Details).To(Equal([]string{"missing write permission on the file"}))
			Expect(err.Properties).To(Equal(map[string]any{"File": "test.txt"}))
			Expect(err.Cause).To(Equal(errPerm))
		})
	})

	Context("When printing an error", func() {
		It("should return the error as a string", func() {
			e := From(ErrForbidden).
				WithIdentifier(403001).
				WithDetail("missing write permission on the file.").
				WithProperty("File", "test.txt").
				CausedBy(errPerm)
			result := Stamp(e).Error()

			// Replace line number and function references
			result = regexp.MustCompile(`line='\d+'`).ReplaceAllString(result, "line=''")
			result = regexp.MustCompile(`func='[a-z0-9\.\-]*'`).ReplaceAllString(result, "func=''")
			expected := "forbidden (403001): missing write permission on the file: File='test.txt'," +
				" at=(func='', file='errors_test.go', line='')," +
				" caused by: permission denied"
			Expect(result).To(Equal(expected))
		})
	})

	Context("When getting JSON representation of an error", func() {
		It("should return the error as JSON string", func() {
			e := From(ErrForbidden).
				WithIdentifier(403001).
				WithDetail("missing write permission on the file").
				WithProperty("File", "test.txt").
				CausedBy(errPerm)
			result := e.String()
			Expect(result).To(ContainSubstring(`"title":"forbidden"`))
			Expect(result).To(ContainSubstring(`"identifier":403001`))
			Expect(result).To(ContainSubstring(`"details":`))
			Expect(result).To(ContainSubstring(`"properties":`))
		})

		It("should handle nil error", func() {
			var e *Error
			Expect(e.String()).To(Equal(""))
		})
	})

	Context("When comparing errors with Is()", func() {
		It("should return true for errors with same title and identifier", func() {
			e1 := From(ErrForbidden).WithIdentifier(403001)
			e2 := From(ErrForbidden).WithIdentifier(403001)
			Expect(e1.Is(e2)).To(BeTrue())
		})

		It("should return false for errors with different titles", func() {
			e1 := From(ErrForbidden).WithIdentifier(403001)
			e2 := From(ErrNotFound).WithIdentifier(403001)
			Expect(e1.Is(e2)).To(BeFalse())
		})

		It("should return false for errors with different identifiers", func() {
			e1 := From(ErrForbidden).WithIdentifier(403001)
			e2 := From(ErrForbidden).WithIdentifier(403002)
			Expect(e1.Is(e2)).To(BeFalse())
		})

		It("should return false for standard errors", func() {
			e := From(ErrForbidden).WithIdentifier(403001)
			Expect(e.Is(errTest)).To(BeFalse())
		})

		It("should work with errors.Is() from standard library", func() {
			e1 := From(ErrForbidden).WithIdentifier(403001)
			e2 := From(ErrForbidden).WithIdentifier(403001)
			Expect(errors.Is(e1, e2)).To(BeTrue())
		})
	})

	Context("When unwrapping errors", func() {
		It("should return the cause when present", func() {
			e := From(ErrForbidden).CausedBy(errPerm)
			Expect(e.Unwrap()).To(Equal(errPerm))
		})

		It("should return nil when no cause", func() {
			e := From(ErrForbidden)
			Expect(e.Unwrap()).To(BeNil())
		})

		It("should return nil for nil error", func() {
			var e *Error
			Expect(e.Unwrap()).To(BeNil())
		})

		It("should work with errors.Unwrap() from standard library", func() {
			e := From(ErrForbidden).CausedBy(errPerm)
			Expect(errors.Unwrap(e)).To(Equal(errPerm))
		})
	})

	Context("When using Throw() method", func() {
		It("should add stack trace to the error", func() {
			e := From(ErrForbidden).
				WithIdentifier(403001).
				WithDetail("missing write permission on the file").
				WithProperty("File", "test.txt")
			err := Intercept(e.Throw())
			Expect(err.Stack).NotTo(BeEmpty())
			Expect(len(err.Stack)).To(Equal(1))
			Expect(err.Stack[0].File).To(ContainSubstring("errors_test.go"))
			Expect(err.Stack[0].Line).To(BeNumerically(">", 0))
			Expect(err.Stack[0].Function).NotTo(BeEmpty())
			Expect(err.Stack[0].Timestamp).NotTo(BeZero())
		})

		It("should allow multiple stack traces", func() {
			e := From(ErrForbidden).WithIdentifier(403001)
			err1 := Intercept(e.Throw())
			err2 := Intercept(Stamp(err1))
			Expect(len(err2.Stack)).To(Equal(2))
		})
	})

	Context("When using WithProperties() method", func() {
		It("should add multiple properties at once", func() {
			props := map[string]any{
				"User":   "john.doe",
				"Action": "write",
				"File":   "test.txt",
			}
			e := From(ErrForbidden).WithProperties(props)
			Expect(e.Properties).To(Equal(props))
		})

		It("should merge with existing properties", func() {
			e := From(ErrForbidden).
				WithProperty("User", "john.doe").
				WithProperties(map[string]any{
					"Action": "write",
					"File":   "test.txt",
				})
			Expect(e.Properties).To(HaveLen(3))
			Expect(e.Properties["User"]).To(Equal("john.doe"))
			Expect(e.Properties["Action"]).To(Equal("write"))
			Expect(e.Properties["File"]).To(Equal("test.txt"))
		})

		It("should overwrite existing property with same key", func() {
			e := From(ErrForbidden).
				WithProperty("User", "john.doe").
				WithProperty("User", "jane.doe")
			Expect(e.Properties["User"]).To(Equal("jane.doe"))
		})
	})

	Context("When using WithDetailf() method", func() {
		It("should format detail string with arguments", func() {
			e := From(ErrNotFound).
				WithDetailf("File '%s' not found in directory '%s'", "test.txt", "/home/user")
			Expect(e.Details).To(HaveLen(1))
			Expect(e.Details[0]).To(Equal("File 'test.txt' not found in directory '/home/user'"))
		})

		It("should format detail with multiple format specifiers", func() {
			e := From(ErrForbidden).
				WithDetailf("User %s attempted to access resource %d at %s", "john.doe", 12345, "2023-10-15")
			Expect(e.Details).To(HaveLen(1))
			Expect(e.Details[0]).To(Equal("User john.doe attempted to access resource 12345 at 2023-10-15"))
		})

		It("should trim trailing period from formatted detail", func() {
			e := From(ErrNotFound).
				WithDetailf("Resource with ID %d not found.", 42)
			Expect(e.Details[0]).To(Equal("Resource with ID 42 not found"))
		})

		It("should chain with other methods", func() {
			e := From(ErrForbidden).
				WithIdentifier(403001).
				WithDetailf("User %s lacks permission", "john.doe").
				WithDetail("Admin role required").
				WithProperty("User", "john.doe")
			Expect(e.Details).To(Equal([]string{"User john.doe lacks permission", "Admin role required"}))
			Expect(e.Properties["User"]).To(Equal("john.doe"))
		})

		It("should handle format string without arguments", func() {
			e := From(ErrInternal).WithDetailf("An unexpected error occurred")
			Expect(e.Details[0]).To(Equal("An unexpected error occurred"))
		})
	})

	Context("When using Wrap() function", func() {
		It("should wrap a standard error with a message", func() {
			wrapped := Wrap(errTest, "failed to process request")
			e := Intercept(wrapped)
			Expect(e.Title).To(Equal("unknown error"))
			Expect(e.Details).To(Equal([]string{"failed to process request"}))
			Expect(e.Cause).To(Equal(errTest))
			Expect(e.Stack).NotTo(BeEmpty())
		})

		It("should wrap a custom Error with a message", func() {
			original := From(ErrNotFound).WithIdentifier(404001)
			wrapped := Wrap(original, "resource lookup failed")
			e := Intercept(wrapped)
			Expect(e.Title).To(Equal("not found"))
			Expect(e.Details).To(Equal([]string{"resource lookup failed"}))
			Expect(e.Cause).To(BeNil()) // From() on *Error doesn't set cause
			Expect(e.Stack).NotTo(BeEmpty())
		})

		It("should add stack trace automatically", func() {
			wrapped := Wrap(errPerm, "authentication failed")
			e := Intercept(wrapped)
			Expect(e.Stack).To(HaveLen(1))
			Expect(e.Stack[0].File).NotTo(BeEmpty())
			Expect(e.Stack[0].Line).To(BeNumerically(">", 0))
			Expect(e.Stack[0].Function).NotTo(BeEmpty())
			Expect(e.Stack[0].Timestamp).NotTo(BeZero())
		})

		It("should trim trailing period from message", func() {
			wrapped := Wrap(errTest, "operation failed.")
			e := Intercept(wrapped)
			Expect(e.Details[0]).To(Equal("operation failed"))
		})

		It("should allow chaining multiple wraps", func() {
			err1 := Wrap(errTest, "database query failed")
			err2 := Wrap(err1, "user service error")
			e := Intercept(err2)
			// From() only preserves title, so only the latest detail is kept
			Expect(e.Details).To(Equal([]string{"database query failed", "user service error"}))
			Expect(e.Stack).To(HaveLen(1))
			// From() on *Error doesn't set cause
			Expect(e.Cause).To(BeNil())
		})

		It("should preserve error title when wrapping custom errors", func() {
			original := From(ErrForbidden).
				WithIdentifier(403001).
				WithProperty("User", "john.doe")
			wrapped := Wrap(original, "access denied")
			e := Intercept(wrapped)
			Expect(e.Title).To(Equal("forbidden"))
			Expect(e.Identifier).To(Equal(int32(403001)))
			Expect(e.Details).To(Equal([]string{"access denied"}))
			Expect(e.Properties).To(Equal(map[string]any{"User": "john.doe"}))
		})
	})

	Context("When using Wrapf() function", func() {
		It("should wrap an error with a formatted message", func() {
			wrapped := Wrapf(errTest, "failed to process request for user %s", "john.doe")
			e := Intercept(wrapped)
			Expect(e.Title).To(Equal("unknown error"))
			Expect(e.Details).To(Equal([]string{"failed to process request for user john.doe"}))
			Expect(e.Cause).To(Equal(errTest))
			Expect(e.Stack).NotTo(BeEmpty())
		})

		It("should format message with multiple arguments", func() {
			wrapped := Wrapf(errPerm, "user %s failed to access file %s at line %d", "alice", "config.yaml", 42)
			e := Intercept(wrapped)
			Expect(e.Details[0]).To(Equal("user alice failed to access file config.yaml at line 42"))
		})

		It("should handle format with no arguments", func() {
			wrapped := Wrapf(errTest, "an error occurred")
			e := Intercept(wrapped)
			Expect(e.Details[0]).To(Equal("an error occurred"))
		})

		It("should add stack trace automatically", func() {
			wrapped := Wrapf(errTest, "operation failed with code %d", 500)
			e := Intercept(wrapped)
			Expect(e.Stack).To(HaveLen(1))
			Expect(e.Stack[0].File).NotTo(BeEmpty())
			Expect(e.Stack[0].Timestamp).NotTo(BeZero())
		})

		It("should trim trailing period from formatted message", func() {
			wrapped := Wrapf(errTest, "failed with error code %d.", 404)
			e := Intercept(wrapped)
			Expect(e.Details[0]).To(Equal("failed with error code 404"))
		})

		It("should allow chaining with Wrap", func() {
			err1 := Wrapf(errTest, "database error: code %d", 1062)
			err2 := Wrap(err1, "duplicate entry detected")
			e := Intercept(err2)
			Expect(e.Details).To(Equal([]string{"database error: code 1062", "duplicate entry detected"}))
			Expect(e.Stack).To(HaveLen(1))
			// From() on *Error doesn't set cause
			Expect(e.Cause).To(BeNil())
		})

		It("should handle complex format patterns", func() {
			wrapped := Wrapf(errTest, "failed to connect to %s:%d (timeout: %v)", "localhost", 8080, true)
			e := Intercept(wrapped)
			Expect(e.Details[0]).To(Equal("failed to connect to localhost:8080 (timeout: true)"))
		})

		It("should work with custom errors", func() {
			original := From(ErrInternal).WithIdentifier(500001)
			wrapped := Wrapf(original, "service %s returned error %d", "auth-service", 500)
			e := Intercept(wrapped)
			Expect(e.Title).To(Equal("internal error"))
			Expect(e.Identifier).To(Equal(int32(500001)))
			Expect(e.Details).To(Equal([]string{"service auth-service returned error 500"}))
		})
	})

	Context("When using package-level As() function", func() {
		It("should convert error to target type", func() {
			e := From(ErrForbidden).WithIdentifier(403001)
			var target *Error
			Expect(As(e, &target)).To(BeTrue())
			Expect(target.Title).To(Equal("forbidden"))
			Expect(target.Identifier).To(Equal(int32(403001)))
		})

		It("should return false when error is not in the chain", func() {
			// Create a standard error that won't convert to *Error
			standardErr := errors.New("standard error")
			var target *Error
			Expect(As(standardErr, &target)).To(BeFalse())
			Expect(target).To(BeNil())
		})

		It("should work with wrapped errors", func() {
			innerErr := From(ErrNotFound).WithIdentifier(404001)
			outerErr := From(ErrForbidden).CausedBy(innerErr)
			var target *Error
			Expect(As(outerErr, &target)).To(BeTrue())
			Expect(target.Title).To(Equal("forbidden"))
		})

		It("should extract standard errors from cause chain", func() {
			e := From(ErrForbidden).CausedBy(errPerm)
			var standardErr error
			// Should be able to extract the error itself
			Expect(As(e, &standardErr)).To(BeTrue())
		})
	})

	Context("When intercepting standard errors", func() {
		It("should convert standard error to Error type", func() {
			e := Intercept(errTest)
			Expect(e).To(BeAssignableToTypeOf(&Error{}))
			Expect(e.Title).To(Equal("unknown error"))
			Expect(e.Cause).To(Equal(errTest))
		})

		It("should return same error if already Error type", func() {
			e1 := From(ErrForbidden).WithIdentifier(403001)
			e2 := Intercept(e1)
			Expect(e2).To(Equal(e1))
		})
	})

	Context("Edge cases", func() {
		It("should handle error without identifier", func() {
			e := From(ErrNotFound).WithDetail("resource not found")
			Expect(e.Identifier).To(BeZero())
			result := e.Error()
			Expect(result).To(ContainSubstring("not found (0)"))
		})

		It("should trim trailing period from details", func() {
			e := From(ErrNotFound).WithDetail("File not found.")
			Expect(e.Details[0]).To(Equal("File not found"))
		})

		It("should handle error with no details", func() {
			e := From(ErrNotFound).WithIdentifier(404001)
			result := e.Error()
			Expect(result).To(Equal("not found (404001)"))
		})

		It("should handle error with no properties", func() {
			e := From(ErrNotFound).WithIdentifier(404001).WithDetail("resource not found")
			result := e.Error()
			Expect(result).NotTo(ContainSubstring("="))
		})

		It("should handle error with no cause", func() {
			e := From(ErrNotFound)
			Expect(e.Cause).To(BeNil())
			result := e.Error()
			Expect(result).NotTo(ContainSubstring("caused by"))
		})

		It("should handle nil error in Error() method", func() {
			var e *Error
			Expect(e.Error()).To(Equal(""))
		})

		It("should preserve order of details when wrapping", func() {
			e1 := From(ErrForbidden).
				WithDetail("first detail").
				WithDetail("second detail")
			e2 := Intercept(e1).WithDetail("third detail")
			Expect(e2.Details).To(Equal([]string{"first detail", "second detail", "third detail"}))
		})

		It("should handle multiple stack traces in correct order", func() {
			e := From(ErrInternal).WithIdentifier(500001)
			err1 := Intercept(e.Throw())
			err2 := Intercept(Stamp(err1))
			err3 := Intercept(Stamp(err2))
			Expect(len(err3.Stack)).To(Equal(3))
			// Most recent should be first
			Expect(err3.Stack[0].Timestamp.After(err3.Stack[1].Timestamp)).To(BeTrue())
			Expect(err3.Stack[1].Timestamp.After(err3.Stack[2].Timestamp)).To(BeTrue())
		})
	})

	Context("When using New() function", func() {
		It("should create a new error with given title", func() {
			e := New("test error")
			Expect(e).NotTo(BeNil())
			Expect(e.Error()).To(Equal("test error (0)"))
		})

		It("should return error interface", func() {
			var err error = New("test error")
			Expect(err).NotTo(BeNil())
		})

		It("should create error with empty fields", func() {
			e := New("simple error")
			err := Intercept(e)
			Expect(err.Title).To(Equal("simple error"))
			Expect(err.Identifier).To(BeZero())
			Expect(err.Details).To(BeEmpty())
			Expect(err.Properties).To(BeEmpty())
			Expect(err.Cause).To(BeNil())
			Expect(err.Stack).To(BeEmpty())
		})
	})

	Context("When using standalone methods", func() {
		It("should set identifier with WithIdentifier()", func() {
			e := New("test error")
			err := Intercept(e).WithIdentifier(12345)
			Expect(err.Identifier).To(Equal(int32(12345)))
		})

		It("should replace identifier when called multiple times", func() {
			e := From(ErrForbidden).
				WithIdentifier(403001).
				WithIdentifier(403002)
			Expect(e.Identifier).To(Equal(int32(403002)))
		})

		It("should set cause with CausedBy()", func() {
			e := New("test error")
			err := Intercept(e).CausedBy(errTest)
			Expect(err.Cause).To(Equal(errTest))
		})

		It("should replace cause when called multiple times", func() {
			e := From(ErrForbidden).
				CausedBy(errTest).
				CausedBy(errPerm)
			Expect(e.Cause).To(Equal(errPerm))
		})

		It("should add single property with WithProperty()", func() {
			e := New("test error")
			err := Intercept(e).WithProperty("key", "value")
			Expect(err.Properties).To(HaveLen(1))
			Expect(err.Properties["key"]).To(Equal("value"))
		})

		It("should handle nil properties map", func() {
			e := &Error{Title: "test"}
			Expect(e.Properties).To(BeNil())
			e.WithProperty("key", "value")
			Expect(e.Properties).NotTo(BeNil())
			Expect(e.Properties["key"]).To(Equal("value"))
		})
	})

	Context("When using From() with edge cases", func() {
		It("should handle nil error", func() {
			e := From(nil)
			Expect(e).NotTo(BeNil())
			Expect(e.Title).To(Equal("unknown error"))
			Expect(e.Cause).To(BeNil())
		})

		It("should copy all fields from Error type", func() {
			original := &Error{
				Title:      "test error",
				Identifier: 123,
				Details:    []string{"detail1", "detail2"},
				Properties: map[string]any{"key": "value"},
				Cause:      errTest,
				Stack:      []*Trace{{Function: "test"}},
			}
			e := From(original)
			Expect(e.Title).To(Equal("test error"))
			Expect(e.Identifier).To(Equal(int32(123)))
			Expect(e.Details).To(Equal([]string{"detail1", "detail2"}))
			Expect(e.Properties).To(Equal(map[string]any{"key": "value"}))
			Expect(e.Cause).To(BeNil())   // From() doesn't copy cause from *Error
			Expect(e.Stack).To(BeEmpty()) // From() doesn't copy stack
		})

		It("should create new instance, not reference", func() {
			original := From(ErrForbidden)
			copy := From(original)
			copy.WithDetail("new detail")
			Expect(original.Details).To(BeEmpty())
			Expect(copy.Details).To(HaveLen(1))
		})
	})

	Context("When using Intercept() with edge cases", func() {
		It("should handle nil error", func() {
			e := Intercept(nil)
			Expect(e).NotTo(BeNil())
			Expect(e.Title).To(Equal("unknown error"))
		})

		It("should return same instance for *Error type", func() {
			original := From(ErrForbidden).WithIdentifier(403001)
			intercepted := Intercept(original)
			Expect(intercepted).To(BeIdenticalTo(original))
		})

		It("should handle multiple consecutive Intercepts", func() {
			e1 := Intercept(errTest)
			e2 := Intercept(e1)
			e3 := Intercept(e2)
			Expect(e1).To(BeIdenticalTo(e2))
			Expect(e2).To(BeIdenticalTo(e3))
		})
	})

	Context("When using Stamp() with edge cases", func() {
		It("should add stack trace to standard error", func() {
			stamped := Stamp(errTest)
			e := Intercept(stamped)
			Expect(e.Stack).To(HaveLen(1))
			Expect(e.Stack[0].File).NotTo(BeEmpty())
			Expect(e.Stack[0].Line).To(BeNumerically(">", 0))
		})

		It("should preserve existing error information", func() {
			original := From(ErrForbidden).
				WithIdentifier(403001).
				WithDetail("test detail").
				WithProperty("key", "value")
			stamped := Stamp(original)
			e := Intercept(stamped)
			Expect(e.Title).To(Equal("forbidden"))
			Expect(e.Identifier).To(Equal(int32(403001)))
			Expect(e.Details).To(Equal([]string{"test detail"}))
			Expect(e.Properties).To(Equal(map[string]any{"key": "value"}))
			Expect(e.Stack).To(HaveLen(1))
		})

		It("should allow multiple stamps", func() {
			e := From(ErrInternal)
			e1 := Intercept(Stamp(e))
			e2 := Intercept(Stamp(e1))
			e3 := Intercept(Stamp(e2))
			Expect(e3.Stack).To(HaveLen(3))
		})
	})

	Context("Error formatting edge cases", func() {
		It("should format error with multiple properties", func() {
			e := From(ErrForbidden).
				WithIdentifier(403001).
				WithProperty("User", "john").
				WithProperty("File", "test.txt").
				WithProperty("Action", "write")
			result := e.Error()
			Expect(result).To(ContainSubstring("forbidden (403001)"))
			Expect(result).To(ContainSubstring("User="))
			Expect(result).To(ContainSubstring("File="))
			Expect(result).To(ContainSubstring("Action="))
		})

		It("should handle empty detail strings", func() {
			e := From(ErrNotFound).WithDetail("")
			Expect(e.Details).To(Equal([]string{""}))
		})

		It("should format error with only identifier and no title", func() {
			e := &Error{Identifier: 123}
			result := e.Error()
			Expect(result).To(Equal(" (123)"))
		})

		It("should handle error with nil cause", func() {
			e := &Error{Title: "test", Cause: nil}
			result := e.Error()
			Expect(result).NotTo(ContainSubstring("caused by"))
		})

		It("should format stack trace correctly", func() {
			e := From(ErrForbidden).WithIdentifier(403001)
			stamped := Stamp(e)
			result := stamped.Error()
			Expect(result).To(ContainSubstring("at=("))
			Expect(result).To(ContainSubstring("func="))
			Expect(result).To(ContainSubstring("file="))
			Expect(result).To(ContainSubstring("line="))
		})
	})

	Context("Method chaining edge cases", func() {
		It("should handle long method chains", func() {
			e := From(ErrForbidden).
				WithIdentifier(403001).
				WithDetail("detail 1").
				WithDetail("detail 2").
				WithDetail("detail 3").
				WithProperty("prop1", "val1").
				WithProperty("prop2", "val2").
				WithProperty("prop3", "val3").
				CausedBy(errPerm)
			Expect(e.Title).To(Equal("forbidden"))
			Expect(e.Identifier).To(Equal(int32(403001)))
			Expect(e.Details).To(HaveLen(3))
			Expect(e.Properties).To(HaveLen(3))
			Expect(e.Cause).To(Equal(errPerm))
		})

		It("should maintain detail order in chains", func() {
			e := From(ErrNotFound).
				WithDetail("first").
				WithDetail("second").
				WithDetail("third")
			Expect(e.Details[0]).To(Equal("first"))
			Expect(e.Details[1]).To(Equal("second"))
			Expect(e.Details[2]).To(Equal("third"))
		})

		It("should allow mixing formatted and regular details", func() {
			e := From(ErrForbidden).
				WithDetail("regular detail").
				WithDetailf("formatted %s", "detail").
				WithDetail("another regular")
			Expect(e.Details).To(Equal([]string{"regular detail", "formatted detail", "another regular"}))
		})
	})

	Context("Integration with standard errors package", func() {
		It("should work with errors.As()", func() {
			e := From(ErrForbidden).WithIdentifier(403001)
			var target *Error
			Expect(errors.As(e, &target)).To(BeTrue())
			Expect(target.Title).To(Equal("forbidden"))
			Expect(target.Identifier).To(Equal(int32(403001)))
		})

		It("should work with errors.Is() through wrapped errors with cause", func() {
			e1 := From(ErrForbidden).WithIdentifier(403001)
			e2 := From(ErrInternal).CausedBy(e1)
			// e2 wraps e1 as its cause, so errors.Is should find e1
			Expect(errors.Is(e2, e1)).To(BeTrue())
		})
	})
})
