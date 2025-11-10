package xerr

import (
	"slices"
	"strings"
)

type MultiErr struct {
	Errors []error
	Msg    string
	Indent string
}

func (err MultiErr) Unwrap() []error { return err.Errors }

func (err MultiErr) Error() string {
	if err.Msg == "" && len(err.Errors) == 1 {
		return err.Errors[0].Error()
	}

	if err.Msg == "" {
		err.Msg = "errors"
	}
	switch len(err.Errors) {
	case 0:
		return err.Msg

	case 1:
		return err.Msg + ": " + err.Errors[0].Error()

	default:
		if err.Indent == "" {
			err.Indent = "  "
		}

		var builder strings.Builder

		builder.WriteString(err.Msg + ":")
		for _, e := range err.Errors {
			builder.WriteString("\n" + indent("- "+e.Error(), err.Indent))
		}

		return builder.String()
	}
}

func indent(value, indent string) string {
	return indent + strings.ReplaceAll(value, "\n", "\n"+indent)
}

func MultiErrWithIndentFrom(msg, indent string, errs ...error) error {
	var nonNilErrs []error
	for _, err := range errs {
		if err != nil {
			nonNilErrs = append(nonNilErrs, err)
		}
	}
	if len(nonNilErrs) == 0 {
		return nil
	}
	return MultiErr{
		Errors: nonNilErrs,
		Msg:    msg,
		Indent: indent,
	}
}

func Join(errs ...error) error {
	return MultiErrFrom("", errs...)
}

func JoinOrdered(errs ...error) error {
	return MultiErrOrderedFrom("", errs...)
}

func MultiErrFrom(msg string, errs ...error) error {
	return MultiErrWithIndentFrom(msg, "  ", errs...)
}

func MultiErrOrderedFrom(msg string, errs ...error) error {
	copy := append([]error{}, errs...)
	slices.SortFunc(copy, func(a, b error) int {
		if a == nil || b == nil {
			return 0
		}
		return strings.Compare(a.Error(), b.Error())
	})
	return MultiErrFrom(msg, copy...)
}
