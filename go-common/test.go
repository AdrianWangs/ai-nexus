package main

import (
	"fmt"
	"github.com/AdrianWangs/ai-nexus/go-common/nacos"
	"reflect"
	"runtime"
)

// printFunctionName prints the name of the function passed as an argument.
func printFunctionName(i interface{}) {
	f := reflect.ValueOf(i).Pointer()
	fn := runtime.FuncForPC(f)

	if fn == nil {
		fmt.Println("Function name not found")
		return
	}
	fmt.Println("Function name:", fn.Name())
}

// printFunctionPackage prints the package of the function passed as an argument.
func printFunctionPackage() {
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

	if fn == nil {
		fmt.Println("Function package not found")
		return
	}
	fmt.Println("Function package:", fn.Name())
}

func exampleFunction() {
	// This is an example function.
}

func main() {
	printFunctionName(nacos.GetNacosRegistry)

	//注册一个 rpc 客户端

}
