package main

import "fmt"

type Student struct {
	Name string
	Age  int
}

type Student2 struct {
	S1 Student
	Number int
}

func printf(students ...Student) {
	for _, value := range students {
		fmt.Println(value.Name)
	}
}

func main() {
	var student1 Student//声明一个结构体，里面的变量是会被初始化的
	var student2 Student2//嵌套的结构体也会被初始化
	students := []Student{{"lpk", 12}, {"ty", 13}}
	printf(students...)
	fmt.Println("test",student1.Age)
	fmt.Println("test2",student2.S1.Age)
	var a= [][]byte{}

	fmt.Printf("a point to %p\n", &a)
	b := [][]byte{[]byte("456"),[]byte("789")}
	fmt.Printf("b point to %p\n", &b)
	fmt.Printf("b[0] point to %p\n", &b[0])
	a=b//里面的各个元素复制了一下地址 a[0] b[0] 在栈里 存下同样的堆地址
	fmt.Printf("a point to %p\n", &a)
	fmt.Printf("a[0] point to %p\n", &a[0])
	fmt.Printf("b point to %p\n", &b)
	fmt.Printf("b[0] point to %p\n", &b[0])
	//slice地址不是首元素地址，它指向堆中的一个结构体，结构体里有两个元素，首元素地址，和slice长度

	c := b[0]
	fmt.Printf("c pointer to %p\n",&c)
	fmt.Println(b[0])
	c=[]byte("123")
	fmt.Println(b[0])


}
