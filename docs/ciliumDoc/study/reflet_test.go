package study

import (
	"container/list"
	"fmt"
	"reflect"
	"testing"
)

type User struct {
	Name string
	Age  int
}

var user = User{Name: "john", Age: 39}

func TestBasic(t *testing.T) {

	userType := reflect.TypeOf(user)
	fmt.Printf("the user variable type is %v\n", userType)
	userValue := reflect.ValueOf(user)
	fmt.Printf("the user variable value is %v\n", userValue)

}

func TestInspectType(t *testing.T) {
	userType := reflect.TypeOf(user)
	fmt.Printf("the user variable type is %v and the name is %v\n", userType.Kind(), userType.Name())

	if userType.Kind() == reflect.Struct {
		fieldsNum := userType.NumField()
		for i := 0; i < fieldsNum; i++ {
			field := userType.Field(i)
			fmt.Printf("the field name is %v and type is %v \n ", field.Name, field.Type)
		}
	}
}

func TestInspectAndModifyValue(t *testing.T) {
	userValue := reflect.ValueOf(&user)
	fmt.Printf("the user variable type is %v and canSet %v\n", userValue.Kind(), userValue.CanSet())
	if userValue.Kind() == reflect.Ptr {
		userValue = userValue.Elem()
	}
	if !userValue.CanSet() {
		fmt.Println("variable can not be modified")
		return
	}
	if userValue.Kind() == reflect.Struct {

		fieldsNum := userValue.NumField()
		for i := 0; i < fieldsNum; i++ {
			field := userValue.Field(i)
			if field.Kind() == reflect.String && field.CanSet() {
				field.SetString("modified")
			}
			//fmt.Printf("the field name is %v and type is %v \n ", field.Name, field.Type)
		}
		fmt.Print(userValue)
	}
}

func funcInspecExample(name string, age int) User {
	return User{Name: name, Age: age}
}

func funcInspect(fn interface{}) {
	funcIns := reflect.TypeOf(fn)
	if funcIns.Kind() != reflect.Func {
		fmt.Println("not a func")
		return
	}
	fmt.Printf("num of input %v; num of output %v \n", funcIns.NumIn(), funcIns.NumOut())
	for i := 0; i < funcIns.NumIn(); i++ {
		param := funcIns.In(i)
		fmt.Printf("the input param is %v \n", param)
	}
	for i := 0; i < funcIns.NumOut(); i++ {
		param := funcIns.Out(i)
		fmt.Printf("the out param is %v \n", param)
	}

}

func callFunc(fn interface{}, args ...interface{}) []reflect.Value {
	funcVal := reflect.ValueOf(fn)
	fmt.Printf("the func value type is %v \n", funcVal.Type())
	callParams := make([]reflect.Value, len(args))
	for i := 0; i < len(args); i++ {
		callParams[i] = reflect.ValueOf(args[i])
	}
	return funcVal.Call(callParams)
}

func TestFuncInspect(t *testing.T) {
	funcInspect(funcInspecExample)

}

func TestFuncDynamicCallInspect(t *testing.T) {
	result := callFunc(funcInspecExample, "john", 3)
	fmt.Printf("the result is %v \n", result[0])
	fmt.Print(result[0].FieldByName("Name").String())

}

type Config struct {
	Host string `json:"host" validate:"required"`
	Port int    `json:"port" validate:"min=1,max=10"`
}

func inspectStructTags(in interface{}) {
	t := reflect.TypeOf(in)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		portTag := field.Tag.Get("validate")
		fmt.Printf("the field name is %v and tags are %v %v\n", field.Name, jsonTag, portTag)
	}
}

func TestInspectStructTags(t *testing.T) {
	input := Config{
		Host: "nyit.edu",
		Port: 80,
	}
	inspectStructTags(&input)
}

func createInstance(t reflect.Type) interface{} {
	newValue := reflect.New(t)
	return newValue.Interface()
}

func TestCreateInstance(t *testing.T) {
	newType := reflect.TypeOf((*User)(nil)).Elem()
	newInstance := createInstance(newType)
	fmt.Print(newInstance.(*User).Name, newInstance.(*User).Age)
}

func embedTypes(i interface{}, e reflect.Type) bool {
	if i == nil {
		return false
	}
	t, ok := i.(reflect.Type)
	if !ok {
		t = reflect.TypeOf(i)
	}

	types := list.New()
	types.PushBack(t)
	for types.Len() > 0 {
		t := types.Remove(types.Front()).(reflect.Type)
		if t == e{
			return  true
		}
		if t.Kind() != reflect.Struct{
			continue
		}
		for i := 0 ; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.Anonymous{
				types.PushBack(field.Anonymous)
			}
			
		}
	}
	return  false

}
