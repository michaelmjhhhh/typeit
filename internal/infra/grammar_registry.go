package infra

import (
	c "github.com/michaelmjhhhh/typeit/internal/grammars/c"
	clojure "github.com/michaelmjhhhh/typeit/internal/grammars/clojure"
	cpp "github.com/michaelmjhhhh/typeit/internal/grammars/cpp"
	csharp "github.com/michaelmjhhhh/typeit/internal/grammars/csharp"
	dart "github.com/michaelmjhhhh/typeit/internal/grammars/dart"
	elixir "github.com/michaelmjhhhh/typeit/internal/grammars/elixir"
	erlang "github.com/michaelmjhhhh/typeit/internal/grammars/erlang"
	golang "github.com/michaelmjhhhh/typeit/internal/grammars/go"
	haskell "github.com/michaelmjhhhh/typeit/internal/grammars/haskell"
	java "github.com/michaelmjhhhh/typeit/internal/grammars/java"
	javascript "github.com/michaelmjhhhh/typeit/internal/grammars/javascript"
	kotlin "github.com/michaelmjhhhh/typeit/internal/grammars/kotlin"
	php "github.com/michaelmjhhhh/typeit/internal/grammars/php"
	python "github.com/michaelmjhhhh/typeit/internal/grammars/python"
	ruby "github.com/michaelmjhhhh/typeit/internal/grammars/ruby"
	rust "github.com/michaelmjhhhh/typeit/internal/grammars/rust"
	scala "github.com/michaelmjhhhh/typeit/internal/grammars/scala"
	swift "github.com/michaelmjhhhh/typeit/internal/grammars/swift"
	typescript "github.com/michaelmjhhhh/typeit/internal/grammars/typescript"
	zig "github.com/michaelmjhhhh/typeit/internal/grammars/zig"
	"unsafe"
)

var grammars = map[string]func() unsafe.Pointer{
	"c":          c.Language,
	"clojure":    clojure.Language,
	"cpp":        cpp.Language,
	"csharp":     csharp.Language,
	"dart":       dart.Language,
	"elixir":     elixir.Language,
	"erlang":     erlang.Language,
	"go":         golang.Language,
	"haskell":    haskell.Language,
	"java":       java.Language,
	"javascript": javascript.Language,
	"kotlin":     kotlin.Language,
	"php":        php.Language,
	"python":     python.Language,
	"ruby":       ruby.Language,
	"rust":       rust.Language,
	"scala":      scala.Language,
	"swift":      swift.Language,
	"typescript": typescript.Language,
	"zig":        zig.Language,
}
