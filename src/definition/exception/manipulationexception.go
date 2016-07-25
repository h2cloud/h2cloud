package exception

import "errors"

var EX_UNMATCHED_MERGE=errors.New("exception.manipulate.merger.unmatch")
var EX_UNSUPPORTED_TYPESTAMP=errors.New("exception.manipulate.unsupported_timestamp")
var EX_INCONSISTENT_TYPE=errors.New("exception.manipulate.inconsistent_type")

var EX_FAIL_TO_FETCH_INTRALINK=errors.New("exception.manipulate.fail_to_fetch_intralink")
var EX_METADATA_NEEDS_TO_BE_SPECIFIED=errors.New("exception.manipulate.meta_need_specify")
