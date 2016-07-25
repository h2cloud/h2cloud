package public

import (
    "runtime"
    "definition/configinfo"
    "strconv"
    . "logger"
)

func prepEnv_SetConcurrency() {
    num:=configinfo.THREAD_UTILISED
    if (num<=0) {
        num=runtime.NumCPU()
    }
    runtime.GOMAXPROCS(num)
    Secretary.Log("mainpkg::prepEnv_SetConcurrency", "Set GOMAXPROCS to "+strconv.Itoa(runtime.GOMAXPROCS(-1)))
}
// Only run once when start.
func StartUp() {
    Secretary.Log("mainpkg::StartUp", "Midware-MH2 is starting...")
    prepEnv_SetConcurrency()
    Secretary.Log("mainpkg::StartUp", "Premise checked. Now lauching Web server...")
}
