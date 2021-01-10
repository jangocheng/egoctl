package gen

import (
	"egoctl/logger"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"sort"
	"strings"
)

type astParser struct {
	objectM     map[string]*SpecType // parser struct文件
	modelArr    []SpecType           // 模型生成的描述文件
	readContent string               // 读取原文件数据
	userOption  UserOption
	tmplOption  TmplOption
}

func AstParserBuild(userOption UserOption, tmplOption TmplOption) *astParser {
	a := &astParser{
		userOption: userOption,
		tmplOption: tmplOption,
		objectM:    make(map[string]*SpecType),
	}
	err := a.initReadContent()
	if err != nil {
		logger.Log.Fatalf("egoctl parse struct error, err: %s", err)
		return nil
	}
	a.parserStruct()
	return a
}

func (a *astParser) initReadContent() error {
	readinfo, err := ioutil.ReadFile(a.userOption.ScaffoldDSLFile)
	if err != nil {
		return err
	}
	a.readContent = string(readinfo)
	return nil
}

func (a *astParser) parserStruct() error {
	fSet := token.NewFileSet()

	// strings.NewReader
	f, err := parser.ParseFile(fSet, "", strings.NewReader(a.readContent), parser.ParseComments)
	if err != nil {
		panic(err)
	}

	commentMap := ast.NewCommentMap(fSet, f, f.Comments)
	f.Comments = commentMap.Filter(f).Comments()

	scope := f.Scope
	if scope == nil {
		return errors.New("struct nil")
	}
	objects := scope.Objects
	structs := make([]*SpecType, 0)
	for structName, obj := range objects {
		st, err := a.parseObject(structName, obj)
		if err != nil {
			return err
		}
		structs = append(structs, st)
	}
	sort.Slice(structs, func(i, j int) bool {
		return structs[i].Name < structs[j].Name
	})

	resp := make([]SpecType, 0)
	for _, item := range structs {
		resp = append(resp, *item)
	}
	a.modelArr = resp
	return nil
}

func (t *astParser) GetRenderInfos(descriptor Descriptor) (output []RenderInfo) {
	output = make([]RenderInfo, 0)
	// model table name, model table schema
	for _, content := range t.modelArr {
		output = append(output, RenderInfo{
			Module:     descriptor.Module,
			ModelName:  content.Name,
			Content:    content.ToModelInfos(),
			Option:     t.userOption,
			Descriptor: descriptor,
			TmplPath:   t.tmplOption.RenderPath,
		})
	}
	return
}
