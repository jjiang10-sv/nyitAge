/* queue_test.go - Test for generic queue implementation
 *
 * This file is part of ds library.
 *
 * The MIT License (MIT)
 * Copyright (c) <2016> Alexander Kuleshov <kuleshovmail@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software, and to permit persons to whom the Software is furnished to do
 * so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED,
 * INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
 * PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
 * HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF
 * CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
 * OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package queue

import (
	"fmt"
	"testing"
)

func TestQueue(t *testing.T) {
	var queue *Queue = New()

	queue.Enqueue(1)
	queue.Enqueue(2)
	queue.Enqueue(3)
	queue.Enqueue(4)
	queue.Enqueue(5)

	for i := 1; i < 6; i++ {
		item := queue.Dequeue()

		if item != i {
			t.Error("TestQueue failed...", i)
		}
	}
}

func TestQueue1(t *testing.T) {
	var q1 *Queue1 = New1()
	arr1 := [6]int{1, 2, 3, 4, 5, 6}
	for _, item := range arr1 {
		q1.Enqueue(item)
	}

	for i := 1; i <= len(arr1); i++ {
		item := q1.Dequeue()
		fmt.Println(item, i)

		if item != i {
			t.Error("TestQueue1 failed...", i)
		}
	}
}
