package haskell

/*
// The scanner's Array casts violate strict aliasing under GCC optimization.
// https://github.com/tree-sitter/tree-sitter-haskell/issues/144
#cgo CFLAGS: -I${SRCDIR}/src -fno-strict-aliasing
#include <stdlib.h>
typedef struct TSLanguage TSLanguage;
const TSLanguage *tree_sitter_haskell(void);
*/
import "C"
import "unsafe"

func Language() unsafe.Pointer { return unsafe.Pointer(C.tree_sitter_haskell()) }
