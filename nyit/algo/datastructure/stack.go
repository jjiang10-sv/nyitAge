package main

import (
	"fmt"
	"strconv"
	"strings"
)

// Stack represents a stack data structure
type Stack struct {
	items []interface{}
}

// Push adds an item to the top of the stack
func (s *Stack) Push(item interface{}) {
	s.items = append(s.items, item)
}

// Pop removes and returns the top item from the stack
func (s *Stack) Pop() interface{} {
	if len(s.items) == 0 {
		panic("Stack is empty")
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item
}

// IsEmpty checks if the stack is empty
func (s *Stack) IsEmpty() bool {
	return len(s.items) == 0
}

// EvaluatePostfix evaluates a postfix expression using a stack
func EvaluatePostfix(expression string) int {
	stack := Stack{}
	tokens := strings.Fields(expression) // Split the expression into tokens

	for _, token := range tokens {
		// If the token is a number, push it onto the stack
		if num, err := strconv.Atoi(token); err == nil {
			stack.Push(num)
		} else {
			// If the token is an operator, pop two operands and apply the operator
			operand2 := stack.Pop().(int)
			operand1 := stack.Pop().(int)

			switch token {
			case "+":
				stack.Push((operand1 + operand2))
			case "-":
				stack.Push(operand1 - operand2)
			case "*":
				stack.Push(operand1 * operand2)
			case "/":
				stack.Push(operand1 / operand2)
			default:
				panic(fmt.Sprintf("Unknown operator: %s", token))
			}
		}
	}

	// The final result is the only item left on the stack
	if stack.IsEmpty() {
		panic("Invalid expression")
	}
	return stack.Pop().(int)
}

func PostfixMain() {
	// Example postfix expression: "3 4 + 2 *" (equivalent to (3 + 4) * 2)
	expression := "3 4 + 2 *"
	result := EvaluatePostfix(expression)
	fmt.Printf("Result of '%s' is: %d\n", expression, result)

	// Another example: "5 1 2 + 4 * + 3 -" (equivalent to 5 + (1 + 2) * 4 - 3)
	expression = "5 1 2 + 4 * + 3 -"
	result = EvaluatePostfix(expression)
	fmt.Printf("Result of '%s' is: %d\n", expression, result)
}

// InfixToPostfix converts an infix expression to postfix notation
func InfixToPostfix(expression string) string {
	stack := Stack{}
	output := []string{}
	precedence := map[string]int{"+": 1, "-": 1, "*": 2, "/": 2}

	tokens := strings.Fields(expression)
	for _, token := range tokens {
		if num, err := strconv.Atoi(token); err == nil {
			// If the token is a number, add it to the output
			output = append(output, strconv.Itoa(num))
		} else if token == "(" {
			// If the token is '(', push it onto the stack
			stack.Push(token)
		} else if token == ")" {
			// If the token is ')', pop from the stack until '(' is found
			for !stack.IsEmpty() && stack.items[len(stack.items)-1] != "(" {
				output = append(output, stack.Pop().(string))
			}
			stack.Pop() // Pop the '(' from the stack
		} else {
			// If the token is an operator, pop from the stack while precedence is higher
			for !stack.IsEmpty() && precedence[stack.items[len(stack.items)-1].(string)] >= precedence[token] {
				output = append(output, stack.Pop().(string))
			}
			stack.Push(token)
		}
	}

	// Pop any remaining operators from the stack
	for !stack.IsEmpty() {
		output = append(output, stack.Pop().(string))
	}

	return strings.Join(output, " ")
}

func postfixToInfixmain() {
	// Example infix expression: "3 + 4 * 2 / ( 1 - 5 )"
	infixExpression := "3 + 4 * 2 / ( 1 - 5 )"
	postfixExpression := InfixToPostfix(infixExpression)
	fmt.Printf("Infix Expression: %s\n", infixExpression)
	fmt.Printf("Postfix Expression: %s\n", postfixExpression)

	// Evaluate the postfix expression
	result := EvaluatePostfix(postfixExpression)
	fmt.Printf("Result: %d\n", result)
}

// TreeNode represents a node in the binary expression tree
type TreeNode struct {
	value string
	left  *TreeNode
	right *TreeNode
}

// isOperator checks if a token is an operator
func isOperator(token string) bool {
	return token == "+" || token == "-" || token == "*" || token == "/"
}

// precedence returns the precedence of an operator
func precedence(token string) int {
	switch token {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	default:
		return 0
	}
}

// buildExpressionTree builds a binary expression tree from an infix expression
func buildExpressionTree(expression string) *TreeNode {
	// Convert the expression into a slice of tokens
	tokens := strings.Fields(expression)

	// Stack for operands
	operandStack := []*TreeNode{}

	// Stack for operators
	operatorStack := []string{}

	for _, token := range tokens {
		if _, err := strconv.Atoi(token); err == nil {
			// If the token is a number, push it onto the operand stack
			operandStack = append(operandStack, &TreeNode{value: token})
		} else if token == "(" {
			// If the token is '(', push it onto the operator stack
			operatorStack = append(operatorStack, token)
		} else if token == ")" {
			// If the token is ')', pop from the operator stack until '(' is found
			for len(operatorStack) > 0 && operatorStack[len(operatorStack)-1] != "(" {
				// Pop an operator and two operands, then build a subtree
				operator := operatorStack[len(operatorStack)-1]
				operatorStack = operatorStack[:len(operatorStack)-1]

				right := operandStack[len(operandStack)-1]
				operandStack = operandStack[:len(operandStack)-1]

				left := operandStack[len(operandStack)-1]
				operandStack = operandStack[:len(operandStack)-1]

				// Create a new subtree and push it onto the operand stack
				operandStack = append(operandStack, &TreeNode{value: operator, left: left, right: right})
			}
			// Pop the '(' from the operator stack
			operatorStack = operatorStack[:len(operatorStack)-1]
		} else if isOperator(token) {
			// If the token is an operator, pop from the operator stack while precedence is higher
			for len(operatorStack) > 0 && precedence(operatorStack[len(operatorStack)-1]) >= precedence(token) {
				operator := operatorStack[len(operatorStack)-1]
				operatorStack = operatorStack[:len(operatorStack)-1]

				right := operandStack[len(operandStack)-1]
				operandStack = operandStack[:len(operandStack)-1]

				left := operandStack[len(operandStack)-1]
				operandStack = operandStack[:len(operandStack)-1]

				// Create a new subtree and push it onto the operand stack
				operandStack = append(operandStack, &TreeNode{value: operator, left: left, right: right})
			}
			// Push the current operator onto the operator stack
			operatorStack = append(operatorStack, token)
		}
	}

	// Pop any remaining operators from the stack
	for len(operatorStack) > 0 {
		operator := operatorStack[len(operatorStack)-1]
		operatorStack = operatorStack[:len(operatorStack)-1]

		right := operandStack[len(operandStack)-1]
		operandStack = operandStack[:len(operandStack)-1]

		left := operandStack[len(operandStack)-1]
		operandStack = operandStack[:len(operandStack)-1]

		// Create a new subtree and push it onto the operand stack
		operandStack = append(operandStack, &TreeNode{value: operator, left: left, right: right})
	}

	// The final tree is the only item left on the operand stack
	return operandStack[0]
}

// postorderTraversal performs a postorder traversal of the tree to generate the postfix expression
func postorderTraversal(root *TreeNode) string {
	if root == nil {
		return ""
	}

	left := postorderTraversal(root.left)
	right := postorderTraversal(root.right)

	// Concatenate the left, right, and root values
	return strings.TrimSpace(left + " " + right + " " + root.value)
}

func expressionTreeMain() {
	// Example infix expression: "3 + 4 * 2 / ( 1 - 5 )"
	infixExpression := "3 + 4 * 2 / ( 1 - 5 )"

	// Build the expression tree
	root := buildExpressionTree(infixExpression)

	// Generate the postfix expression using postorder traversal
	postfixExpression := postorderTraversal(root)
	fmt.Printf("Infix Expression: %s\n", infixExpression)
	fmt.Printf("Postfix Expression: %s\n", postfixExpression)
}
