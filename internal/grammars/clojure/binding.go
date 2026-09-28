package clojure

/*
#cgo CFLAGS: -I${SRCDIR}/grammar-src/src -I${SRCDIR}/src
#include <stdlib.h>
typedef struct TSLanguage TSLanguage;
const TSLanguage *tree_sitter_clojure(void);
*/
import "C"
import "unsafe"

func Language() unsafe.Pointer { return unsafe.Pointer(C.tree_sitter_clojure()) }
