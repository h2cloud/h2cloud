package pathman

import (
    "regexp"
)

var shortcutRegexp, _=regexp.Compile(`\[\[SC\](.+)\](.*)`)

// Resolve parameter [[SC]inodename]{followingPath} to (inodename, followingPath)
// if no shortcut exists, inodename will be empty
func ShortcutResolver(scpath string) (string, string) {
    var result=shortcutRegexp.FindStringSubmatch(scpath)
    if len(result)>0 {
        return result[1], result[2]
    }
    return "", scpath
}
