package numerical

//using recursion O(2 power n)
func fibo(num int) int {
	if num <= 1 {
		return num
	}
	return fibo(num-1) + fibo(num-2)
}

func fibo1001(num int) int {
	if num <= 1 {
		return num
	}
	return fibo1001(num - 1) + fibo1001(num - 2)
}
// O(n)
func fiboIter0521(num int) int {
	if num <= 1 {
		return num
	}
	res, first,second :=0, 0,1
	for num > 1 {
		res = (first+second)
		first = second
		second = res
		num--
	}
	return res
}

func fiboIter1001(num int) int {
	res, first, second := 0,0,1
	for num > 1 {
		res = first + second
		first = second
		second = res
		num--
	}
	return res
}

// num = 5
// 0,1
// res = 1 first = 1, secod = 1 num-- => 4
// res = 2 first = 1, second = 2 num-- => 3
// res = 3 first = 2, second = 3 num-- => 2
// res = 5 first = 3, second = 5 num-- => 1
// 0 1 1 2 3 5 8 13 21 34 55
// return 8

func fiboDpGetSeq0521(num int) []int {
	if num < 0 {
		panic("argument shall be above 0")
	}
	res := make([]int, num+1)
	res[0] = 0
	if num >= 1 {
		res[1] = 1
	}
	if num <= 1 {
		return res
	}
	for i:=2; i <= num; i++{
		res[i] = (res[i-1]+res[i-2])
	}
	return res
}