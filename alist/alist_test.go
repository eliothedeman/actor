package alist

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMake(t *testing.T) {
	assert.Equal(t, Make[int](10).Len(), 10)
}

func TestInsertIndex(t *testing.T) {
	m := Make[int](400)
	m.Insert(30, 30)
	assert.Equal(t, 30, m.Index(30))
}

func TestAppendIndex(t *testing.T) {
	m := Make[int](0)
	assert.Panics(t, func() {
		m.Index(0)
	})
	m.Append(4)
	assert.Equal(t, m.Index(0), 4)
}

func TestInsertOutOfRange(t *testing.T) {
	m := Make[int](400)
	assert.PanicsWithError(t, ErrIndexOutOfRange.Error(), func() {
		m.Insert(500, 400)
	})
}

func TestPopFront(t *testing.T) {
	m := Make[string](0)
	m.Append("end")
	assert.Equal(t, "end", m.PopFront())
	assert.Panics(t, func() { m.PopFront() })
}
func TestAppendPopAppend(t *testing.T) {
	m := Make[string](0)
	m.Append("end")
	m.PopFront()
	m.Append("again")
	assert.Equal(t, "again", m.PopFront())
	assert.Panics(t, func() { m.PopFront() })
}

func BenchmarkInsert(b *testing.B) {
	for i := 4; i < 10; i++ {
		size := 1 << (i * 2)
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			b.Run("slice", func(b *testing.B) {
				m := make([]int, size)
				for i := 0; i < b.N; i++ {
					m[i%size] = i
				}
			})
			b.Run("alist", func(b *testing.B) {
				m := Make[int](size)
				for i := 0; i < b.N; i++ {
					m.Insert(i%size, i)
				}
			})
		})
	}
}

func BenchmarkAppend(b *testing.B) {
	for i := 4; i < 10; i++ {
		size := 1 << (i * 2)
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			b.Run("slice", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					m := make([]int, 0)
					for i := 0; i < size; i++ {
						m = append(m, i)
					}
					_ = m
				}
				b.ReportMetric(float64(size*b.N), "appends/op")
			})
			b.Run("alist", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					m := Make[int](0)
					for i := 0; i < size; i++ {
						m.Append(i)
					}
					_ = m
				}
				b.ReportMetric(float64(size*b.N), "appends/op")
			})
		})
	}
}
