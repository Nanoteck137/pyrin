package test

type TestStruct struct {
	Field1 string
	Wooh string
}

type TestStruct3 struct {
	TestStruct

	Boop int
}

type TestStruct2 struct {
	TestStruct3

	Field2, Hello int
}
