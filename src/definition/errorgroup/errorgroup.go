package errorgroup

// An error that can hold multiple error prompt

type ErrorAssembly struct {
    errorList []error
}

func AddIn(one *ErrorAssembly, newOne error) *ErrorAssembly {
    if one==nil {
        return &ErrorAssembly {
            errorList: []error{newOne},
        }
    }
    one.errorList=append(one.errorList, newOne)
    return one
}

func (this *ErrorAssembly)Exist(obj error) bool {
    if this==nil {
        return false
    }
    for _, e:=range this.errorList {
        if e==obj {
            return true
        }
    }
    return false
}

func (this *ErrorAssembly)Error() string {
    if this==nil {
        return "nil"
    }
    var result=""
    for i, e:=range this.errorList {
        if i==0 {
            result+="["
        } else {
            result+=","
        }
        result+=" "+e.Error()+" "
    }
    result+="]"
    return result
}

func In(container error, obj error) bool {
    if container==obj {
        return true
    }
    var des, _=container.(*ErrorAssembly)
    return des.Exist(obj)
}
func Nil(obj error) bool {
    switch obj:=obj.(type) {
	case *ErrorAssembly:
		return obj==nil
	default:
		return obj==nil
	}
}
