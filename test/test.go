package test

type TestStruct struct {
	Field1 []string
	Wooh *string
}

type TestStruct2 struct {
	TestStruct

	Field2, Hello []int
}

// TODO(patrik): This is not working

// type TestStruct struct {
// 	Field1 string
// 	Wooh string
// }
//
// type TestStruct3 struct {
// 	TestStruct
// }
//
// type TestStruct2 struct {
// 	TestStruct3
//
// 	Field2, Hello int
// }
