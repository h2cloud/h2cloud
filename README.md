# Metaless H2
The Golang version of Swift Layer-C, which has been renamed to Metaless H2

Dependencies:

- [github.com/ncw/swift](https://github.com/ncw/swift)
- [github.com/levythu/gurgling](https://github.com/levythu/gurgling)

**Attention: Currently the dependency github.com/ncw/swift is replaced by my forked
version: levythu/gurgling. As a result, you cannot use `introducedependencies.sh` to
install it but use this instead:** (With environment variables set using `setenv`)

```sh
mkdir -p $GOPATH/src/github.com/ncw/
cd $GOPATH/src/github.com/ncw/
git clone https://github.com/levythu/swift-1.git
mv swift-1/ swift/
```

**Remember to set environment variable SLCHOME to locate the homepath.**
