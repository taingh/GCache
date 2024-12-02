package gcache

// ByteView 结构体持有一个不可变的字节视图。
type ByteView struct {
	b []byte
}

// Len 方法返回视图的长度。
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice 方法返回数据的一个副本作为字节切片。
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// String 方法将数据作为字符串返回，必要时会创建一个副本。
func (v ByteView) String() string {
	return string(v.b)
}

// cloneBytes 函数创建并返回给定字节切片的一个副本。
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}