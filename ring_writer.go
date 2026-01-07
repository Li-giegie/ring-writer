package ring_writer

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type Writer interface {
	io.Seeker
	io.Writer
	io.WriterAt
	io.Reader
}

// New 指定容量创建环形Writer，limitSize 必须大于 0（如果打开一个历史文件offset不能大于limitSize）
// 当累计写入数据容量大于限制容量（limitSize）时重新覆盖写入
// 注意并发写入需要加锁
func New(w Writer, limitSize int) (r *RingWriter, err error) {
	if _, err = w.Seek(0, 0); err != nil {
		return
	}
	b := make([]byte, 8)
	if n, rErr := w.Read(b); rErr != nil {
		if rErr != io.EOF {
			return nil, rErr
		}
		if n != 0 {
			return nil, errors.New("invalid Writer offset")
		}
		_, err = w.Seek(8, 0)
		return &RingWriter{writer: w, limitSize: limitSize}, err
	}
	offset := binary.LittleEndian.Uint64(b)
	if offset > uint64(limitSize) {
		return nil, fmt.Errorf("open failure offset %d is greater than the limit size %d", offset, limitSize)
	}
	_, err = w.Seek(int64(offset+8), 0)
	return &RingWriter{writer: w, offset: int(offset), limitSize: limitSize}, err
}

// RingWriter 环形Writer实现了：Write数据总容量超过限制大小后从头覆盖写入，并发需要加锁
type RingWriter struct {
	writer    Writer
	offset    int
	limitSize int
}

func (f *RingWriter) Write(p []byte) (n int, err error) {
	defer func() {
		if err != nil {
			f.writer.Seek(int64(f.offset+8), 0)
		}
	}()
	freeSize := f.limitSize - f.offset
	if len(p) <= freeSize {
		if n, err = f.writer.Write(p); err != nil {
			return 0, err
		}
		f.offset += len(p)
		return
	}
	if n = len(p) - f.limitSize; n > 0 {
		if _, err = f.writer.Write(p[n : n+freeSize]); err != nil {
			return 0, err
		}
		if _, err = f.writer.Seek(8, 0); err != nil {
			return 0, err
		}
		if _, err = f.writer.Write(p[n+freeSize:]); err != nil {
			return 0, err
		}
		f.offset = f.limitSize - freeSize
		n = f.limitSize
	} else {
		if _, err = f.writer.Write(p[:freeSize]); err != nil {
			return 0, err
		}
		if _, err = f.writer.Seek(8, 0); err != nil {
			return 0, err
		}
		if _, err = f.writer.Write(p[freeSize:]); err != nil {
			return 0, err
		}
		f.offset = len(p) - freeSize
		n = len(p)
	}
	return
}

// SaveOffset 持久化offset到文件
func (f *RingWriter) SaveOffset() (err error) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(f.offset))
	_, err = f.writer.WriteAt(b, 0)
	return
}

func (f *RingWriter) Offset() int {
	return f.offset
}
