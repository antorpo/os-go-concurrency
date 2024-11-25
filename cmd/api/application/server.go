package application

import (
	"fmt"
	"runtime"
)

func StartApp() (*Application, error) {
	maxCPUs()

	application, err := BuildApplication()
	if err != nil {
		return nil, err
	}

	RegisterRoutes(application)

	return application, nil
}

// TODO: Disable concurrency
func maxCPUs() {
	cpu := runtime.NumCPU() + 1
	_ = runtime.GOMAXPROCS(cpu)
	printable := fmt.Sprintf(" ⚡ using %d go max processes\n", cpu)
	fmt.Println(printable)
}
