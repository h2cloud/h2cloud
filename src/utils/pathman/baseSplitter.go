package pathman

// Eliminates trailing slashes and splits last file-section with the previous
// returns (base, file)
func SplitPath(path string) (string, string) {
    var trimer=path
    var i int
    for i=len(trimer)-1; i>=0; i-- {
        if trimer[i]!='/' {
            break
        }
    }
    if i<0 {
        // consists all of slashes
        return "", ""
    }
    trimer=trimer[:i+1]

    // Now trimmer eleminates all the trailing slashes

    var j int
    for j=i; j>=0; j-- {
        if trimer[j]=='/' {
            break
        }
    }
    var base=trimer[:j+1]
    trimer=trimer[j+1:]
    // now trimer holds the last foldername
    // base holds the parent folder path

    return base, trimer
}
