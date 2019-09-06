package utils

type DataType interface{}
type Unique interface {
	Sort(dat DataType) []DataType
	UniquieSlice(dat []DataType) []DataType
	Len(dat []DataType) int
	Append(dat []DataType, addition []DataType) []DataType
}
