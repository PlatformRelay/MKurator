package mqpcf

import "fmt"

func errNotImplemented(method string) error {
	return fmt.Errorf("mqpcf: %s not implemented", method)
}
