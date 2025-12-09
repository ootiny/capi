package _rt_package_name_

import (
	"encoding/json"
	"strconv"
)

// ==========================================
// 1. 定义基础结构与通用行为
// ==========================================

// baseJsonField 包含状态位，利用 struct 内存对齐优化空间
// Present: JSON 中出现了该 Key
// Null:    JSON 中该 Key 的值为 null
type baseJsonField struct {
	Present bool
	Null    bool
}

func (b *baseJsonField) IsMissing() bool {
	return !b.Present
}

func (b *baseJsonField) IsNull() bool {
	return b.Present && b.Null
}

func (b *baseJsonField) HasValue() bool {
	return b.Present && !b.Null
}

// 快速判断 null 的 helper，避免字节转字符串分配内存
func isNull(data []byte) bool {
	return len(data) == 4 && data[0] == 'n' && data[1] == 'u' && data[2] == 'l' && data[3] == 'l'
}

// ==========================================
// 2. 自定义 Int (高效版)
// ==========================================

type JsonInt struct {
	baseJsonField
	Val int
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (i *JsonInt) UnmarshalJSON(data []byte) error {
	i.Present = true

	// 1. Check Null (无内存分配)
	if isNull(data) {
		i.Null = true
		i.Val = 0
		return nil
	}

	// 2. Parse Value (使用 strconv 避免 json 库的反射开销)
	// 注意：JSON spec 允许数字也是 float 格式 (如 1.0)，
	// 如果需要极度严谨兼容浮点数转 int，这里可以用 json.Unmarshal。
	// 但为了"高效率"，strconv.Atoi 是最快的。
	val, err := strconv.Atoi(string(data))
	if err != nil {
		// 回退逻辑：如果遇到 1.00 这种格式，strconv.Atoi 会失败
		// 此时回退到标准库解析，保证健壮性
		var v int
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		i.Val = v
	} else {
		i.Val = val
	}

	i.Null = false
	return nil
}

// MarshalJSON 支持序列化：Missing -> 0 (或被 omitempty 忽略), Null -> null, Value -> 123
func (i JsonInt) MarshalJSON() ([]byte, error) {
	if i.Null {
		return []byte("null"), nil
	}
	if !i.Present {
		// 如果作为结构体字段且没有 omitempty，这里返回零值
		return []byte("0"), nil
	}
	return []byte(strconv.Itoa(i.Val)), nil
}

// ==========================================
// 3. 自定义 Float64 (高效版)
// ==========================================

type JsonFloat64 struct {
	baseJsonField
	Val float64
}

func (f *JsonFloat64) UnmarshalJSON(data []byte) error {
	f.Present = true
	if isNull(data) {
		f.Null = true
		f.Val = 0
		return nil
	}

	// strconv.ParseFloat 比 json.Unmarshal 快
	val, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		return err
	}
	f.Val = val
	f.Null = false
	return nil
}

func (f JsonFloat64) MarshalJSON() ([]byte, error) {
	if f.Null {
		return []byte("null"), nil
	}
	if !f.Present {
		return []byte("0"), nil
	}
	// 'f' (-ddd.dddd), -1 (自动精度), 64 (float64)
	return []byte(strconv.FormatFloat(f.Val, 'f', -1, 64)), nil
}

// ==========================================
// 4. 自定义 Bool (高效版)
// ==========================================

type JsonBool struct {
	baseJsonField
	Val bool
}

func (b *JsonBool) UnmarshalJSON(data []byte) error {
	b.Present = true
	if isNull(data) {
		b.Null = true
		b.Val = false
		return nil
	}

	val, err := strconv.ParseBool(string(data))
	if err != nil {
		return err
	}
	b.Val = val
	b.Null = false
	return nil
}

func (b JsonBool) MarshalJSON() ([]byte, error) {
	if b.Null {
		return []byte("null"), nil
	}
	if !b.Present {
		return []byte("false"), nil
	}
	if b.Val {
		return []byte("true"), nil
	}
	return []byte("false"), nil
}

// ==========================================
// 5. 自定义 String
// ==========================================

type JsonString struct {
	baseJsonField
	Val string
}

func (s *JsonString) UnmarshalJSON(data []byte) error {
	s.Present = true
	if isNull(data) {
		s.Null = true
		s.Val = ""
		return nil
	}

	// 字符串必须使用 json.Unmarshal，因为它处理转义字符 (如 \u003c, \n)
	// 简单的 string(data) 是错误的，因为它包含引号
	// 虽然有 strconv.Unquote，但 json.Unmarshal 更符合 JSON 标准且对 string 开销可接受
	if err := json.Unmarshal(data, &s.Val); err != nil {
		return err
	}
	s.Null = false
	return nil
}

func (s JsonString) MarshalJSON() ([]byte, error) {
	if s.Null {
		return []byte("null"), nil
	}
	if !s.Present {
		return []byte(`""`), nil
	}
	return json.Marshal(s.Val)
}
