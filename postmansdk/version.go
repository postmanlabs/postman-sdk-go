package postmansdk

import (
	"fmt"
	"runtime/debug"
)

const SDK_VERSION = "0.0.1"

func PrintVersion() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Println("Version information not available")
		return
	}
	fmt.Printf("Version: %s\n", info.Main.Version)

	// for _, dep := range info.Deps {
	// 	fmt.Printf("Dep: %+v\n", dep)
	// }
}
