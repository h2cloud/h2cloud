package exception

import "errors"

var EX_WRONG_FILEFORMAT=errors.New("exception.io.wrong_format")
var EX_IMPROPER_DATA=errors.New("exception.io.improper_data")
var EX_IO_ERROR=errors.New("exception.io.error")
var EX_INDEX_ERROR=errors.New("exception.io.put_but_index_not_established")
var EX_CONCURRENT_CHAOS=errors.New("exception.io.concurrent_chaos")
