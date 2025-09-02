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
	JSONPkg      string
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
	GinPkg    string // 用于模板中的 gin 包引用
	HTTPPkg   string // 用于模板中的 http 包引用
	CodesPkg  string // 用于模板中的 codes 包引用
	StatusPkg string // 用于模板中的 status 包引用
	JSONPkg   string // 用于模板中的 json 包引用

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

func ProtoFieldName(str string) string {
	if str == "" {
		return ""
	}

	str = strings.ReplaceAll(str, "-", "_")
	parts := strings.Split(str, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}

	return strings.Join(parts, "")
}
