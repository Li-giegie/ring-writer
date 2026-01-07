package ring_writer

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

func TestRingWriter(t *testing.T) {
	// 打开日志文件
	file, err := os.OpenFile("./1.txt", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	// 创建RingWriter 用于循环覆盖写入文件 长度为 5 字节
	rw, err := New(file, 5) // limitSize 5 byte
	if err != nil {
		t.Error(err)
		return
	}
	// 保存写入偏移量，进程重启后找到下次写入的offset位置
	defer rw.SaveOffset()
	// 输出offset位置
	fmt.Println("offset:", rw.Offset())

	// 创建缓冲区，减少写入次数
	bw := bufio.NewWriter(rw)
	defer bw.Flush()

	_, err = bw.WriteString("hello world")
	if err != nil {
		t.Error(err)
	}
}
