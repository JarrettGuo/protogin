package generator

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"strings"
)

//go:embed template.go.tpl
var tpl string

// method represents an RPC method
type method struct {
	Name     string // GetDemoName
	Num      int    // 一个 rpc 方法可以对应多个http请求
	Request  string // *Req
	Response string // *Resp

	// http rule
	Path         string
	PathParams   []string
	Method       string
	Body         string
	ResponseBody string
}

func (m *method) HandlerName() string {
	return fmt.Sprintf("%s_%d", m.Name, m.Num)
}

// HasPathParams 是否包含路由参数
func (m *method) HasPathParams() bool {
	return len(m.PathParams) > 0
}

// initPathParams 转换参数路由 {xx} --> :xx
func (m *method) initPathParams() {
	if m.Path == "" {
		return
	}

	paths := strings.Split(m.Path, "/")
	m.PathParams = []string{} // 重置参数列表

	for i, p := range paths {
		if p != "" {
			// 处理 {param} 格式
			if len(p) > 2 && p[0] == '{' && p[len(p)-1] == '}' {
				paramName := p[1 : len(p)-1]
				paths[i] = ":" + paramName
				m.PathParams = append(m.PathParams, paramName)
			} else if len(p) > 1 && p[0] == ':' {
				// 已经是 :param 格式
				m.PathParams = append(m.PathParams, p[1:])
			}
		}
	}

	m.Path = strings.Join(paths, "/")
}

// 在 template.go 中修改 service 结构体
type service struct {
	Name      string
	FullName  string
	Methods   []*method
	MethodSet map[string]*method

	// 添加包引用
	GinPkg  string // 用于模板中的 gin 包引用
	HTTPPkg string // 用于模板中的 http 包引用
}

func (s *service) execute() string {
	if s.MethodSet == nil {
		s.MethodSet = make(map[string]*method, len(s.Methods))

		for _, m := range s.Methods {
			s.MethodSet[m.Name] = m
		}
	}

	buf := new(bytes.Buffer)

	// 注册 ProtoFieldName 函数（使用 protobuf 的命名规则）
	tmpl, err := template.New("http").Funcs(template.FuncMap{
		"ProtoFieldName": ProtoFieldName,
	}).Parse(strings.TrimSpace(tpl))

	if err != nil {
		panic(err)
	}

	if err := tmpl.Execute(buf, s); err != nil {
		panic(err)
	}

	return buf.String()
}

func (s *service) ServiceName() string {
	return s.Name + "Server"
}

// ProtoFieldName 将下划线分隔的字符串转换为 protobuf 生成的字段名
// 例如: user_id -> UserId (不是 UserID)
func ProtoFieldName(str string) string {
	if str == "" {
		return ""
	}

	// 替换连字符为下划线
	str = strings.ReplaceAll(str, "-", "_")

	// 分割并转换
	parts := strings.Split(str, "_")
	for i, part := range parts {
		if len(part) > 0 {
			// protobuf 只将每个部分的首字母大写，不做特殊处理
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}

	return strings.Join(parts, "")
}

// GoCamelCase 保留原来的函数，以防其他地方需要
func GoCamelCase(str string) string {
	if str == "" {
		return ""
	}

	// 替换连字符为下划线
	str = strings.ReplaceAll(str, "-", "_")

	// 分割并转换
	parts := strings.Split(str, "_")
	for i, part := range parts {
		if len(part) > 0 {
			// 特殊处理常见缩写
			upperPart := strings.ToUpper(part)
			if upperPart == "ID" || upperPart == "URL" || upperPart == "API" || upperPart == "HTTP" || upperPart == "HTTPS" {
				parts[i] = upperPart
			} else {
				// 首字母大写，其余小写
				parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
			}
		}
	}

	return strings.Join(parts, "")
}

// 原来的辅助函数
func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

func isASCIIDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

// GoCamelCaseProto - protobuf 风格的驼峰命名转换（备用）
func GoCamelCaseProto(str string) string {
	// Invariant: if the next letter is lower case, it must be converted
	// to upper case.
	// That is, we process a word at a time, where words are marked by _ or
	// upper case letter. Digits are treated as words.
	var b []byte
	for i := 0; i < len(str); i++ {
		c := str[i]
		switch {
		case c == '.' && i+1 < len(str) && isASCIILower(str[i+1]):
			// Skip over '.' in ".{{lowercase}}".
		case c == '.':
			b = append(b, '_') // convert '.' to '_'
		case c == '_' && (i == 0 || str[i-1] == '.'):
			// Convert initial '_' to ensure we start with a capital letter.
			// Do the same for '_' after '.' to match historic behavior.
			b = append(b, 'X') // convert '_' to 'X'
		case c == '_' && i+1 < len(str) && isASCIILower(str[i+1]):
			// Skip over '_' in "_{{lowercase}}".
		case isASCIIDigit(c):
			b = append(b, c)
		default:
			// Assume we have a letter now - if not, it's a bogus identifier.
			// The next word is a sequence of characters that must start upper case.
			if isASCIILower(c) {
				c -= 'a' - 'A' // convert lowercase to uppercase
			}
			b = append(b, c)

			// Accept lower case sequence that follows.
			for ; i+1 < len(str) && isASCIILower(str[i+1]); i++ {
				b = append(b, str[i+1])
			}
		}
	}
	return string(b)
}
