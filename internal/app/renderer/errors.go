package renderer

import "io"

// Error writes the error message writer in red color.
func Error(err error, w io.Writer) error {
	_, wErr := red.Fprintln(w, err.Error())
	return wErr
}
