/* 指针
1、题目 ：编写一个Go程序，定义一个函数，该函数接收一个整数指针作为参数，在函数内部将该指针指向的值增加10，然后在主函数中调用该函数并输出修改后的值。
考察点 ：指针的使用、值传递与引用传递的区别。
2、题目 ：实现一个函数，接收一个整数切片的指针，将切片中的每个元素乘以2。
考察点 ：指针运算、切片操作。 */

package main

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

/*
1、题目 ：编写一个Go程序，定义一个函数，该函数接收一个整数指针作为参数，在函数内部将该指针指向的值增加10，然后在主函数中调用该函数并输出修改后的值。
考察点 ：指针的使用、值传递与引用传递的区别。
*/
func addTen(x *int) int {
	*x += 10
	return *x
}

/* 2、题目 ：实现一个函数，接收一个整数切片的指针，将切片中的每个元素乘以2。 */

func multTwo(slice *[]int) []int {
	for i := range len(*slice) {
		(*slice)[i] *= 2
	}
	// fmt.Println(*slice)
	return *slice
}

/*
	## :white_check_mark:Goroutine

1. 题目 ：编写一个程序，使用 go 关键字启动两个协程，一个协程打印从1到10的奇数，另一个协程打印从2到10的偶数。
  - 考察点 ： go 关键字的使用、协程的并发执行。
*/
func basicGoroutine() {
	go printOddNumber()
	time.Sleep(time.Second)
	go printEvenNumber()
	time.Sleep(time.Second)
	fmt.Println()

}
func printOddNumber() {
	for i := 1; i <= 10; i++ {
		if i%2 == 0 {
			fmt.Println(i)
		}
	}
}
func printEvenNumber() {
	for i := 1; i < 10; i++ {
		if i%2 != 0 {
			fmt.Println(i)
		}
	}
}

/* 2. 题目 ：设计一个任务调度器，接收一组任务（可以用函数表示），并使用协程并发执行这些任务，同时统计每个任务的执行时间。
   - 考察点 ：协程原理、并发任务调度。 */

type Task struct {
	Name string
	Run  func()
}

type TaskResult struct {
	Name     string
	Duration time.Duration
}

func Scheduler(tasks []Task) []TaskResult {
	var wg sync.WaitGroup
	resultsChan := make(chan TaskResult, len(tasks))
	for _, task := range tasks {
		wg.Add(1)
		go func(t Task) {
			defer wg.Done()
			start := time.Now()
			t.Run()
			duration := time.Since(start)
			resultsChan <- TaskResult{
				Name:     t.Name,
				Duration: duration,
			}
		}(task)
	}
	go func() {
		wg.Wait()
		close(resultsChan)
	}()
	results := make([]TaskResult, 0, len(tasks))
	for result := range resultsChan {
		results = append(results, result)
	}
	return results
}

/*
	## :white_check_mark:面向对象

1. 题目 ：定义一个 Shape 接口，包含 Area() 和 Perimeter() 两个方法。然后创建 Rectangle 和 Circle 结构体，实现 Shape 接口。在主函数中，创建这两个结构体的实例，并调用它们的 Area() 和 Perimeter() 方法。
  - 考察点 ：接口的定义与实现、面向对象编程风格。
*/
type Shape interface {
	Area()
	Perimeter()
}
type Rectangle struct {
	i int
	j int
	k int
	l int
}
type Circle struct {
	r int
}

func (r *Rectangle) Area() {
	fmt.Printf("\n三角形面积：%d", (*r).i*(*r).l/2)
}
func (r *Rectangle) Perimeter() {
	fmt.Printf("\n三角形周长：%d", (*r).i+(*r).j+(*r).k)
}
func (r Circle) Area() {
	fmt.Printf("\n圆面积：%f", math.Pi*float64(r.r)*float64(r.r))
}
func (r Circle) Perimeter() {
	fmt.Printf("\n圆周长：%f", 2.0*math.Pi*float64(r.r))
}
func mix(shape Shape) {
	shape.Area()
	shape.Perimeter()
}

/*
2. 题目 ：使用组合的方式创建一个 Person 结构体，包含 Name 和 Age 字段，再创建一个 Employee 结构体，组合 Person 结构体并添加 EmployeeID 字段。为 Employee 结构体实现一个 PrintInfo() 方法，输出员工的信息。
  - 考察点 ：组合的使用、方法接收者。
*/
type Person struct {
	Name string
	Age  int
}

type Employee struct {
	EmployeeID int
	Person
}

func PrintInfo(e Employee) {
	fmt.Printf("\n名字：%s，年龄：%d，员工号：%d", e.Name, e.Age, e.EmployeeID)
}

/*
	## :white_check_mark:Channel

1. 题目 ：编写一个程序，使用通道实现两个协程之间的通信。一个协程生成从1到10的整数，并将这些整数发送到通道中，另一个协程从通道中接收这些整数并打印出来。
  - 考察点 ：通道的基本使用、协程间通信。
*/
func demo1() {
	m := make(chan int)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer close(m)
		for i := 1; i <= 10; i++ {
			m <- i
		}
		// m<-generateTen()
	}()
	go func() {
		defer wg.Done()
		for value := range m {
			fmt.Printf("\n收到 %d", value)
		}

	}()
	wg.Wait()

}

// func generateTen(){
// 	for i:=1;i<=10;i++{
// 		println(i)
// 	}
// 	// wg.sync.WaitGroup
// }

/*
2. 题目 ：实现一个带有缓冲的通道，生产者协程向通道中发送100个整数，消费者协程从通道中接收这些整数并打印。
  - 考察点 ：通道的缓冲机制。
*/
func demo2() {
	d := make(chan int, 100)
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		defer close(d)
		for i := 1; i <= 100; i++ {
			d <- i

		}
	}()

	go func() {
		defer wg.Done()
		for k := range d {
			fmt.Printf("\n缓冲通道收到 %d", k)
		}
		// fmt.Printf("\n缓冲通道收到 %d", d)

	}()
	wg.Wait()
}

/*
## :white_check_mark:锁机制
1. 题目 ：编写一个程序，使用 sync.Mutex 来保护一个共享的计数器。启动10个协程，每个协程对计数器进行1000次递增操作，最后输出计数器的值。
  - 考察点 ： sync.Mutex 的使用、并发数据安全。
*/
var (
	wg sync.WaitGroup
	mu sync.Mutex
	k  int
)

func demo3() int {
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 1; j <= 1000; j++ {
				mu.Lock()
				k++
				mu.Unlock()
			}

		}()

		// time.Sleep(500 * time.Millisecond)

	}
	wg.Wait()
	return k
}

/*
2. 题目 ：使用原子操作（ sync/atomic 包）实现一个无锁的计数器。启动10个协程，每个协程对计数器进行1000次递增操作，最后输出计数器的值。
  - 考察点 ：原子操作、并发数据安全。
*/
func demo4() int64 {
	var (
		wg      sync.WaitGroup
		counter atomic.Int64
	)
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 1; j <= 1000; j++ {
				counter.Add(1)
			}

		}()

	}
	wg.Wait()
	return counter.Load()
}

func main() {
	// 指针部分 1、编写一个Go程序，定义一个函数，该函数接收一个整数指针作为参数，在函数内部将该指针指向的值增加10，然后在主函数中调用该函数并输出修改后的值
	p := 15
	fmt.Printf("%d加10等于%d\n", p, addTen(&p))
	fmt.Println(p)
	// 指针部分 2、实现一个函数，接收一个整数切片的指针，将切片中的每个元素乘以2。
	slice := []int{1, 2, 3, 4, 5}
	fmt.Println(multTwo(&slice)) //输出[2,4,6,8,10]
	// Goroutine 1、实现一个函数，接收一个整数切片的指针，将切片中的每个元素乘以2
	basicGoroutine()
	// Goroutine 2、设计一个任务调度器，接收一组任务（可以用函数表示），并使用协程并发执行这些任务，同时统计每个任务的执行时间。
	tasks := []Task{
		{
			Name: "任务1",
			Run: func() {
				time.Sleep(1 * time.Second)
				fmt.Println("任务1执行完成！")
			},
		},
		{
			Name: "任务2",
			Run: func() {
				time.Sleep(500 * time.Millisecond)
				fmt.Println("任务2执行完成！")
			},
		},
		{
			Name: "任务3",
			Run: func() {
				time.Sleep(1500 * time.Millisecond)
				fmt.Println("任务3执行完成！")
			},
		},
	}
	totalStart := time.Now()
	results := Scheduler(tasks)
	fmt.Println("\n任务执行统计：")
	for _, result := range results {
		fmt.Printf("%s 耗时 %v\n", result.Name, result.Duration)
	}
	fmt.Printf("总耗时：%v", time.Since(totalStart))

	r := Rectangle{i: 1, j: 2, k: 3, l: 4}
	c := Circle{r: 5}
	mix(&r)
	mix(c)

	e := Employee{Person: Person{Name: "bob", Age: 20}, EmployeeID: 5}
	PrintInfo(e)

	demo1()                             //1. 题目 ：编写一个程序，使用通道实现两个协程之间的通信。一个协程生成从1到10的整数，并将这些整数发送到通道中，另一个协程从通道中接收这些整数并打印出来。
	demo2()                             //2. 题目 ：实现一个带有缓冲的通道，生产者协程向通道中发送100个整数，消费者协程从通道中接收这些整数并打印。
	fmt.Printf("\ndemo3的值：%d", demo3()) //1. 题目 ：编写一个程序，使用 sync.Mutex 来保护一个共享的计数器。启动10个协程，每个协程对计数器进行1000次递增操作，最后输出计数器的值。
	fmt.Printf("\ndemo4的值：%d", demo4()) //2. 题目 ：使用原子操作（ sync/atomic 包）实现一个无锁的计数器。启动10个协程，每个协程对计数器进行1000次递增操作，最后输出计数器的值。

}
