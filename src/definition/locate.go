package definition

import "path/filepath"
import "os"

var base=os.Getenv("SLCHOME")

func GetABSPath(relpath string) (string, error) {
    if base!="" {
        return filepath.Join(base,relpath), nil
    }
    return filepath.Abs(relpath)
}
