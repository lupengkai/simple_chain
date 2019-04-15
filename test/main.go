package main

import "fmt"

type Student struct {
	Name string
	Age  int
}

func printf(students ...Student) {
	for _, value := range students {
		fmt.Println(value.Name)
	}
}

func main() {
	students := []Student{{"lpk", 12}, {"ty", 13}}
	printf(students...)

}
