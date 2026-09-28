package javascript

/*
#cgo CFLAGS: -I${SRCDIR}/src -I${SRCDIR}/src
#include <stdlib.h>
typedef struct TSLanguage TSLanguage;
const TSLanguage *tree_sitter_javascript(void);
*/
import "C"
import "unsafe"

func Language() unsafe.Pointer { return unsafe.Pointer(C.tree_sitter_javascript()) }
