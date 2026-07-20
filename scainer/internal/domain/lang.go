package domain

type Lang string

const (
	LangCPP        Lang = "cpp"
	LangPython     Lang = "python"
	LangJava       Lang = "java"
	LangGo         Lang = "go"
	LangJavaScript Lang = "javascript"
)

var FileExt = map[Lang]string{
	LangCPP:        ".cpp",
	LangPython:     ".py",
	LangJava:       ".java",
	LangGo:         ".go",
	LangJavaScript: ".js",
}
