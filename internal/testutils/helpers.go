// code-from-spec: SPEC/golang/test/utils/helpers@tgVx1GTgCuTwF7xdqV5JUhyvboo
package testutils

func Ptr[T any](v T) *T { return &v }
