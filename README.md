# ring-writer
一个基于 Go 实现的环形（Ring）Writer，核心特性是当写入数据的总容量超过预设字节限制后，新写入的数据会从头覆盖旧数据。适用于循环存储有限容量日志、临时数据等场景。

## 安装
```go
go get github.com/Li-giegie/ring-writer
```

## 使用

```go
package ring_writer

import (
	"fmt"
	"os"
	"testing"
)

func TestRingWriter(t *testing.T) {
	file, err := os.OpenFile("./1.txt", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	writer, err := New(file, 5) // limitSize 5 byte
	if err != nil {
		t.Error(err)
		return
	}
	defer writer.SaveOffset()
	fmt.Println("offset:", writer.Offset())
	fmt.Println(writer.Write([]byte("1234567890")))
	fmt.Println("offset:", writer.Offset())
}
```