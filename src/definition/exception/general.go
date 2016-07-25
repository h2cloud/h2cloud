package exception

import "errors"

var (
    LOGICAL_ERROR=errors.New("Logical error happens and may be very severe: please commit the debugger logs to the author of the code.")
)
