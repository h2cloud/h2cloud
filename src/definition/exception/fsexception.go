package exception

import "errors"

var EX_INVALID_FILENAME=errors.New("exception.fs.lookup.invalid_filename")
var EX_INODE_NONEXIST=errors.New("exception.fs.lookup.inode_not_exist")

var EX_FAIL_TO_LOOKUP=errors.New("exception.fs.lookup.fail")

var EX_FOLDER_ALREADY_EXIST=errors.New("exception.fs.mkdir.folder_already_exist")

var EX_FILE_NOT_EXIST=errors.New("exception.fs.streamio.file_not_exist")

var EX_TRASHBOX_NOT_INITED=errors.New("exception.fs.trashbox_not_initialized")
