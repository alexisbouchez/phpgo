package interpreter

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alexisbouchez/phpgo/ast"
	"github.com/alexisbouchez/phpgo/runtime"
)

func (i *Interpreter) registerBuiltins() {
	// Register built-in interfaces
	i.registerArrayAccessInterface()
	i.registerIteratorInterfaces()
}

func (i *Interpreter) registerArrayAccessInterface() {
	// ArrayAccess interface with its 4 methods
	arrayAccess := &runtime.Interface{
		Name: "ArrayAccess",
		Methods: map[string]*runtime.Method{
			"offsetExists": {
				Name:     "offsetExists",
				Params:   []string{"offset"},
				IsPublic: true,
			},
			"offsetGet": {
				Name:     "offsetGet",
				Params:   []string{"offset"},
				IsPublic: true,
			},
			"offsetSet": {
				Name:     "offsetSet",
				Params:   []string{"offset", "value"},
				IsPublic: true,
			},
			"offsetUnset": {
				Name:     "offsetUnset",
				Params:   []string{"offset"},
				IsPublic: true,
			},
		},
	}
	i.env.DefineInterface("ArrayAccess", arrayAccess)
}

func (i *Interpreter) registerIteratorInterfaces() {
	// Traversable is a marker interface (no methods)
	traversable := &runtime.Interface{
		Name:    "Traversable",
		Methods: map[string]*runtime.Method{},
	}
	i.env.DefineInterface("Traversable", traversable)

	// Iterator interface extends Traversable
	iterator := &runtime.Interface{
		Name: "Iterator",
		Methods: map[string]*runtime.Method{
			"current": {
				Name:     "current",
				Params:   []string{},
				IsPublic: true,
			},
			"key": {
				Name:     "key",
				Params:   []string{},
				IsPublic: true,
			},
			"next": {
				Name:     "next",
				Params:   []string{},
				IsPublic: true,
			},
			"rewind": {
				Name:     "rewind",
				Params:   []string{},
				IsPublic: true,
			},
			"valid": {
				Name:     "valid",
				Params:   []string{},
				IsPublic: true,
			},
		},
	}
	i.env.DefineInterface("Iterator", iterator)
}

func (i *Interpreter) getBuiltin(name string) runtime.BuiltinFunc {
	switch strings.ToLower(name) {
	// String functions
	case "strlen":
		return builtinStrlen
	case "substr":
		return builtinSubstr
	case "strpos":
		return builtinStrpos
	case "str_replace":
		return builtinStrReplace
	case "strtoupper":
		return builtinStrtoupper
	case "strtolower":
		return builtinStrtolower
	case "trim":
		return builtinTrim
	case "ltrim":
		return builtinLtrim
	case "rtrim":
		return builtinRtrim
	case "explode":
		return builtinExplode
	case "implode", "join":
		return builtinImplode
	case "sprintf":
		return builtinSprintf
	case "str_repeat":
		return builtinStrRepeat
	case "ucfirst":
		return builtinUcfirst
	case "lcfirst":
		return builtinLcfirst
	case "ucwords":
		return builtinUcwords
	case "str_pad":
		return builtinStrPad
	case "str_split":
		return builtinStrSplit
	case "chunk_split":
		return builtinChunkSplit
	case "wordwrap":
		return builtinWordwrap
	case "nl2br":
		return builtinNl2br
	case "ord":
		return builtinOrd
	case "chr":
		return builtinChr

	// Array functions
	case "count", "sizeof":
		return builtinCount
	case "array_push":
		return builtinArrayPush
	case "array_pop":
		return builtinArrayPop
	case "array_shift":
		return builtinArrayShift
	case "array_unshift":
		return builtinArrayUnshift
	case "array_merge":
		return builtinArrayMerge
	case "array_keys":
		return builtinArrayKeys
	case "array_values":
		return builtinArrayValues
	case "array_reverse":
		return builtinArrayReverse
	case "array_slice":
		return builtinArraySlice
	case "array_search":
		return builtinArraySearch
	case "in_array":
		return builtinInArray
	case "array_key_exists":
		return builtinArrayKeyExists
	case "array_map":
		return i.builtinArrayMap
	case "array_filter":
		return i.builtinArrayFilter
	case "array_reduce":
		return i.builtinArrayReduce
	case "array_unique":
		return builtinArrayUnique
	case "array_flip":
		return builtinArrayFlip
	case "array_sum":
		return builtinArraySum
	case "array_product":
		return builtinArrayProduct
	case "range":
		return builtinRange
	case "sort":
		return builtinSort
	case "rsort":
		return builtinRsort

	// Math functions
	case "abs":
		return builtinAbs
	case "ceil":
		return builtinCeil
	case "floor":
		return builtinFloor
	case "round":
		return builtinRound
	case "max":
		return builtinMax
	case "min":
		return builtinMin
	case "pow":
		return builtinPow
	case "sqrt":
		return builtinSqrt
	case "rand":
		return builtinRand
	case "mt_rand":
		return builtinMtRand

	// Type functions
	case "gettype":
		return builtinGettype
	case "is_null":
		return builtinIsNull
	case "is_bool":
		return builtinIsBool
	case "is_int", "is_integer", "is_long":
		return builtinIsInt
	case "is_float", "is_double", "is_real":
		return builtinIsFloat
	case "is_string":
		return builtinIsString
	case "is_array":
		return builtinIsArray
	case "is_object":
		return builtinIsObject
	case "is_numeric":
		return builtinIsNumeric
	case "intval":
		return builtinIntval
	case "floatval", "doubleval":
		return builtinFloatval
	case "strval":
		return builtinStrval
	case "boolval":
		return builtinBoolval

	// Output functions
	case "var_dump":
		return i.builtinVarDump
	case "print_r":
		return i.builtinPrintR

	// Output buffering functions
	case "ob_start":
		return i.builtinObStart
	case "ob_end_clean":
		return i.builtinObEndClean
	case "ob_end_flush":
		return i.builtinObEndFlush
	case "ob_get_contents":
		return i.builtinObGetContents
	case "ob_get_clean":
		return i.builtinObGetClean
	case "ob_get_flush":
		return i.builtinObGetFlush
	case "ob_get_level":
		return i.builtinObGetLevel
	case "ob_flush":
		return i.builtinObFlush
	case "ob_clean":
		return i.builtinObClean

	// Misc functions
	case "defined":
		return i.builtinDefined
	case "function_exists":
		return i.builtinFunctionExists
	case "class_exists":
		return i.builtinClassExists
	case "call_user_func":
		return i.builtinCallUserFunc
	case "call_user_func_array":
		return i.builtinCallUserFuncArray
	case "func_get_args":
		return i.builtinFuncGetArgs
	case "func_num_args":
		return i.builtinFuncNumArgs

	// Regex functions
	case "preg_match":
		return builtinPregMatch
	case "preg_match_all":
		return builtinPregMatchAll
	case "preg_replace":
		return builtinPregReplace
	case "preg_split":
		return builtinPregSplit

	// JSON functions
	case "json_encode":
		return builtinJsonEncode
	case "json_decode":
		return builtinJsonDecode
	case "serialize":
		return i.builtinSerialize
	case "unserialize":
		return i.builtinUnserialize

	// File functions
	case "file_get_contents":
		return builtinFileGetContents
	case "file_put_contents":
		return builtinFilePutContents
	case "file_exists":
		return builtinFileExists
	case "is_file":
		return builtinIsFile
	case "is_dir":
		return builtinIsDir
	case "is_readable":
		return builtinIsReadable
	case "is_writable", "is_writeable":
		return builtinIsWritable
	case "file":
		return builtinFile
	case "dirname":
		return builtinDirname
	case "basename":
		return builtinBasename
	case "pathinfo":
		return builtinPathinfo
	case "realpath":
		return builtinRealpath
	case "glob":
		return builtinGlob

	// Date/time functions
	case "time":
		return builtinTime
	case "date":
		return builtinDate
	case "strtotime":
		return builtinStrtotime
	case "mktime":
		return builtinMktime
	case "microtime":
		return builtinMicrotime

	// Hash functions
	case "md5":
		return builtinMd5
	case "sha1":
		return builtinSha1
	case "hash":
		return builtinHash
	case "base64_encode":
		return builtinBase64Encode
	case "base64_decode":
		return builtinBase64Decode

	// Additional string functions
	case "str_contains":
		return builtinStrContains
	case "str_starts_with":
		return builtinStrStartsWith
	case "str_ends_with":
		return builtinStrEndsWith
	case "number_format":
		return builtinNumberFormat
	case "money_format":
		return builtinNumberFormat
	case "htmlspecialchars":
		return builtinHtmlspecialchars
	case "htmlentities":
		return builtinHtmlentities
	case "strip_tags":
		return builtinStripTags
	case "addslashes":
		return builtinAddslashes
	case "stripslashes":
		return builtinStripslashes

	// Additional array functions
	case "array_combine":
		return builtinArrayCombine
	case "array_fill":
		return builtinArrayFill
	case "array_chunk":
		return builtinArrayChunk
	case "array_column":
		return builtinArrayColumn
	case "array_count_values":
		return builtinArrayCountValues
	case "array_diff":
		return builtinArrayDiff
	case "array_intersect":
		return builtinArrayIntersect
	case "usort":
		return i.builtinUsort
	case "uasort":
		return i.builtinUasort
	case "uksort":
		return i.builtinUksort
	case "array_walk":
		return i.builtinArrayWalk
	case "array_walk_recursive":
		return i.builtinArrayWalkRecursive
	case "array_rand":
		return builtinArrayRand
	case "shuffle":
		return builtinShuffle

	// Additional math functions
	case "sin":
		return builtinSin
	case "cos":
		return builtinCos
	case "tan":
		return builtinTan
	case "log":
		return builtinLog
	case "exp":
		return builtinExp

	// URL functions
	case "parse_url":
		return builtinParseUrl
	case "http_build_query":
		return builtinHttpBuildQuery
	case "urlencode":
		return builtinUrlencode
	case "urldecode":
		return builtinUrldecode
	case "rawurlencode":
		return builtinRawurlencode
	case "rawurldecode":
		return builtinRawurldecode
	case "parse_str":
		return i.builtinParseStr

	// Object/Class introspection
	case "get_class":
		return builtinGetClass
	case "get_parent_class":
		return builtinGetParentClass
	case "get_class_methods":
		return builtinGetClassMethods
	case "method_exists":
		return builtinMethodExists
	case "property_exists":
		return builtinPropertyExists
	case "is_subclass_of":
		return i.builtinIsSubclassOf
	case "is_a":
		return i.builtinIsA

	// Additional string functions
	case "strstr", "strchr":
		return builtinStrstr
	case "strrchr":
		return builtinStrrchr
	case "substr_count":
		return builtinSubstrCount
	case "substr_compare":
		return builtinSubstrCompare
	case "strtr":
		return builtinStrtr
	case "str_ireplace":
		return builtinStrIreplace

	// Additional array functions
	case "asort":
		return builtinAsort
	case "arsort":
		return builtinArsort
	case "ksort":
		return builtinKsort
	case "krsort":
		return builtinKrsort
	case "array_splice":
		return builtinArraySplice

	// File stream functions
	case "fopen":
		return i.builtinFopen
	case "fclose":
		return builtinFclose
	case "fread":
		return builtinFread
	case "fwrite", "fputs":
		return builtinFwrite
	case "fgets":
		return builtinFgets
	case "feof":
		return builtinFeof
	case "fseek":
		return builtinFseek
	case "ftell":
		return builtinFtell
	case "rewind":
		return builtinRewind
	case "readfile":
		return builtinReadfile
	case "fgetcsv":
		return builtinFgetcsv
	case "fputcsv":
		return builtinFputcsv
	case "unlink":
		return builtinUnlink
	case "copy":
		return builtinCopy
	case "rename":
		return builtinRename
	case "chmod":
		return builtinChmod
	case "touch":
		return builtinTouch

	// Directory functions
	case "mkdir":
		return builtinMkdir
	case "rmdir":
		return builtinRmdir
	case "scandir":
		return builtinScandir
	case "chdir":
		return i.builtinChdir
	case "getcwd":
		return i.builtinGetcwd

	// Variable handling
	case "compact":
		return i.builtinCompact
	case "extract":
		return i.builtinExtract
	case "array_pad":
		return builtinArrayPad

	default:
		return nil
	}
}

// ----------------------------------------------------------------------------
// String functions

func builtinStrlen(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewInt(0)
	}
	return runtime.NewInt(int64(len(args[0].ToString())))
}

func builtinSubstr(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	str := args[0].ToString()
	start := int(args[1].ToInt())
	length := len(str)

	if len(args) >= 3 {
		length = int(args[2].ToInt())
	}

	if start < 0 {
		start = len(str) + start
	}
	if start < 0 {
		start = 0
	}
	if start >= len(str) {
		return runtime.NewString("")
	}

	end := start + length
	if length < 0 {
		end = len(str) + length
	}
	if end > len(str) {
		end = len(str)
	}
	if end < start {
		return runtime.NewString("")
	}

	return runtime.NewString(str[start:end])
}

func builtinStrpos(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	haystack := args[0].ToString()
	needle := args[1].ToString()
	offset := 0
	if len(args) >= 3 {
		offset = int(args[2].ToInt())
	}

	pos := strings.Index(haystack[offset:], needle)
	if pos == -1 {
		return runtime.FALSE
	}
	return runtime.NewInt(int64(pos + offset))
}

func builtinStrReplace(args ...runtime.Value) runtime.Value {
	if len(args) < 3 {
		return runtime.NewString("")
	}
	search := args[0].ToString()
	replace := args[1].ToString()
	subject := args[2].ToString()
	return runtime.NewString(strings.ReplaceAll(subject, search, replace))
}

func builtinStrtoupper(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	return runtime.NewString(strings.ToUpper(args[0].ToString()))
}

func builtinStrtolower(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	return runtime.NewString(strings.ToLower(args[0].ToString()))
}

func builtinTrim(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	s := args[0].ToString()
	if len(args) >= 2 {
		return runtime.NewString(strings.Trim(s, args[1].ToString()))
	}
	return runtime.NewString(strings.TrimSpace(s))
}

func builtinLtrim(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	s := args[0].ToString()
	if len(args) >= 2 {
		return runtime.NewString(strings.TrimLeft(s, args[1].ToString()))
	}
	return runtime.NewString(strings.TrimLeft(s, " \t\n\r\x00\x0B"))
}

func builtinRtrim(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	s := args[0].ToString()
	if len(args) >= 2 {
		return runtime.NewString(strings.TrimRight(s, args[1].ToString()))
	}
	return runtime.NewString(strings.TrimRight(s, " \t\n\r\x00\x0B"))
}

func builtinExplode(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	delimiter := args[0].ToString()
	str := args[1].ToString()
	limit := -1
	if len(args) >= 3 {
		limit = int(args[2].ToInt())
	}

	var parts []string
	if limit > 0 {
		parts = strings.SplitN(str, delimiter, limit)
	} else {
		parts = strings.Split(str, delimiter)
	}

	arr := runtime.NewArray()
	for _, part := range parts {
		arr.Set(nil, runtime.NewString(part))
	}
	return arr
}

func builtinImplode(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}

	var glue string
	var arr *runtime.Array

	if len(args) == 1 {
		if a, ok := args[0].(*runtime.Array); ok {
			arr = a
			glue = ""
		} else {
			return runtime.NewString("")
		}
	} else {
		glue = args[0].ToString()
		if a, ok := args[1].(*runtime.Array); ok {
			arr = a
		} else {
			return runtime.NewString("")
		}
	}

	parts := make([]string, 0, len(arr.Keys))
	for _, key := range arr.Keys {
		parts = append(parts, arr.Elements[key].ToString())
	}
	return runtime.NewString(strings.Join(parts, glue))
}

func builtinSprintf(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	format := args[0].ToString()
	fmtArgs := make([]interface{}, len(args)-1)
	for i := 1; i < len(args); i++ {
		switch v := args[i].(type) {
		case *runtime.Int:
			fmtArgs[i-1] = v.Value
		case *runtime.Float:
			fmtArgs[i-1] = v.Value
		case *runtime.String:
			fmtArgs[i-1] = v.Value
		default:
			fmtArgs[i-1] = args[i].ToString()
		}
	}
	return runtime.NewString(fmt.Sprintf(format, fmtArgs...))
}

func builtinStrRepeat(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewString("")
	}
	return runtime.NewString(strings.Repeat(args[0].ToString(), int(args[1].ToInt())))
}

func builtinUcfirst(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	s := args[0].ToString()
	if len(s) == 0 {
		return runtime.NewString("")
	}
	return runtime.NewString(strings.ToUpper(s[:1]) + s[1:])
}

func builtinLcfirst(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	s := args[0].ToString()
	if len(s) == 0 {
		return runtime.NewString("")
	}
	return runtime.NewString(strings.ToLower(s[:1]) + s[1:])
}

func builtinUcwords(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	return runtime.NewString(strings.Title(args[0].ToString()))
}

func builtinStrPad(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewString("")
	}
	s := args[0].ToString()
	length := int(args[1].ToInt())
	padStr := " "
	padType := 1 // STR_PAD_RIGHT

	if len(args) >= 3 {
		padStr = args[2].ToString()
	}
	if len(args) >= 4 {
		padType = int(args[3].ToInt())
	}

	if len(s) >= length {
		return runtime.NewString(s)
	}

	padLen := length - len(s)
	switch padType {
	case 0: // STR_PAD_LEFT
		return runtime.NewString(strings.Repeat(padStr, padLen/len(padStr)+1)[:padLen] + s)
	case 2: // STR_PAD_BOTH
		left := padLen / 2
		right := padLen - left
		return runtime.NewString(strings.Repeat(padStr, left/len(padStr)+1)[:left] + s + strings.Repeat(padStr, right/len(padStr)+1)[:right])
	default: // STR_PAD_RIGHT
		return runtime.NewString(s + strings.Repeat(padStr, padLen/len(padStr)+1)[:padLen])
	}
}

func builtinStrSplit(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	s := args[0].ToString()
	length := 1
	if len(args) >= 2 {
		length = int(args[1].ToInt())
	}

	arr := runtime.NewArray()
	for i := 0; i < len(s); i += length {
		end := i + length
		if end > len(s) {
			end = len(s)
		}
		arr.Set(nil, runtime.NewString(s[i:end]))
	}
	return arr
}

func builtinChunkSplit(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	s := args[0].ToString()
	chunklen := 76
	end := "\r\n"
	if len(args) >= 2 {
		chunklen = int(args[1].ToInt())
	}
	if len(args) >= 3 {
		end = args[2].ToString()
	}

	var result strings.Builder
	for i := 0; i < len(s); i += chunklen {
		e := i + chunklen
		if e > len(s) {
			e = len(s)
		}
		result.WriteString(s[i:e])
		result.WriteString(end)
	}
	return runtime.NewString(result.String())
}

func builtinWordwrap(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	s := args[0].ToString()
	width := 75
	breakStr := "\n"
	cut := false

	if len(args) >= 2 {
		width = int(args[1].ToInt())
	}
	if len(args) >= 3 {
		breakStr = args[2].ToString()
	}
	if len(args) >= 4 {
		cut = args[3].ToBool()
	}

	if !cut {
		// Simple word wrap
		words := strings.Fields(s)
		var result strings.Builder
		lineLen := 0
		for i, word := range words {
			if i > 0 {
				if lineLen+1+len(word) > width {
					result.WriteString(breakStr)
					lineLen = 0
				} else {
					result.WriteString(" ")
					lineLen++
				}
			}
			result.WriteString(word)
			lineLen += len(word)
		}
		return runtime.NewString(result.String())
	}

	var result strings.Builder
	for i := 0; i < len(s); i += width {
		end := i + width
		if end > len(s) {
			end = len(s)
		}
		if i > 0 {
			result.WriteString(breakStr)
		}
		result.WriteString(s[i:end])
	}
	return runtime.NewString(result.String())
}

func builtinNl2br(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	s := args[0].ToString()
	s = strings.ReplaceAll(s, "\r\n", "<br />\r\n")
	s = strings.ReplaceAll(s, "\n", "<br />\n")
	s = strings.ReplaceAll(s, "\r", "<br />\r")
	return runtime.NewString(s)
}

func builtinOrd(args ...runtime.Value) runtime.Value {
	if len(args) < 1 || len(args[0].ToString()) == 0 {
		return runtime.NewInt(0)
	}
	return runtime.NewInt(int64(args[0].ToString()[0]))
}

func builtinChr(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	return runtime.NewString(string(rune(args[0].ToInt())))
}

// ----------------------------------------------------------------------------
// Array functions

func builtinCount(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewInt(0)
	}
	if arr, ok := args[0].(*runtime.Array); ok {
		return runtime.NewInt(int64(arr.Len()))
	}
	if _, ok := args[0].(*runtime.Null); ok {
		return runtime.NewInt(0)
	}
	return runtime.NewInt(1)
}

func builtinArrayPush(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NULL
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NULL
	}
	for i := 1; i < len(args); i++ {
		arr.Set(nil, args[i])
	}
	return runtime.NewInt(int64(arr.Len()))
}

func builtinArrayPop(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NULL
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok || arr.Len() == 0 {
		return runtime.NULL
	}
	lastKey := arr.Keys[len(arr.Keys)-1]
	lastVal := arr.Elements[lastKey]
	delete(arr.Elements, lastKey)
	arr.Keys = arr.Keys[:len(arr.Keys)-1]
	return lastVal
}

func builtinArrayShift(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NULL
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok || arr.Len() == 0 {
		return runtime.NULL
	}
	firstKey := arr.Keys[0]
	firstVal := arr.Elements[firstKey]
	delete(arr.Elements, firstKey)
	arr.Keys = arr.Keys[1:]
	// Re-index numeric keys
	newElements := make(map[runtime.Value]runtime.Value)
	newKeys := make([]runtime.Value, 0, len(arr.Keys))
	idx := int64(0)
	for _, key := range arr.Keys {
		if _, isInt := key.(*runtime.Int); isInt {
			newKey := runtime.NewInt(idx)
			newElements[newKey] = arr.Elements[key]
			newKeys = append(newKeys, newKey)
			idx++
		} else {
			newElements[key] = arr.Elements[key]
			newKeys = append(newKeys, key)
		}
	}
	arr.Elements = newElements
	arr.Keys = newKeys
	arr.NextIndex = idx
	return firstVal
}

func builtinArrayUnshift(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewInt(0)
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewInt(0)
	}

	// Create new array with prepended elements
	newElements := make(map[runtime.Value]runtime.Value)
	newKeys := make([]runtime.Value, 0, len(arr.Keys)+len(args)-1)

	idx := int64(0)
	for i := 1; i < len(args); i++ {
		key := runtime.NewInt(idx)
		newElements[key] = args[i]
		newKeys = append(newKeys, key)
		idx++
	}

	for _, key := range arr.Keys {
		if _, isInt := key.(*runtime.Int); isInt {
			newKey := runtime.NewInt(idx)
			newElements[newKey] = arr.Elements[key]
			newKeys = append(newKeys, newKey)
			idx++
		} else {
			newElements[key] = arr.Elements[key]
			newKeys = append(newKeys, key)
		}
	}

	arr.Elements = newElements
	arr.Keys = newKeys
	arr.NextIndex = idx
	return runtime.NewInt(int64(arr.Len()))
}

func builtinArrayMerge(args ...runtime.Value) runtime.Value {
	result := runtime.NewArray()
	for _, arg := range args {
		if arr, ok := arg.(*runtime.Array); ok {
			for _, key := range arr.Keys {
				if _, isInt := key.(*runtime.Int); isInt {
					result.Set(nil, arr.Elements[key])
				} else {
					result.Set(key, arr.Elements[key])
				}
			}
		}
	}
	return result
}

func builtinArrayKeys(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewArray()
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewArray()
	}

	result := runtime.NewArray()
	for _, key := range arr.Keys {
		result.Set(nil, key)
	}
	return result
}

func builtinArrayValues(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewArray()
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewArray()
	}

	result := runtime.NewArray()
	for _, key := range arr.Keys {
		result.Set(nil, arr.Elements[key])
	}
	return result
}

func builtinArrayReverse(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewArray()
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewArray()
	}

	result := runtime.NewArray()
	for i := len(arr.Keys) - 1; i >= 0; i-- {
		key := arr.Keys[i]
		result.Set(nil, arr.Elements[key])
	}
	return result
}

func builtinArraySlice(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewArray()
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewArray()
	}

	offset := int(args[1].ToInt())
	length := arr.Len()
	if len(args) >= 3 {
		length = int(args[2].ToInt())
	}

	if offset < 0 {
		offset = arr.Len() + offset
	}
	if offset < 0 {
		offset = 0
	}

	if length < 0 {
		length = arr.Len() + length - offset
	}

	result := runtime.NewArray()
	for i := offset; i < offset+length && i < len(arr.Keys); i++ {
		key := arr.Keys[i]
		result.Set(nil, arr.Elements[key])
	}
	return result
}

func builtinArraySearch(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	needle := args[0]
	arr, ok := args[1].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}

	strict := false
	if len(args) >= 3 {
		strict = args[2].ToBool()
	}

	for _, key := range arr.Keys {
		if strict {
			if runtime.IsIdentical(needle, arr.Elements[key]) {
				return key
			}
		} else {
			if runtime.IsEqual(needle, arr.Elements[key]) {
				return key
			}
		}
	}
	return runtime.FALSE
}

func builtinInArray(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	needle := args[0]
	arr, ok := args[1].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}

	strict := false
	if len(args) >= 3 {
		strict = args[2].ToBool()
	}

	for _, key := range arr.Keys {
		if strict {
			if runtime.IsIdentical(needle, arr.Elements[key]) {
				return runtime.TRUE
			}
		} else {
			if runtime.IsEqual(needle, arr.Elements[key]) {
				return runtime.TRUE
			}
		}
	}
	return runtime.FALSE
}

func builtinArrayKeyExists(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	key := args[0]
	arr, ok := args[1].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}

	if _, exists := arr.Elements[key]; exists {
		return runtime.TRUE
	}

	// Check with type coercion
	for k := range arr.Elements {
		if runtime.IsEqual(key, k) {
			return runtime.TRUE
		}
	}
	return runtime.FALSE
}

func (i *Interpreter) builtinArrayMap(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewArray()
	}

	callback, ok := args[0].(*runtime.Function)
	if !ok {
		return runtime.NewArray()
	}
	arr, ok := args[1].(*runtime.Array)
	if !ok {
		return runtime.NewArray()
	}

	result := runtime.NewArray()
	for _, key := range arr.Keys {
		val := arr.Elements[key]
		mapped := i.callFunctionWithArgs(callback, []runtime.Value{val})
		result.Set(nil, mapped)
	}
	return result
}

func (i *Interpreter) builtinArrayFilter(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewArray()
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewArray()
	}

	result := runtime.NewArray()

	if len(args) < 2 {
		// No callback - filter falsy values
		for _, key := range arr.Keys {
			val := arr.Elements[key]
			if val.ToBool() {
				result.Set(key, val)
			}
		}
	} else {
		callback, ok := args[1].(*runtime.Function)
		if !ok {
			return arr
		}
		for _, key := range arr.Keys {
			val := arr.Elements[key]
			keep := i.callFunctionWithArgs(callback, []runtime.Value{val})
			if keep.ToBool() {
				result.Set(key, val)
			}
		}
	}
	return result
}

func (i *Interpreter) builtinArrayReduce(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NULL
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NULL
	}
	callback, ok := args[1].(*runtime.Function)
	if !ok {
		return runtime.NULL
	}

	var carry runtime.Value = runtime.NULL
	if len(args) >= 3 {
		carry = args[2]
	}

	for _, key := range arr.Keys {
		val := arr.Elements[key]
		carry = i.callFunctionWithArgs(callback, []runtime.Value{carry, val})
	}
	return carry
}

func (i *Interpreter) callFunctionWithArgs(fn *runtime.Function, args []runtime.Value) runtime.Value {
	env := runtime.NewEnclosedEnvironment(fn.Env)
	oldEnv := i.env
	i.env = env

	// Save and set func args for func_get_args/func_num_args
	oldFuncArgs := i.currentFuncArgs
	i.currentFuncArgs = args

	for idx, param := range fn.Params {
		if idx < len(args) {
			env.Set(param, args[idx])
		}
	}

	var result runtime.Value = runtime.NULL
	if block, ok := fn.Body.(*ast.BlockStmt); ok {
		result = i.evalBlock(block)
	}

	i.env = oldEnv
	i.currentFuncArgs = oldFuncArgs

	if ret, ok := result.(*runtime.ReturnValue); ok {
		return ret.Value
	}
	return result
}

func builtinArrayUnique(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewArray()
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewArray()
	}

	seen := make(map[string]bool)
	result := runtime.NewArray()
	for _, key := range arr.Keys {
		val := arr.Elements[key]
		strVal := val.ToString()
		if !seen[strVal] {
			seen[strVal] = true
			result.Set(key, val)
		}
	}
	return result
}

func builtinArrayFlip(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewArray()
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewArray()
	}

	result := runtime.NewArray()
	for _, key := range arr.Keys {
		val := arr.Elements[key]
		result.Set(val, key)
	}
	return result
}

func builtinArraySum(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewInt(0)
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewInt(0)
	}

	var sum float64
	isFloat := false
	for _, key := range arr.Keys {
		val := arr.Elements[key]
		if _, ok := val.(*runtime.Float); ok {
			isFloat = true
		}
		sum += val.ToFloat()
	}
	if isFloat {
		return runtime.NewFloat(sum)
	}
	return runtime.NewInt(int64(sum))
}

func builtinArrayProduct(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewInt(0)
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewInt(0)
	}

	if arr.Len() == 0 {
		return runtime.NewInt(1)
	}

	product := 1.0
	isFloat := false
	for _, key := range arr.Keys {
		val := arr.Elements[key]
		if _, ok := val.(*runtime.Float); ok {
			isFloat = true
		}
		product *= val.ToFloat()
	}
	if isFloat {
		return runtime.NewFloat(product)
	}
	return runtime.NewInt(int64(product))
}

func builtinRange(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewArray()
	}

	start := args[0].ToInt()
	end := args[1].ToInt()
	step := int64(1)
	if len(args) >= 3 {
		step = args[2].ToInt()
		if step == 0 {
			step = 1
		}
	}

	result := runtime.NewArray()
	if start <= end {
		for i := start; i <= end; i += step {
			result.Set(nil, runtime.NewInt(i))
		}
	} else {
		if step > 0 {
			step = -step
		}
		for i := start; i >= end; i += step {
			result.Set(nil, runtime.NewInt(i))
		}
	}
	return result
}

func builtinSort(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}

	// Sort values and re-index
	vals := make([]runtime.Value, 0, len(arr.Keys))
	for _, key := range arr.Keys {
		vals = append(vals, arr.Elements[key])
	}

	sort.Slice(vals, func(i, j int) bool {
		return runtime.Compare(vals[i], vals[j]) < 0
	})

	arr.Elements = make(map[runtime.Value]runtime.Value)
	arr.Keys = make([]runtime.Value, len(vals))
	arr.NextIndex = 0
	for i, v := range vals {
		key := runtime.NewInt(int64(i))
		arr.Keys[i] = key
		arr.Elements[key] = v
		arr.NextIndex = int64(i + 1)
	}

	return runtime.TRUE
}

func builtinRsort(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}

	vals := make([]runtime.Value, 0, len(arr.Keys))
	for _, key := range arr.Keys {
		vals = append(vals, arr.Elements[key])
	}

	sort.Slice(vals, func(i, j int) bool {
		return runtime.Compare(vals[i], vals[j]) > 0
	})

	arr.Elements = make(map[runtime.Value]runtime.Value)
	arr.Keys = make([]runtime.Value, len(vals))
	arr.NextIndex = 0
	for i, v := range vals {
		key := runtime.NewInt(int64(i))
		arr.Keys[i] = key
		arr.Elements[key] = v
		arr.NextIndex = int64(i + 1)
	}

	return runtime.TRUE
}

// ----------------------------------------------------------------------------
// Math functions

func builtinAbs(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewInt(0)
	}
	if f, ok := args[0].(*runtime.Float); ok {
		return runtime.NewFloat(math.Abs(f.Value))
	}
	v := args[0].ToInt()
	if v < 0 {
		return runtime.NewInt(-v)
	}
	return runtime.NewInt(v)
}

func builtinCeil(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewFloat(0)
	}
	return runtime.NewFloat(math.Ceil(args[0].ToFloat()))
}

func builtinFloor(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewFloat(0)
	}
	return runtime.NewFloat(math.Floor(args[0].ToFloat()))
}

func builtinRound(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewFloat(0)
	}
	precision := 0
	if len(args) >= 2 {
		precision = int(args[1].ToInt())
	}
	multiplier := math.Pow(10, float64(precision))
	return runtime.NewFloat(math.Round(args[0].ToFloat()*multiplier) / multiplier)
}

func builtinMax(args ...runtime.Value) runtime.Value {
	if len(args) == 0 {
		return runtime.NULL
	}
	if len(args) == 1 {
		if arr, ok := args[0].(*runtime.Array); ok {
			if arr.Len() == 0 {
				return runtime.NULL
			}
			var max runtime.Value = nil
			for _, key := range arr.Keys {
				val := arr.Elements[key]
				if max == nil || runtime.Compare(val, max) > 0 {
					max = val
				}
			}
			return max
		}
		return args[0]
	}
	var max runtime.Value = args[0]
	for i := 1; i < len(args); i++ {
		if runtime.Compare(args[i], max) > 0 {
			max = args[i]
		}
	}
	return max
}

func builtinMin(args ...runtime.Value) runtime.Value {
	if len(args) == 0 {
		return runtime.NULL
	}
	if len(args) == 1 {
		if arr, ok := args[0].(*runtime.Array); ok {
			if arr.Len() == 0 {
				return runtime.NULL
			}
			var min runtime.Value = nil
			for _, key := range arr.Keys {
				val := arr.Elements[key]
				if min == nil || runtime.Compare(val, min) < 0 {
					min = val
				}
			}
			return min
		}
		return args[0]
	}
	var min runtime.Value = args[0]
	for i := 1; i < len(args); i++ {
		if runtime.Compare(args[i], min) < 0 {
			min = args[i]
		}
	}
	return min
}

func builtinPow(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewInt(0)
	}
	return runtime.NewFloat(math.Pow(args[0].ToFloat(), args[1].ToFloat()))
}

func builtinSqrt(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewFloat(0)
	}
	return runtime.NewFloat(math.Sqrt(args[0].ToFloat()))
}

func builtinRand(args ...runtime.Value) runtime.Value {
	min := int64(0)
	max := int64(32767)
	if len(args) >= 2 {
		min = args[0].ToInt()
		max = args[1].ToInt()
	}
	// Simple pseudo-random (not cryptographically secure)
	return runtime.NewInt(min + (max-min+1)/2)
}

func builtinMtRand(args ...runtime.Value) runtime.Value {
	return builtinRand(args...)
}

// ----------------------------------------------------------------------------
// Type functions

func builtinGettype(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("NULL")
	}
	return runtime.NewString(args[0].Type())
}

func builtinIsNull(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.TRUE
	}
	_, ok := args[0].(*runtime.Null)
	return runtime.NewBool(ok)
}

func builtinIsBool(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	_, ok := args[0].(*runtime.Bool)
	return runtime.NewBool(ok)
}

func builtinIsInt(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	_, ok := args[0].(*runtime.Int)
	return runtime.NewBool(ok)
}

func builtinIsFloat(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	_, ok := args[0].(*runtime.Float)
	return runtime.NewBool(ok)
}

func builtinIsString(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	_, ok := args[0].(*runtime.String)
	return runtime.NewBool(ok)
}

func builtinIsArray(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	_, ok := args[0].(*runtime.Array)
	return runtime.NewBool(ok)
}

func builtinIsObject(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	_, ok := args[0].(*runtime.Object)
	return runtime.NewBool(ok)
}

func builtinIsNumeric(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	switch args[0].(type) {
	case *runtime.Int, *runtime.Float:
		return runtime.TRUE
	case *runtime.String:
		s := args[0].ToString()
		_, err1 := fmt.Sscanf(s, "%d", new(int))
		_, err2 := fmt.Sscanf(s, "%f", new(float64))
		return runtime.NewBool(err1 == nil || err2 == nil)
	}
	return runtime.FALSE
}

func builtinIntval(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewInt(0)
	}
	return runtime.NewInt(args[0].ToInt())
}

func builtinFloatval(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewFloat(0)
	}
	return runtime.NewFloat(args[0].ToFloat())
}

func builtinStrval(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	return runtime.NewString(args[0].ToString())
}

func builtinBoolval(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	return runtime.NewBool(args[0].ToBool())
}

// ----------------------------------------------------------------------------
// Output functions

func (i *Interpreter) builtinVarDump(args ...runtime.Value) runtime.Value {
	for _, arg := range args {
		i.writeOutput(i.inspectValue(arg))
		i.writeOutput("\n")
	}
	return runtime.NULL
}

func (i *Interpreter) builtinPrintR(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.TRUE
	}
	returnOutput := false
	if len(args) >= 2 {
		returnOutput = args[1].ToBool()
	}

	output := i.inspectValue(args[0])
	if returnOutput {
		return runtime.NewString(output)
	}
	i.writeOutput(output)
	return runtime.TRUE
}

// inspectValue returns a string representation, using __debugInfo for objects if available
func (i *Interpreter) inspectValue(v runtime.Value) string {
	if obj, ok := v.(*runtime.Object); ok {
		// Check for __debugInfo magic method
		if debugInfoMethod, _ := i.findMethod(obj.Class, "__debugInfo"); debugInfoMethod != nil {
			result := i.callArrayAccessMethod(obj, "__debugInfo", []runtime.Value{})
			if arr, ok := result.(*runtime.Array); ok {
				return i.formatDebugInfo(obj.Class.Name, arr)
			}
		}
	}
	return v.Inspect()
}

func (i *Interpreter) formatDebugInfo(className string, arr *runtime.Array) string {
	var sb strings.Builder
	sb.WriteString("object(")
	sb.WriteString(className)
	sb.WriteString(")#1 (")
	sb.WriteString(strconv.Itoa(arr.Len()))
	sb.WriteString(") {\n")
	for _, key := range arr.Keys {
		sb.WriteString("  [\"")
		sb.WriteString(key.ToString())
		sb.WriteString("\"] => ")
		sb.WriteString(arr.Elements[key].Inspect())
		sb.WriteString("\n")
	}
	sb.WriteString("}")
	return sb.String()
}

// ----------------------------------------------------------------------------
// Output buffering functions

func (i *Interpreter) builtinObStart(args ...runtime.Value) runtime.Value {
	i.outputBuffers = append(i.outputBuffers, &strings.Builder{})
	return runtime.TRUE
}

func (i *Interpreter) builtinObEndClean(args ...runtime.Value) runtime.Value {
	if len(i.outputBuffers) == 0 {
		return runtime.FALSE
	}
	i.outputBuffers = i.outputBuffers[:len(i.outputBuffers)-1]
	return runtime.TRUE
}

func (i *Interpreter) builtinObEndFlush(args ...runtime.Value) runtime.Value {
	if len(i.outputBuffers) == 0 {
		return runtime.FALSE
	}
	content := i.outputBuffers[len(i.outputBuffers)-1].String()
	i.outputBuffers = i.outputBuffers[:len(i.outputBuffers)-1]
	i.flushToOutput(content)
	return runtime.TRUE
}

func (i *Interpreter) builtinObGetContents(args ...runtime.Value) runtime.Value {
	if len(i.outputBuffers) == 0 {
		return runtime.FALSE
	}
	return runtime.NewString(i.outputBuffers[len(i.outputBuffers)-1].String())
}

func (i *Interpreter) builtinObGetClean(args ...runtime.Value) runtime.Value {
	if len(i.outputBuffers) == 0 {
		return runtime.FALSE
	}
	content := i.outputBuffers[len(i.outputBuffers)-1].String()
	i.outputBuffers = i.outputBuffers[:len(i.outputBuffers)-1]
	return runtime.NewString(content)
}

func (i *Interpreter) builtinObGetFlush(args ...runtime.Value) runtime.Value {
	if len(i.outputBuffers) == 0 {
		return runtime.FALSE
	}
	content := i.outputBuffers[len(i.outputBuffers)-1].String()
	i.outputBuffers = i.outputBuffers[:len(i.outputBuffers)-1]
	i.flushToOutput(content)
	return runtime.NewString(content)
}

func (i *Interpreter) builtinObGetLevel(args ...runtime.Value) runtime.Value {
	return runtime.NewInt(int64(len(i.outputBuffers)))
}

func (i *Interpreter) builtinObFlush(args ...runtime.Value) runtime.Value {
	if len(i.outputBuffers) == 0 {
		return runtime.FALSE
	}
	content := i.outputBuffers[len(i.outputBuffers)-1].String()
	i.outputBuffers[len(i.outputBuffers)-1].Reset()
	i.flushToOutput(content)
	return runtime.TRUE
}

func (i *Interpreter) builtinObClean(args ...runtime.Value) runtime.Value {
	if len(i.outputBuffers) == 0 {
		return runtime.FALSE
	}
	i.outputBuffers[len(i.outputBuffers)-1].Reset()
	return runtime.TRUE
}

// ----------------------------------------------------------------------------
// Misc functions

func (i *Interpreter) builtinDefined(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	name := args[0].ToString()
	_, ok := i.env.GetConstant(name)
	return runtime.NewBool(ok)
}

func (i *Interpreter) builtinFunctionExists(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	name := args[0].ToString()
	if i.getBuiltin(name) != nil {
		return runtime.TRUE
	}
	_, ok := i.env.GetFunction(name)
	return runtime.NewBool(ok)
}

func (i *Interpreter) builtinClassExists(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	name := args[0].ToString()
	_, ok := i.env.GetClass(name)
	return runtime.NewBool(ok)
}

func (i *Interpreter) builtinCallUserFunc(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NULL
	}

	callback := args[0]
	callArgs := args[1:]

	return i.callCallback(callback, callArgs)
}

func (i *Interpreter) builtinCallUserFuncArray(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NULL
	}

	callback := args[0]
	argsArray, ok := args[1].(*runtime.Array)
	if !ok {
		return runtime.NULL
	}

	// Convert array to slice of values
	callArgs := make([]runtime.Value, 0, argsArray.Len())
	for _, key := range argsArray.Keys {
		callArgs = append(callArgs, argsArray.Elements[key])
	}

	return i.callCallback(callback, callArgs)
}

func (i *Interpreter) builtinFuncGetArgs(args ...runtime.Value) runtime.Value {
	result := runtime.NewArray()
	for idx, arg := range i.currentFuncArgs {
		result.Set(runtime.NewInt(int64(idx)), arg)
	}
	return result
}

func (i *Interpreter) builtinFuncNumArgs(args ...runtime.Value) runtime.Value {
	return runtime.NewInt(int64(len(i.currentFuncArgs)))
}

// callCallback handles calling various callback types
func (i *Interpreter) callCallback(callback runtime.Value, args []runtime.Value) runtime.Value {
	switch cb := callback.(type) {
	case *runtime.Function:
		// Closure or anonymous function
		return i.callFunctionWithArgs(cb, args)

	case *runtime.String:
		// Function name as string
		funcName := cb.Value

		// Check for builtin first
		if builtin := i.getBuiltin(funcName); builtin != nil {
			return builtin(args...)
		}

		// Check for user function
		resolvedName := i.resolveFunctionName(funcName)
		if fn, ok := i.env.GetFunction(resolvedName); ok {
			return i.callFunctionWithArgs(fn, args)
		}

		// Try original name
		if fn, ok := i.env.GetFunction(funcName); ok {
			return i.callFunctionWithArgs(fn, args)
		}

		return runtime.NULL

	case *runtime.Array:
		// Array callback: [$object, 'method'] or ['ClassName', 'method']
		if cb.Len() != 2 {
			return runtime.NULL
		}

		first := cb.Elements[cb.Keys[0]]
		second := cb.Elements[cb.Keys[1]]
		methodName := second.ToString()

		switch target := first.(type) {
		case *runtime.Object:
			// Instance method call
			if method, foundClass := i.findMethod(target.Class, methodName); method != nil {
				return i.invokeMethodWithArgs(target, method, foundClass, args)
			}

		case *runtime.String:
			// Static method call
			className := i.resolveClassName(target.Value)
			class, ok := i.env.GetClass(className)
			if !ok {
				class, ok = i.env.GetClass(target.Value)
			}
			if ok {
				if method, foundClass := i.findMethod(class, methodName); method != nil && method.IsStatic {
					return i.invokeStaticMethodWithArgs(class, method, foundClass, args)
				}
			}
		}
	}

	return runtime.NULL
}

// invokeMethodWithArgs calls an object method with given args
func (i *Interpreter) invokeMethodWithArgs(obj *runtime.Object, method *runtime.Method, foundClass *runtime.Class, args []runtime.Value) runtime.Value {
	env := runtime.NewEnclosedEnvironment(i.env)
	env.Set("this", obj)
	oldEnv := i.env
	oldClass := i.currentClass
	oldThis := i.currentThis
	oldFuncArgs := i.currentFuncArgs
	i.env = env
	i.currentClass = foundClass.Name
	i.currentThis = obj
	i.currentFuncArgs = args

	// Bind parameters
	for idx, param := range method.Params {
		if idx < len(args) {
			env.Set(param, args[idx])
		} else if idx < len(method.Defaults) && method.Defaults[idx] != nil {
			env.Set(param, method.Defaults[idx])
		}
	}

	// Handle variadic
	if method.Variadic && len(method.Params) > 0 {
		lastParam := method.Params[len(method.Params)-1]
		variadicArgs := runtime.NewArray()
		for idx := len(method.Params) - 1; idx < len(args); idx++ {
			variadicArgs.Set(nil, args[idx])
		}
		env.Set(lastParam, variadicArgs)
	}

	var result runtime.Value = runtime.NULL
	if block, ok := method.Body.(*ast.BlockStmt); ok {
		result = i.evalBlock(block)
	}

	i.env = oldEnv
	i.currentClass = oldClass
	i.currentThis = oldThis
	i.currentFuncArgs = oldFuncArgs

	if ret, ok := result.(*runtime.ReturnValue); ok {
		return ret.Value
	}
	return result
}

// invokeStaticMethodWithArgs calls a static method with given args
func (i *Interpreter) invokeStaticMethodWithArgs(class *runtime.Class, method *runtime.Method, foundClass *runtime.Class, args []runtime.Value) runtime.Value {
	env := runtime.NewEnclosedEnvironment(i.env)
	oldEnv := i.env
	oldClass := i.currentClass
	oldFuncArgs := i.currentFuncArgs
	i.env = env
	i.currentClass = foundClass.Name
	i.currentFuncArgs = args

	// Bind parameters
	for idx, param := range method.Params {
		if idx < len(args) {
			env.Set(param, args[idx])
		} else if idx < len(method.Defaults) && method.Defaults[idx] != nil {
			env.Set(param, method.Defaults[idx])
		}
	}

	var result runtime.Value = runtime.NULL
	if block, ok := method.Body.(*ast.BlockStmt); ok {
		result = i.evalBlock(block)
	}

	i.env = oldEnv
	i.currentClass = oldClass
	i.currentFuncArgs = oldFuncArgs

	if ret, ok := result.(*runtime.ReturnValue); ok {
		return ret.Value
	}
	return result
}

// ----------------------------------------------------------------------------
// Regex functions

func builtinPregMatch(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	pattern := args[0].ToString()
	subject := args[1].ToString()

	// Convert PHP regex delimiters to Go regex
	pattern = convertPHPRegex(pattern)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return runtime.FALSE
	}

	match := re.FindStringSubmatch(subject)
	if match == nil {
		return runtime.NewInt(0)
	}

	// If a third argument is provided, populate it with matches
	if len(args) >= 3 {
		if arr, ok := args[2].(*runtime.Array); ok {
			arr.Elements = make(map[runtime.Value]runtime.Value)
			arr.Keys = make([]runtime.Value, 0)
			arr.NextIndex = 0
			for _, m := range match {
				arr.Set(nil, runtime.NewString(m))
			}
		}
	}

	return runtime.NewInt(1)
}

func builtinPregMatchAll(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	pattern := args[0].ToString()
	subject := args[1].ToString()

	pattern = convertPHPRegex(pattern)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return runtime.FALSE
	}

	matches := re.FindAllStringSubmatch(subject, -1)
	if matches == nil {
		return runtime.NewInt(0)
	}

	// If a third argument is provided, populate it with matches
	if len(args) >= 3 {
		if arr, ok := args[2].(*runtime.Array); ok {
			arr.Elements = make(map[runtime.Value]runtime.Value)
			arr.Keys = make([]runtime.Value, 0)
			arr.NextIndex = 0

			// Group by capture index
			numGroups := len(matches[0])
			for g := 0; g < numGroups; g++ {
				groupArr := runtime.NewArray()
				for _, match := range matches {
					if g < len(match) {
						groupArr.Set(nil, runtime.NewString(match[g]))
					}
				}
				arr.Set(nil, groupArr)
			}
		}
	}

	return runtime.NewInt(int64(len(matches)))
}

func builtinPregReplace(args ...runtime.Value) runtime.Value {
	if len(args) < 3 {
		return runtime.NULL
	}
	pattern := args[0].ToString()
	replacement := args[1].ToString()
	subject := args[2].ToString()

	pattern = convertPHPRegex(pattern)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return runtime.NewString(subject)
	}

	result := re.ReplaceAllString(subject, replacement)
	return runtime.NewString(result)
}

func builtinPregSplit(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	pattern := args[0].ToString()
	subject := args[1].ToString()

	pattern = convertPHPRegex(pattern)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return runtime.FALSE
	}

	parts := re.Split(subject, -1)
	arr := runtime.NewArray()
	for _, part := range parts {
		arr.Set(nil, runtime.NewString(part))
	}
	return arr
}

func convertPHPRegex(pattern string) string {
	// Remove PHP regex delimiters (e.g., /pattern/flags)
	if len(pattern) >= 2 {
		delimiter := pattern[0]
		if delimiter == '/' || delimiter == '#' || delimiter == '~' {
			lastDelim := strings.LastIndexByte(pattern, delimiter)
			if lastDelim > 0 {
				// Extract pattern without delimiters and flags
				pattern = pattern[1:lastDelim]
			}
		}
	}
	return pattern
}

// ----------------------------------------------------------------------------
// JSON functions

func builtinJsonEncode(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	data := valueToInterface(args[0])
	result, err := json.Marshal(data)
	if err != nil {
		return runtime.FALSE
	}
	return runtime.NewString(string(result))
}

func builtinJsonDecode(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NULL
	}

	jsonStr := args[0].ToString()
	assoc := false
	if len(args) >= 2 {
		assoc = args[1].ToBool()
	}

	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return runtime.NULL
	}

	return interfaceToValue(data, assoc)
}

func valueToInterface(v runtime.Value) interface{} {
	switch val := v.(type) {
	case *runtime.Null:
		return nil
	case *runtime.Bool:
		return val.Value
	case *runtime.Int:
		return val.Value
	case *runtime.Float:
		return val.Value
	case *runtime.String:
		return val.Value
	case *runtime.Array:
		// Check if it's a sequential array or associative
		isSequential := true
		for i, key := range val.Keys {
			if intKey, ok := key.(*runtime.Int); !ok || intKey.Value != int64(i) {
				isSequential = false
				break
			}
		}

		if isSequential {
			result := make([]interface{}, len(val.Keys))
			for i, key := range val.Keys {
				result[i] = valueToInterface(val.Elements[key])
			}
			return result
		}

		result := make(map[string]interface{})
		for _, key := range val.Keys {
			result[key.ToString()] = valueToInterface(val.Elements[key])
		}
		return result
	default:
		return v.ToString()
	}
}

func interfaceToValue(data interface{}, assoc bool) runtime.Value {
	switch v := data.(type) {
	case nil:
		return runtime.NULL
	case bool:
		return runtime.NewBool(v)
	case float64:
		if v == float64(int64(v)) {
			return runtime.NewInt(int64(v))
		}
		return runtime.NewFloat(v)
	case string:
		return runtime.NewString(v)
	case []interface{}:
		arr := runtime.NewArray()
		for _, item := range v {
			arr.Set(nil, interfaceToValue(item, assoc))
		}
		return arr
	case map[string]interface{}:
		arr := runtime.NewArray()
		for key, val := range v {
			arr.Set(runtime.NewString(key), interfaceToValue(val, assoc))
		}
		return arr
	default:
		return runtime.NULL
	}
}

// ----------------------------------------------------------------------------
// Serialization functions

func (i *Interpreter) builtinSerialize(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	return runtime.NewString(i.serializeValue(args[0]))
}

func (i *Interpreter) serializeValue(v runtime.Value) string {
	switch val := v.(type) {
	case *runtime.Null:
		return "N;"
	case *runtime.Bool:
		if val.Value {
			return "b:1;"
		}
		return "b:0;"
	case *runtime.Int:
		return fmt.Sprintf("i:%d;", val.Value)
	case *runtime.Float:
		return fmt.Sprintf("d:%s;", strconv.FormatFloat(val.Value, 'G', -1, 64))
	case *runtime.String:
		return fmt.Sprintf("s:%d:\"%s\";", len(val.Value), val.Value)
	case *runtime.Array:
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("a:%d:{", val.Len()))
		for _, key := range val.Keys {
			sb.WriteString(i.serializeValue(key))
			sb.WriteString(i.serializeValue(val.Elements[key]))
		}
		sb.WriteString("}")
		return sb.String()
	case *runtime.Object:
		return i.serializeObject(val)
	default:
		return "N;"
	}
}

func (i *Interpreter) serializeObject(obj *runtime.Object) string {
	className := obj.Class.Name

	// Check for __sleep magic method
	var propsToSerialize []string
	if sleepMethod, _ := i.findMethod(obj.Class, "__sleep"); sleepMethod != nil {
		result := i.callArrayAccessMethod(obj, "__sleep", []runtime.Value{})
		if arr, ok := result.(*runtime.Array); ok {
			for _, key := range arr.Keys {
				propsToSerialize = append(propsToSerialize, arr.Elements[key].ToString())
			}
		}
	} else {
		// Serialize all properties
		for propName := range obj.Properties {
			propsToSerialize = append(propsToSerialize, propName)
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("O:%d:\"%s\":%d:{", len(className), className, len(propsToSerialize)))
	for _, propName := range propsToSerialize {
		sb.WriteString(i.serializeValue(runtime.NewString(propName)))
		if val, ok := obj.Properties[propName]; ok {
			sb.WriteString(i.serializeValue(val))
		} else {
			sb.WriteString("N;")
		}
	}
	sb.WriteString("}")
	return sb.String()
}

func (i *Interpreter) builtinUnserialize(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	data := args[0].ToString()
	result, _ := i.unserializeValue(data, 0)
	return result
}

func (i *Interpreter) unserializeValue(data string, pos int) (runtime.Value, int) {
	if pos >= len(data) {
		return runtime.FALSE, pos
	}

	switch data[pos] {
	case 'N':
		// N;
		return runtime.NULL, pos + 2
	case 'b':
		// b:0; or b:1;
		if pos+3 < len(data) {
			val := data[pos+2] == '1'
			return runtime.NewBool(val), pos + 4
		}
		return runtime.FALSE, pos
	case 'i':
		// i:123;
		pos += 2 // skip "i:"
		end := strings.Index(data[pos:], ";")
		if end == -1 {
			return runtime.FALSE, pos
		}
		num, _ := strconv.ParseInt(data[pos:pos+end], 10, 64)
		return runtime.NewInt(num), pos + end + 1
	case 'd':
		// d:1.5;
		pos += 2 // skip "d:"
		end := strings.Index(data[pos:], ";")
		if end == -1 {
			return runtime.FALSE, pos
		}
		num, _ := strconv.ParseFloat(data[pos:pos+end], 64)
		return runtime.NewFloat(num), pos + end + 1
	case 's':
		// s:5:"hello";
		pos += 2 // skip "s:"
		colonPos := strings.Index(data[pos:], ":")
		if colonPos == -1 {
			return runtime.FALSE, pos
		}
		length, _ := strconv.Atoi(data[pos : pos+colonPos])
		pos += colonPos + 2 // skip length, :, and opening "
		str := data[pos : pos+length]
		return runtime.NewString(str), pos + length + 2 // +2 for closing ";
	case 'a':
		// a:2:{...}
		return i.unserializeArray(data, pos)
	case 'O':
		// O:8:"ClassName":2:{...}
		return i.unserializeObject(data, pos)
	default:
		return runtime.FALSE, pos + 1
	}
}

func (i *Interpreter) unserializeArray(data string, pos int) (runtime.Value, int) {
	pos += 2 // skip "a:"
	colonPos := strings.Index(data[pos:], ":")
	if colonPos == -1 {
		return runtime.FALSE, pos
	}
	count, _ := strconv.Atoi(data[pos : pos+colonPos])
	pos += colonPos + 2 // skip count, :, and {

	arr := runtime.NewArray()
	for idx := 0; idx < count; idx++ {
		var key, val runtime.Value
		key, pos = i.unserializeValue(data, pos)
		val, pos = i.unserializeValue(data, pos)
		arr.Set(key, val)
	}
	return arr, pos + 1 // +1 for closing }
}

func (i *Interpreter) unserializeObject(data string, pos int) (runtime.Value, int) {
	pos += 2 // skip "O:"

	// Get class name length
	colonPos := strings.Index(data[pos:], ":")
	if colonPos == -1 {
		return runtime.FALSE, pos
	}
	nameLen, _ := strconv.Atoi(data[pos : pos+colonPos])
	pos += colonPos + 2 // skip length, :, and opening "

	className := data[pos : pos+nameLen]
	pos += nameLen + 2 // skip name and closing "

	// Get property count
	colonPos = strings.Index(data[pos:], ":")
	if colonPos == -1 {
		return runtime.FALSE, pos
	}
	propCount, _ := strconv.Atoi(data[pos : pos+colonPos])
	pos += colonPos + 2 // skip count, :, and {

	// Get the class
	class, ok := i.env.GetClass(className)
	if !ok {
		return runtime.FALSE, pos
	}

	// Create object
	obj := runtime.NewObject(class)

	// Initialize default properties
	for propName, propDef := range class.Properties {
		if propDef.Default != nil {
			obj.Properties[propName] = propDef.Default
		}
	}

	// Read serialized properties
	for idx := 0; idx < propCount; idx++ {
		var propName, propVal runtime.Value
		propName, pos = i.unserializeValue(data, pos)
		propVal, pos = i.unserializeValue(data, pos)
		obj.Properties[propName.ToString()] = propVal
	}
	pos++ // skip closing }

	// Call __wakeup if it exists
	if wakeupMethod, _ := i.findMethod(class, "__wakeup"); wakeupMethod != nil {
		i.callArrayAccessMethod(obj, "__wakeup", []runtime.Value{})
	}

	return obj, pos
}

// ----------------------------------------------------------------------------
// File functions

func builtinFileGetContents(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	filename := args[0].ToString()
	data, err := os.ReadFile(filename)
	if err != nil {
		return runtime.FALSE
	}
	return runtime.NewString(string(data))
}

func builtinFilePutContents(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	filename := args[0].ToString()
	content := args[1].ToString()

	flags := 0
	if len(args) >= 3 {
		flags = int(args[2].ToInt())
	}

	var mode int
	if flags&8 != 0 { // FILE_APPEND
		mode = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	} else {
		mode = os.O_TRUNC | os.O_CREATE | os.O_WRONLY
	}

	f, err := os.OpenFile(filename, mode, 0644)
	if err != nil {
		return runtime.FALSE
	}
	defer f.Close()

	n, err := f.WriteString(content)
	if err != nil {
		return runtime.FALSE
	}
	return runtime.NewInt(int64(n))
}

func builtinFileExists(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	filename := args[0].ToString()
	_, err := os.Stat(filename)
	return runtime.NewBool(err == nil)
}

func builtinIsFile(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	filename := args[0].ToString()
	info, err := os.Stat(filename)
	if err != nil {
		return runtime.FALSE
	}
	return runtime.NewBool(!info.IsDir())
}

func builtinIsDir(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	filename := args[0].ToString()
	info, err := os.Stat(filename)
	if err != nil {
		return runtime.FALSE
	}
	return runtime.NewBool(info.IsDir())
}

func builtinIsReadable(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	filename := args[0].ToString()
	f, err := os.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		return runtime.FALSE
	}
	f.Close()
	return runtime.TRUE
}

func builtinIsWritable(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	filename := args[0].ToString()
	info, err := os.Stat(filename)
	if err != nil {
		return runtime.FALSE
	}
	// Check if file is writable by checking permissions
	return runtime.NewBool(info.Mode().Perm()&0200 != 0)
}

func builtinFile(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	filename := args[0].ToString()
	data, err := os.ReadFile(filename)
	if err != nil {
		return runtime.FALSE
	}

	lines := strings.Split(string(data), "\n")
	arr := runtime.NewArray()
	for _, line := range lines {
		arr.Set(nil, runtime.NewString(line+"\n"))
	}
	return arr
}

func builtinDirname(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	path := args[0].ToString()
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash == -1 {
		lastSlash = strings.LastIndex(path, "\\")
	}
	if lastSlash == -1 {
		return runtime.NewString(".")
	}
	if lastSlash == 0 {
		return runtime.NewString("/")
	}
	return runtime.NewString(path[:lastSlash])
}

func builtinBasename(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	path := args[0].ToString()
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash == -1 {
		lastSlash = strings.LastIndex(path, "\\")
	}
	base := path
	if lastSlash != -1 {
		base = path[lastSlash+1:]
	}

	// Remove suffix if provided
	if len(args) >= 2 {
		suffix := args[1].ToString()
		if strings.HasSuffix(base, suffix) {
			base = base[:len(base)-len(suffix)]
		}
	}
	return runtime.NewString(base)
}

func builtinPathinfo(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewArray()
	}
	path := args[0].ToString()

	dirname := builtinDirname(runtime.NewString(path)).ToString()
	basename := builtinBasename(runtime.NewString(path)).ToString()

	extension := ""
	filename := basename
	if dotIdx := strings.LastIndex(basename, "."); dotIdx != -1 {
		extension = basename[dotIdx+1:]
		filename = basename[:dotIdx]
	}

	arr := runtime.NewArray()
	arr.Set(runtime.NewString("dirname"), runtime.NewString(dirname))
	arr.Set(runtime.NewString("basename"), runtime.NewString(basename))
	arr.Set(runtime.NewString("extension"), runtime.NewString(extension))
	arr.Set(runtime.NewString("filename"), runtime.NewString(filename))
	return arr
}

func builtinRealpath(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	path := args[0].ToString()
	absPath, err := os.Getwd()
	if err != nil {
		return runtime.FALSE
	}
	if strings.HasPrefix(path, "/") {
		absPath = path
	} else {
		absPath = absPath + "/" + path
	}
	// Verify file exists
	if _, err := os.Stat(absPath); err != nil {
		return runtime.FALSE
	}
	return runtime.NewString(absPath)
}

func builtinGlob(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	pattern := args[0].ToString()
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return runtime.FALSE
	}

	arr := runtime.NewArray()
	for _, match := range matches {
		arr.Set(nil, runtime.NewString(match))
	}
	return arr
}

// ----------------------------------------------------------------------------
// Date/time functions

func builtinTime(args ...runtime.Value) runtime.Value {
	return runtime.NewInt(time.Now().Unix())
}

func builtinDate(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	format := args[0].ToString()
	timestamp := time.Now()
	if len(args) >= 2 {
		timestamp = time.Unix(args[1].ToInt(), 0)
	}

	// Convert PHP date format to Go format
	result := convertPHPDateFormat(format, timestamp)
	return runtime.NewString(result)
}

func convertPHPDateFormat(format string, t time.Time) string {
	var result strings.Builder
	for i := 0; i < len(format); i++ {
		c := format[i]
		switch c {
		case 'Y':
			result.WriteString(fmt.Sprintf("%04d", t.Year()))
		case 'y':
			result.WriteString(fmt.Sprintf("%02d", t.Year()%100))
		case 'm':
			result.WriteString(fmt.Sprintf("%02d", t.Month()))
		case 'n':
			result.WriteString(fmt.Sprintf("%d", t.Month()))
		case 'd':
			result.WriteString(fmt.Sprintf("%02d", t.Day()))
		case 'j':
			result.WriteString(fmt.Sprintf("%d", t.Day()))
		case 'H':
			result.WriteString(fmt.Sprintf("%02d", t.Hour()))
		case 'G':
			result.WriteString(fmt.Sprintf("%d", t.Hour()))
		case 'h':
			h := t.Hour() % 12
			if h == 0 {
				h = 12
			}
			result.WriteString(fmt.Sprintf("%02d", h))
		case 'g':
			h := t.Hour() % 12
			if h == 0 {
				h = 12
			}
			result.WriteString(fmt.Sprintf("%d", h))
		case 'i':
			result.WriteString(fmt.Sprintf("%02d", t.Minute()))
		case 's':
			result.WriteString(fmt.Sprintf("%02d", t.Second()))
		case 'a':
			if t.Hour() < 12 {
				result.WriteString("am")
			} else {
				result.WriteString("pm")
			}
		case 'A':
			if t.Hour() < 12 {
				result.WriteString("AM")
			} else {
				result.WriteString("PM")
			}
		case 'w':
			result.WriteString(fmt.Sprintf("%d", t.Weekday()))
		case 'N':
			wd := int(t.Weekday())
			if wd == 0 {
				wd = 7
			}
			result.WriteString(fmt.Sprintf("%d", wd))
		case 'D':
			result.WriteString(t.Weekday().String()[:3])
		case 'l':
			result.WriteString(t.Weekday().String())
		case 'M':
			result.WriteString(t.Month().String()[:3])
		case 'F':
			result.WriteString(t.Month().String())
		case 'U':
			result.WriteString(fmt.Sprintf("%d", t.Unix()))
		case 'c':
			result.WriteString(t.Format(time.RFC3339))
		case 'r':
			result.WriteString(t.Format(time.RFC1123Z))
		default:
			result.WriteByte(c)
		}
	}
	return result.String()
}

func builtinStrtotime(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	dateStr := args[0].ToString()
	baseTime := time.Now()
	if len(args) >= 2 {
		baseTime = time.Unix(args[1].ToInt(), 0)
	}

	// Try common formats
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006/01/02",
		"01/02/2006",
		"02-01-2006",
		"January 2, 2006",
		"Jan 2, 2006",
		time.RFC3339,
		time.RFC1123,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return runtime.NewInt(t.Unix())
		}
	}

	// Handle relative formats
	dateStr = strings.ToLower(strings.TrimSpace(dateStr))

	if dateStr == "now" {
		return runtime.NewInt(baseTime.Unix())
	}
	if dateStr == "today" {
		y, m, d := baseTime.Date()
		return runtime.NewInt(time.Date(y, m, d, 0, 0, 0, 0, baseTime.Location()).Unix())
	}
	if dateStr == "tomorrow" {
		t := baseTime.AddDate(0, 0, 1)
		y, m, d := t.Date()
		return runtime.NewInt(time.Date(y, m, d, 0, 0, 0, 0, t.Location()).Unix())
	}
	if dateStr == "yesterday" {
		t := baseTime.AddDate(0, 0, -1)
		y, m, d := t.Date()
		return runtime.NewInt(time.Date(y, m, d, 0, 0, 0, 0, t.Location()).Unix())
	}

	// Simple relative time parsing
	if strings.HasPrefix(dateStr, "+") || strings.HasPrefix(dateStr, "-") {
		var num int
		var unit string
		fmt.Sscanf(dateStr, "%d %s", &num, &unit)
		unit = strings.TrimSuffix(unit, "s")

		switch unit {
		case "second":
			return runtime.NewInt(baseTime.Add(time.Duration(num) * time.Second).Unix())
		case "minute":
			return runtime.NewInt(baseTime.Add(time.Duration(num) * time.Minute).Unix())
		case "hour":
			return runtime.NewInt(baseTime.Add(time.Duration(num) * time.Hour).Unix())
		case "day":
			return runtime.NewInt(baseTime.AddDate(0, 0, num).Unix())
		case "week":
			return runtime.NewInt(baseTime.AddDate(0, 0, num*7).Unix())
		case "month":
			return runtime.NewInt(baseTime.AddDate(0, num, 0).Unix())
		case "year":
			return runtime.NewInt(baseTime.AddDate(num, 0, 0).Unix())
		}
	}

	return runtime.FALSE
}

func builtinMktime(args ...runtime.Value) runtime.Value {
	now := time.Now()
	hour := now.Hour()
	minute := now.Minute()
	second := now.Second()
	month := int(now.Month())
	day := now.Day()
	year := now.Year()

	if len(args) >= 1 {
		hour = int(args[0].ToInt())
	}
	if len(args) >= 2 {
		minute = int(args[1].ToInt())
	}
	if len(args) >= 3 {
		second = int(args[2].ToInt())
	}
	if len(args) >= 4 {
		month = int(args[3].ToInt())
	}
	if len(args) >= 5 {
		day = int(args[4].ToInt())
	}
	if len(args) >= 6 {
		year = int(args[5].ToInt())
	}

	t := time.Date(year, time.Month(month), day, hour, minute, second, 0, now.Location())
	return runtime.NewInt(t.Unix())
}

func builtinMicrotime(args ...runtime.Value) runtime.Value {
	asFloat := false
	if len(args) >= 1 {
		asFloat = args[0].ToBool()
	}

	now := time.Now()
	if asFloat {
		return runtime.NewFloat(float64(now.UnixNano()) / 1e9)
	}

	usec := now.Nanosecond() / 1000
	sec := now.Unix()
	return runtime.NewString(fmt.Sprintf("0.%06d %d", usec, sec))
}

// ----------------------------------------------------------------------------
// Hash functions

func builtinMd5(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	hash := md5.Sum([]byte(args[0].ToString()))
	return runtime.NewString(hex.EncodeToString(hash[:]))
}

func builtinSha1(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	hash := sha1.Sum([]byte(args[0].ToString()))
	return runtime.NewString(hex.EncodeToString(hash[:]))
}

func builtinHash(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	algo := strings.ToLower(args[0].ToString())
	data := args[1].ToString()

	switch algo {
	case "md5":
		hash := md5.Sum([]byte(data))
		return runtime.NewString(hex.EncodeToString(hash[:]))
	case "sha1":
		hash := sha1.Sum([]byte(data))
		return runtime.NewString(hex.EncodeToString(hash[:]))
	default:
		return runtime.FALSE
	}
}

func builtinBase64Encode(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	return runtime.NewString(base64.StdEncoding.EncodeToString([]byte(args[0].ToString())))
}

func builtinBase64Decode(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	decoded, err := base64.StdEncoding.DecodeString(args[0].ToString())
	if err != nil {
		return runtime.FALSE
	}
	return runtime.NewString(string(decoded))
}

// ----------------------------------------------------------------------------
// Additional string functions

func builtinStrContains(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	haystack := args[0].ToString()
	needle := args[1].ToString()
	return runtime.NewBool(strings.Contains(haystack, needle))
}

func builtinStrStartsWith(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	haystack := args[0].ToString()
	needle := args[1].ToString()
	return runtime.NewBool(strings.HasPrefix(haystack, needle))
}

func builtinStrEndsWith(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	haystack := args[0].ToString()
	needle := args[1].ToString()
	return runtime.NewBool(strings.HasSuffix(haystack, needle))
}

func builtinNumberFormat(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("0")
	}
	num := args[0].ToFloat()
	decimals := 0
	decPoint := "."
	thousandsSep := ","

	if len(args) >= 2 {
		decimals = int(args[1].ToInt())
	}
	if len(args) >= 3 {
		decPoint = args[2].ToString()
	}
	if len(args) >= 4 {
		thousandsSep = args[3].ToString()
	}

	// Format the number
	format := fmt.Sprintf("%%.%df", decimals)
	str := fmt.Sprintf(format, num)

	// Split into integer and decimal parts
	parts := strings.Split(str, ".")
	intPart := parts[0]

	// Add thousands separator
	var result strings.Builder
	isNegative := strings.HasPrefix(intPart, "-")
	if isNegative {
		intPart = intPart[1:]
	}

	for i, c := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result.WriteString(thousandsSep)
		}
		result.WriteRune(c)
	}

	finalStr := result.String()
	if isNegative {
		finalStr = "-" + finalStr
	}

	if decimals > 0 && len(parts) > 1 {
		finalStr += decPoint + parts[1]
	}

	return runtime.NewString(finalStr)
}

func builtinHtmlspecialchars(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	s := args[0].ToString()
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#039;")
	return runtime.NewString(s)
}

func builtinHtmlentities(args ...runtime.Value) runtime.Value {
	return builtinHtmlspecialchars(args...)
}

func builtinStripTags(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	s := args[0].ToString()
	// Simple regex to remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	return runtime.NewString(re.ReplaceAllString(s, ""))
}

func builtinAddslashes(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	s := args[0].ToString()
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\x00", "\\0")
	return runtime.NewString(s)
}

func builtinStripslashes(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	s := args[0].ToString()
	s = strings.ReplaceAll(s, "\\\\", "\\")
	s = strings.ReplaceAll(s, "\\'", "'")
	s = strings.ReplaceAll(s, "\\\"", "\"")
	s = strings.ReplaceAll(s, "\\0", "\x00")
	return runtime.NewString(s)
}

// ----------------------------------------------------------------------------
// Additional array functions

func builtinArrayCombine(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	keys, ok1 := args[0].(*runtime.Array)
	values, ok2 := args[1].(*runtime.Array)
	if !ok1 || !ok2 || keys.Len() != values.Len() {
		return runtime.FALSE
	}

	result := runtime.NewArray()
	for i := range keys.Keys {
		keyVal := keys.Elements[keys.Keys[i]]
		valVal := values.Elements[values.Keys[i]]
		result.Set(keyVal, valVal)
	}
	return result
}

func builtinArrayFill(args ...runtime.Value) runtime.Value {
	if len(args) < 3 {
		return runtime.FALSE
	}
	startIndex := args[0].ToInt()
	num := args[1].ToInt()
	value := args[2]

	result := runtime.NewArray()
	for i := int64(0); i < num; i++ {
		result.Set(runtime.NewInt(startIndex+i), value)
	}
	return result
}

func builtinArrayChunk(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}
	size := int(args[1].ToInt())
	if size < 1 {
		return runtime.FALSE
	}

	preserveKeys := false
	if len(args) >= 3 {
		preserveKeys = args[2].ToBool()
	}

	result := runtime.NewArray()
	chunk := runtime.NewArray()
	count := 0

	for _, key := range arr.Keys {
		if preserveKeys {
			chunk.Set(key, arr.Elements[key])
		} else {
			chunk.Set(nil, arr.Elements[key])
		}
		count++

		if count >= size {
			result.Set(nil, chunk)
			chunk = runtime.NewArray()
			count = 0
		}
	}

	if chunk.Len() > 0 {
		result.Set(nil, chunk)
	}

	return result
}

func builtinArrayColumn(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}

	columnKey := args[1]
	var indexKey runtime.Value
	if len(args) >= 3 && args[2] != runtime.NULL {
		indexKey = args[2]
	}

	result := runtime.NewArray()
	for _, key := range arr.Keys {
		row, ok := arr.Elements[key].(*runtime.Array)
		if !ok {
			continue
		}

		colVal := row.Get(columnKey)
		if colVal == runtime.NULL {
			continue
		}

		if indexKey != nil {
			idx := row.Get(indexKey)
			result.Set(idx, colVal)
		} else {
			result.Set(nil, colVal)
		}
	}
	return result
}

func builtinArrayCountValues(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewArray()
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewArray()
	}

	counts := make(map[string]int64)
	for _, key := range arr.Keys {
		val := arr.Elements[key].ToString()
		counts[val]++
	}

	result := runtime.NewArray()
	for val, count := range counts {
		result.Set(runtime.NewString(val), runtime.NewInt(count))
	}
	return result
}

func builtinArrayDiff(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewArray()
	}
	arr1, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewArray()
	}

	// Collect values from all other arrays
	exclude := make(map[string]bool)
	for i := 1; i < len(args); i++ {
		if arr, ok := args[i].(*runtime.Array); ok {
			for _, key := range arr.Keys {
				exclude[arr.Elements[key].ToString()] = true
			}
		}
	}

	result := runtime.NewArray()
	for _, key := range arr1.Keys {
		val := arr1.Elements[key]
		if !exclude[val.ToString()] {
			result.Set(key, val)
		}
	}
	return result
}

func builtinArrayIntersect(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewArray()
	}
	arr1, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewArray()
	}

	// Collect values that exist in ALL arrays
	counts := make(map[string]int)
	numArrays := len(args)

	for i := 0; i < numArrays; i++ {
		if arr, ok := args[i].(*runtime.Array); ok {
			seen := make(map[string]bool)
			for _, key := range arr.Keys {
				valStr := arr.Elements[key].ToString()
				if !seen[valStr] {
					seen[valStr] = true
					counts[valStr]++
				}
			}
		}
	}

	result := runtime.NewArray()
	for _, key := range arr1.Keys {
		val := arr1.Elements[key]
		if counts[val.ToString()] == numArrays {
			result.Set(key, val)
		}
	}
	return result
}

func (i *Interpreter) builtinUsort(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}
	callback, ok := args[1].(*runtime.Function)
	if !ok {
		return runtime.FALSE
	}

	vals := make([]runtime.Value, 0, len(arr.Keys))
	for _, key := range arr.Keys {
		vals = append(vals, arr.Elements[key])
	}

	sort.Slice(vals, func(x, y int) bool {
		result := i.callFunctionWithArgs(callback, []runtime.Value{vals[x], vals[y]})
		return result.ToInt() < 0
	})

	arr.Elements = make(map[runtime.Value]runtime.Value)
	arr.Keys = make([]runtime.Value, len(vals))
	arr.NextIndex = 0
	for idx, v := range vals {
		key := runtime.NewInt(int64(idx))
		arr.Keys[idx] = key
		arr.Elements[key] = v
		arr.NextIndex = int64(idx + 1)
	}

	return runtime.TRUE
}

func (i *Interpreter) builtinUasort(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}
	callback, ok := args[1].(*runtime.Function)
	if !ok {
		return runtime.FALSE
	}

	// Create slice of key-value pairs
	type kvPair struct {
		key runtime.Value
		val runtime.Value
	}
	pairs := make([]kvPair, 0, len(arr.Keys))
	for _, key := range arr.Keys {
		pairs = append(pairs, kvPair{key, arr.Elements[key]})
	}

	// Sort by value using callback
	sort.Slice(pairs, func(x, y int) bool {
		result := i.callFunctionWithArgs(callback, []runtime.Value{pairs[x].val, pairs[y].val})
		return result.ToInt() < 0
	})

	// Rebuild array with original keys
	arr.Keys = make([]runtime.Value, len(pairs))
	for idx, p := range pairs {
		arr.Keys[idx] = p.key
	}

	return runtime.TRUE
}

func (i *Interpreter) builtinUksort(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}
	callback, ok := args[1].(*runtime.Function)
	if !ok {
		return runtime.FALSE
	}

	// Sort keys using callback
	sort.Slice(arr.Keys, func(x, y int) bool {
		result := i.callFunctionWithArgs(callback, []runtime.Value{arr.Keys[x], arr.Keys[y]})
		return result.ToInt() < 0
	})

	return runtime.TRUE
}

func (i *Interpreter) builtinArrayWalk(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}
	callback, ok := args[1].(*runtime.Function)
	if !ok {
		return runtime.FALSE
	}

	for _, key := range arr.Keys {
		val := arr.Elements[key]
		i.callFunctionWithArgs(callback, []runtime.Value{val, key})
	}
	return runtime.TRUE
}

func (i *Interpreter) builtinArrayWalkRecursive(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}
	callback, ok := args[1].(*runtime.Function)
	if !ok {
		return runtime.FALSE
	}

	var walk func(*runtime.Array, runtime.Value)
	walk = func(a *runtime.Array, parentKey runtime.Value) {
		for _, key := range a.Keys {
			val := a.Elements[key]
			if childArr, ok := val.(*runtime.Array); ok {
				walk(childArr, key)
			} else {
				i.callFunctionWithArgs(callback, []runtime.Value{val, key})
			}
		}
	}

	walk(arr, runtime.NULL)
	return runtime.TRUE
}

func builtinArrayRand(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NULL
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok || arr.Len() == 0 {
		return runtime.NULL
	}

	num := 1
	if len(args) >= 2 {
		num = int(args[1].ToInt())
	}

	if num == 1 {
		// Return single random key
		idx := int(time.Now().UnixNano() % int64(len(arr.Keys)))
		return arr.Keys[idx]
	}

	// Return array of random keys
	result := runtime.NewArray()
	indices := make([]int, len(arr.Keys))
	for i := range indices {
		indices[i] = i
	}
	// Simple shuffle
	seed := time.Now().UnixNano()
	for i := len(indices) - 1; i > 0; i-- {
		j := int(seed % int64(i+1))
		seed = seed * 1103515245 + 12345
		indices[i], indices[j] = indices[j], indices[i]
	}

	for i := 0; i < num && i < len(indices); i++ {
		result.Set(nil, arr.Keys[indices[i]])
	}
	return result
}

func builtinShuffle(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}

	vals := make([]runtime.Value, 0, len(arr.Keys))
	for _, key := range arr.Keys {
		vals = append(vals, arr.Elements[key])
	}

	// Fisher-Yates shuffle
	seed := time.Now().UnixNano()
	for i := len(vals) - 1; i > 0; i-- {
		j := int(seed % int64(i+1))
		seed = seed * 1103515245 + 12345
		vals[i], vals[j] = vals[j], vals[i]
	}

	arr.Elements = make(map[runtime.Value]runtime.Value)
	arr.Keys = make([]runtime.Value, len(vals))
	arr.NextIndex = 0
	for idx, v := range vals {
		key := runtime.NewInt(int64(idx))
		arr.Keys[idx] = key
		arr.Elements[key] = v
		arr.NextIndex = int64(idx + 1)
	}

	return runtime.TRUE
}

// ----------------------------------------------------------------------------
// Additional math functions

func builtinSin(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewFloat(0)
	}
	return runtime.NewFloat(math.Sin(args[0].ToFloat()))
}

func builtinCos(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewFloat(0)
	}
	return runtime.NewFloat(math.Cos(args[0].ToFloat()))
}

func builtinTan(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewFloat(0)
	}
	return runtime.NewFloat(math.Tan(args[0].ToFloat()))
}

func builtinLog(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewFloat(0)
	}
	base := math.E
	if len(args) >= 2 {
		base = args[1].ToFloat()
	}
	if base == math.E {
		return runtime.NewFloat(math.Log(args[0].ToFloat()))
	}
	return runtime.NewFloat(math.Log(args[0].ToFloat()) / math.Log(base))
}

func builtinExp(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewFloat(0)
	}
	return runtime.NewFloat(math.Exp(args[0].ToFloat()))
}

// ----------------------------------------------------------------------------
// URL functions

func builtinParseUrl(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NULL
	}

	urlStr := args[0].ToString()
	u, err := url.Parse(urlStr)
	if err != nil {
		return runtime.FALSE
	}

	// If component is specified, return only that component
	if len(args) >= 2 {
		component := int(args[1].ToInt())
		switch component {
		case -1: // PHP_URL_SCHEME
			if u.Scheme != "" {
				return runtime.NewString(u.Scheme)
			}
			return runtime.NULL
		case 1: // PHP_URL_HOST
			if u.Host != "" {
				// Remove port if present
				host := u.Host
				if strings.Contains(host, ":") {
					host = strings.Split(host, ":")[0]
				}
				return runtime.NewString(host)
			}
			return runtime.NULL
		case 2: // PHP_URL_PORT
			if u.Port() != "" {
				port, _ := strconv.ParseInt(u.Port(), 10, 64)
				return runtime.NewInt(port)
			}
			return runtime.NULL
		case 3: // PHP_URL_USER
			if u.User != nil {
				return runtime.NewString(u.User.Username())
			}
			return runtime.NULL
		case 4: // PHP_URL_PASS
			if u.User != nil {
				if pass, ok := u.User.Password(); ok {
					return runtime.NewString(pass)
				}
			}
			return runtime.NULL
		case 5: // PHP_URL_PATH
			if u.Path != "" {
				return runtime.NewString(u.Path)
			}
			return runtime.NULL
		case 6: // PHP_URL_QUERY
			if u.RawQuery != "" {
				return runtime.NewString(u.RawQuery)
			}
			return runtime.NULL
		case 7: // PHP_URL_FRAGMENT
			if u.Fragment != "" {
				return runtime.NewString(u.Fragment)
			}
			return runtime.NULL
		}
		return runtime.NULL
	}

	// Return associative array with all components
	result := runtime.NewArray()

	if u.Scheme != "" {
		result.Set(runtime.NewString("scheme"), runtime.NewString(u.Scheme))
	}
	if u.Host != "" {
		// Remove port if present
		host := u.Host
		if strings.Contains(host, ":") {
			host = strings.Split(host, ":")[0]
		}
		result.Set(runtime.NewString("host"), runtime.NewString(host))
	}
	if u.Port() != "" {
		port, _ := strconv.ParseInt(u.Port(), 10, 64)
		result.Set(runtime.NewString("port"), runtime.NewInt(port))
	}
	if u.User != nil {
		result.Set(runtime.NewString("user"), runtime.NewString(u.User.Username()))
		if pass, ok := u.User.Password(); ok {
			result.Set(runtime.NewString("pass"), runtime.NewString(pass))
		}
	}
	if u.Path != "" {
		result.Set(runtime.NewString("path"), runtime.NewString(u.Path))
	}
	if u.RawQuery != "" {
		result.Set(runtime.NewString("query"), runtime.NewString(u.RawQuery))
	}
	if u.Fragment != "" {
		result.Set(runtime.NewString("fragment"), runtime.NewString(u.Fragment))
	}

	return result
}

func builtinHttpBuildQuery(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}

	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewString("")
	}

	var parts []string
	for _, key := range arr.Keys {
		val := arr.Elements[key]
		keyStr := key.ToString()
		valStr := val.ToString()
		parts = append(parts, url.QueryEscape(keyStr)+"="+url.QueryEscape(valStr))
	}

	return runtime.NewString(strings.Join(parts, "&"))
}

func builtinUrlencode(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	// PHP's urlencode uses + for spaces
	encoded := url.QueryEscape(args[0].ToString())
	return runtime.NewString(encoded)
}

func builtinUrldecode(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	decoded, err := url.QueryUnescape(args[0].ToString())
	if err != nil {
		return runtime.NewString(args[0].ToString())
	}
	return runtime.NewString(decoded)
}

func builtinRawurlencode(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	// PHP's rawurlencode uses %20 for spaces (RFC 3986)
	encoded := url.PathEscape(args[0].ToString())
	return runtime.NewString(encoded)
}

func builtinRawurldecode(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	decoded, err := url.PathUnescape(args[0].ToString())
	if err != nil {
		return runtime.NewString(args[0].ToString())
	}
	return runtime.NewString(decoded)
}

func (i *Interpreter) builtinParseStr(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NULL
	}

	queryStr := args[0].ToString()
	values, err := url.ParseQuery(queryStr)
	if err != nil {
		return runtime.FALSE
	}

	result := runtime.NewArray()
	for key, vals := range values {
		if len(vals) == 1 {
			result.Set(runtime.NewString(key), runtime.NewString(vals[0]))
		} else {
			// Multiple values - create array
			valsArray := runtime.NewArray()
			for _, v := range vals {
				valsArray.Set(nil, runtime.NewString(v))
			}
			result.Set(runtime.NewString(key), valsArray)
		}
	}

	// If second argument provided, assign to that variable (needs env access)
	// For now, just return the array
	return result
}

// ----------------------------------------------------------------------------
// Object/Class introspection functions

func builtinGetClass(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	obj, ok := args[0].(*runtime.Object)
	if !ok {
		return runtime.FALSE
	}

	return runtime.NewString(obj.Class.Name)
}

func builtinGetParentClass(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	// Can accept object or class name string
	var class *runtime.Class
	switch v := args[0].(type) {
	case *runtime.Object:
		class = v.Class
	case *runtime.String:
		// TODO: Look up class by name from environment
		// For now, return false
		return runtime.FALSE
	default:
		return runtime.FALSE
	}

	if class.Parent != nil {
		return runtime.NewString(class.Parent.Name)
	}

	return runtime.FALSE
}

func builtinGetClassMethods(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	var class *runtime.Class
	switch v := args[0].(type) {
	case *runtime.Object:
		class = v.Class
	case *runtime.String:
		// TODO: Look up class by name from environment
		return runtime.FALSE
	default:
		return runtime.FALSE
	}

	// Use map to track unique method names
	methodSet := make(map[string]bool)

	// Add methods from current class
	for methodName := range class.Methods {
		methodSet[methodName] = true
	}

	// Add inherited methods (won't duplicate due to map)
	current := class.Parent
	for current != nil {
		for methodName := range current.Methods {
			methodSet[methodName] = true
		}
		current = current.Parent
	}

	// Build result array
	result := runtime.NewArray()
	for methodName := range methodSet {
		result.Set(nil, runtime.NewString(methodName))
	}

	return result
}

func builtinMethodExists(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}

	var class *runtime.Class
	switch v := args[0].(type) {
	case *runtime.Object:
		class = v.Class
	case *runtime.String:
		// TODO: Look up class by name from environment
		return runtime.FALSE
	default:
		return runtime.FALSE
	}

	methodName := args[1].ToString()

	// Check in current class
	if _, ok := class.Methods[methodName]; ok {
		return runtime.TRUE
	}

	// Check in parent classes
	current := class.Parent
	for current != nil {
		if _, ok := current.Methods[methodName]; ok {
			return runtime.TRUE
		}
		current = current.Parent
	}

	return runtime.FALSE
}

func builtinPropertyExists(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}

	propertyName := args[1].ToString()

	switch v := args[0].(type) {
	case *runtime.Object:
		// Check instance properties
		if _, ok := v.Properties[propertyName]; ok {
			return runtime.TRUE
		}
		// Check class-defined properties
		if _, ok := v.Class.Properties[propertyName]; ok {
			return runtime.TRUE
		}
		// Check parent class properties
		current := v.Class.Parent
		for current != nil {
			if _, ok := current.Properties[propertyName]; ok {
				return runtime.TRUE
			}
			current = current.Parent
		}
	case *runtime.String:
		// TODO: Look up class by name and check static properties
		return runtime.FALSE
	default:
		return runtime.FALSE
	}

	return runtime.FALSE
}

func (i *Interpreter) builtinIsSubclassOf(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}

	var class *runtime.Class
	switch v := args[0].(type) {
	case *runtime.Object:
		class = v.Class
	case *runtime.String:
		// TODO: Look up class by name from environment
		return runtime.FALSE
	default:
		return runtime.FALSE
	}

	parentName := args[1].ToString()

	// Walk up the inheritance chain
	current := class.Parent
	for current != nil {
		if current.Name == parentName {
			return runtime.TRUE
		}
		current = current.Parent
	}

	return runtime.FALSE
}

func (i *Interpreter) builtinIsA(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}

	var class *runtime.Class
	switch v := args[0].(type) {
	case *runtime.Object:
		class = v.Class
	case *runtime.String:
		// TODO: Look up class by name from environment
		return runtime.FALSE
	default:
		return runtime.FALSE
	}

	className := args[1].ToString()

	// Check if it's the same class
	if class.Name == className {
		return runtime.TRUE
	}

	// Walk up the inheritance chain
	current := class.Parent
	for current != nil {
		if current.Name == className {
			return runtime.TRUE
		}
		current = current.Parent
	}

	// Check interfaces
	for _, iface := range class.Interfaces {
		if iface.Name == className {
			return runtime.TRUE
		}
	}

	return runtime.FALSE
}

// ----------------------------------------------------------------------------
// Additional string functions

func builtinStrstr(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}

	haystack := args[0].ToString()
	needle := args[1].ToString()

	if needle == "" {
		return runtime.FALSE
	}

	beforeNeedle := false
	if len(args) >= 3 {
		beforeNeedle = args[2].ToBool()
	}

	idx := strings.Index(haystack, needle)
	if idx == -1 {
		return runtime.FALSE
	}

	if beforeNeedle {
		return runtime.NewString(haystack[:idx])
	}
	return runtime.NewString(haystack[idx:])
}

func builtinStrrchr(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}

	haystack := args[0].ToString()
	needle := args[1].ToString()

	if needle == "" {
		return runtime.FALSE
	}

	// Find last occurrence (use first character of needle for single-char search)
	char := needle[0]
	idx := strings.LastIndexByte(haystack, char)
	if idx == -1 {
		return runtime.FALSE
	}

	return runtime.NewString(haystack[idx:])
}

func builtinSubstrCount(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewInt(0)
	}

	haystack := args[0].ToString()
	needle := args[1].ToString()

	if needle == "" {
		return runtime.NewInt(0)
	}

	offset := 0
	length := len(haystack)

	if len(args) >= 3 {
		offset = int(args[2].ToInt())
		if offset < 0 {
			offset = len(haystack) + offset
		}
		if offset < 0 || offset >= len(haystack) {
			return runtime.NewInt(0)
		}
	}

	if len(args) >= 4 {
		length = int(args[3].ToInt())
		if length < 0 {
			length = len(haystack) - offset + length
		}
		if offset+length > len(haystack) {
			length = len(haystack) - offset
		}
	}

	substring := haystack[offset : offset+length]
	count := strings.Count(substring, needle)
	return runtime.NewInt(int64(count))
}

func builtinSubstrCompare(args ...runtime.Value) runtime.Value {
	if len(args) < 3 {
		return runtime.FALSE
	}

	mainStr := args[0].ToString()
	str := args[1].ToString()
	offset := int(args[2].ToInt())

	if offset < 0 || offset >= len(mainStr) {
		return runtime.FALSE
	}

	length := len(str)
	if len(args) >= 4 {
		length = int(args[3].ToInt())
	}

	caseInsensitive := false
	if len(args) >= 5 {
		caseInsensitive = args[4].ToBool()
	}

	// Extract substring from mainStr starting at offset
	if offset+length > len(mainStr) {
		length = len(mainStr) - offset
	}
	substring := mainStr[offset : offset+length]

	// Limit str to length as well
	if length < len(str) {
		str = str[:length]
	}

	if caseInsensitive {
		substring = strings.ToLower(substring)
		str = strings.ToLower(str)
	}

	if substring == str {
		return runtime.NewInt(0)
	} else if substring < str {
		return runtime.NewInt(-1)
	}
	return runtime.NewInt(1)
}

func builtinStrtr(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewString("")
	}

	str := args[0].ToString()

	// Two argument form: replace pairs from associative array
	if len(args) == 2 {
		if replaceArray, ok := args[1].(*runtime.Array); ok {
			result := str
			for _, key := range replaceArray.Keys {
				from := key.ToString()
				to := replaceArray.Elements[key].ToString()
				result = strings.ReplaceAll(result, from, to)
			}
			return runtime.NewString(result)
		}
	}

	// Three argument form: character-by-character replacement
	if len(args) >= 3 {
		from := args[1].ToString()
		to := args[2].ToString()

		// Build replacement map
		replacer := make(map[rune]rune)
		fromRunes := []rune(from)
		toRunes := []rune(to)

		for i, r := range fromRunes {
			if i < len(toRunes) {
				replacer[r] = toRunes[i]
			} else {
				replacer[r] = toRunes[len(toRunes)-1]
			}
		}

		result := strings.Map(func(r rune) rune {
			if replacement, ok := replacer[r]; ok {
				return replacement
			}
			return r
		}, str)

		return runtime.NewString(result)
	}

	return runtime.NewString(str)
}

func builtinStrIreplace(args ...runtime.Value) runtime.Value {
	if len(args) < 3 {
		return runtime.NewString("")
	}

	search := args[0].ToString()
	replace := args[1].ToString()
	subject := args[2].ToString()

	// Case-insensitive replace
	lowerSubject := strings.ToLower(subject)
	lowerSearch := strings.ToLower(search)

	var result strings.Builder
	lastIdx := 0

	for {
		idx := strings.Index(lowerSubject[lastIdx:], lowerSearch)
		if idx == -1 {
			result.WriteString(subject[lastIdx:])
			break
		}

		actualIdx := lastIdx + idx
		result.WriteString(subject[lastIdx:actualIdx])
		result.WriteString(replace)
		lastIdx = actualIdx + len(search)
	}

	return runtime.NewString(result.String())
}

// ----------------------------------------------------------------------------
// Additional array functions

func builtinAsort(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}

	// Sort by value, maintaining key association
	type kvPair struct {
		key runtime.Value
		val runtime.Value
	}

	pairs := make([]kvPair, 0, len(arr.Keys))
	for _, k := range arr.Keys {
		pairs = append(pairs, kvPair{k, arr.Elements[k]})
	}

	sort.Slice(pairs, func(i, j int) bool {
		vi := pairs[i].val.ToString()
		vj := pairs[j].val.ToString()
		return vi < vj
	})

	// Rebuild array with new order
	arr.Keys = make([]runtime.Value, 0, len(pairs))
	for _, p := range pairs {
		arr.Keys = append(arr.Keys, p.key)
	}

	return runtime.TRUE
}

func builtinArsort(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}

	// Reverse sort by value, maintaining key association
	type kvPair struct {
		key runtime.Value
		val runtime.Value
	}

	pairs := make([]kvPair, 0, len(arr.Keys))
	for _, k := range arr.Keys {
		pairs = append(pairs, kvPair{k, arr.Elements[k]})
	}

	sort.Slice(pairs, func(i, j int) bool {
		vi := pairs[i].val.ToString()
		vj := pairs[j].val.ToString()
		return vi > vj
	})

	// Rebuild array with new order
	arr.Keys = make([]runtime.Value, 0, len(pairs))
	for _, p := range pairs {
		arr.Keys = append(arr.Keys, p.key)
	}

	return runtime.TRUE
}

func builtinKsort(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}

	// Sort by key
	sort.Slice(arr.Keys, func(i, j int) bool {
		ki := arr.Keys[i].ToString()
		kj := arr.Keys[j].ToString()
		return ki < kj
	})

	return runtime.TRUE
}

func builtinKrsort(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}

	// Reverse sort by key
	sort.Slice(arr.Keys, func(i, j int) bool {
		ki := arr.Keys[i].ToString()
		kj := arr.Keys[j].ToString()
		return ki > kj
	})

	return runtime.TRUE
}

func builtinArraySplice(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewArray()
	}

	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewArray()
	}

	offset := int(args[1].ToInt())
	length := len(arr.Keys)

	// Handle negative offset
	if offset < 0 {
		offset = len(arr.Keys) + offset
		if offset < 0 {
			offset = 0
		}
	}

	// Handle length parameter
	if len(args) >= 3 {
		length = int(args[2].ToInt())
		if length < 0 {
			length = len(arr.Keys) - offset + length
		}
	} else {
		length = len(arr.Keys) - offset
	}

	if offset >= len(arr.Keys) {
		return runtime.NewArray()
	}

	if offset+length > len(arr.Keys) {
		length = len(arr.Keys) - offset
	}

	// Extract removed elements
	removed := runtime.NewArray()
	for i := offset; i < offset+length && i < len(arr.Keys); i++ {
		key := arr.Keys[i]
		removed.Set(nil, arr.Elements[key])
	}

	// Build replacement array if provided
	type replPair struct {
		key runtime.Value
		val runtime.Value
	}
	var replacement []replPair
	if len(args) >= 4 {
		if replArr, ok := args[3].(*runtime.Array); ok {
			for _, k := range replArr.Keys {
				replacement = append(replacement, replPair{k, replArr.Elements[k]})
			}
		}
	}

	// Remove elements from original array
	for i := offset; i < offset+length && i < len(arr.Keys); i++ {
		key := arr.Keys[i]
		delete(arr.Elements, key)
	}

	// Build new keys slice
	newKeys := make([]runtime.Value, 0)
	newKeys = append(newKeys, arr.Keys[:offset]...)

	// Add replacement elements
	nextIdx := arr.NextIndex
	for _, p := range replacement {
		newKey := runtime.NewInt(nextIdx)
		arr.Elements[newKey] = p.val
		newKeys = append(newKeys, newKey)
		nextIdx++
	}

	// Add remaining elements
	if offset+length < len(arr.Keys) {
		newKeys = append(newKeys, arr.Keys[offset+length:]...)
	}

	arr.Keys = newKeys
	arr.NextIndex = nextIdx

	return removed
}

// ----------------------------------------------------------------------------
// File stream functions

func (i *Interpreter) builtinFopen(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}

	filename := args[0].ToString()
	mode := args[1].ToString()

	// Open file with appropriate mode
	var flag int
	switch mode {
	case "r":
		flag = os.O_RDONLY
	case "r+":
		flag = os.O_RDWR
	case "w":
		flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	case "w+":
		flag = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	case "a":
		flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	case "a+":
		flag = os.O_RDWR | os.O_CREATE | os.O_APPEND
	default:
		return runtime.FALSE
	}

	file, err := os.OpenFile(filename, flag, 0666)
	if err != nil {
		return runtime.FALSE
	}

	// Create resource
	resID := i.nextResourceID
	i.nextResourceID++
	resource := runtime.NewResource("stream", file, resID)
	i.resources[resID] = resource

	return resource
}

func builtinFclose(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	res, ok := args[0].(*runtime.Resource)
	if !ok {
		return runtime.FALSE
	}

	if file, ok := res.Handle.(*os.File); ok {
		err := file.Close()
		if err != nil {
			return runtime.FALSE
		}
		return runtime.TRUE
	}

	return runtime.FALSE
}

func builtinFread(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}

	res, ok := args[0].(*runtime.Resource)
	if !ok {
		return runtime.FALSE
	}

	length := int(args[1].ToInt())
	if length <= 0 {
		return runtime.NewString("")
	}

	if file, ok := res.Handle.(*os.File); ok {
		buf := make([]byte, length)
		n, err := file.Read(buf)
		if err != nil && n == 0 {
			return runtime.FALSE
		}
		return runtime.NewString(string(buf[:n]))
	}

	return runtime.FALSE
}

func builtinFwrite(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}

	res, ok := args[0].(*runtime.Resource)
	if !ok {
		return runtime.FALSE
	}

	data := args[1].ToString()
	length := len(data)

	if len(args) >= 3 {
		length = int(args[2].ToInt())
		if length > len(data) {
			length = len(data)
		}
	}

	if file, ok := res.Handle.(*os.File); ok {
		n, err := file.Write([]byte(data[:length]))
		if err != nil {
			return runtime.FALSE
		}
		return runtime.NewInt(int64(n))
	}

	return runtime.FALSE
}

func builtinFgets(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	res, ok := args[0].(*runtime.Resource)
	if !ok {
		return runtime.FALSE
	}

	if file, ok := res.Handle.(*os.File); ok {
		// Read until newline or EOF
		var line []byte
		buf := make([]byte, 1)
		for {
			n, err := file.Read(buf)
			if n == 0 || err != nil {
				break
			}
			line = append(line, buf[0])
			if buf[0] == '\n' {
				break
			}
		}
		if len(line) == 0 {
			return runtime.FALSE
		}
		return runtime.NewString(string(line))
	}

	return runtime.FALSE
}

func builtinFeof(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	res, ok := args[0].(*runtime.Resource)
	if !ok {
		return runtime.TRUE
	}

	if file, ok := res.Handle.(*os.File); ok {
		// Try to read one byte and seek back
		pos, err := file.Seek(0, io.SeekCurrent)
		if err != nil {
			return runtime.TRUE
		}

		buf := make([]byte, 1)
		n, _ := file.Read(buf)
		file.Seek(pos, io.SeekStart)

		if n == 0 {
			return runtime.TRUE
		}
		return runtime.FALSE
	}

	return runtime.TRUE
}

func builtinFseek(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewInt(-1)
	}

	res, ok := args[0].(*runtime.Resource)
	if !ok {
		return runtime.NewInt(-1)
	}

	offset := args[1].ToInt()
	whence := io.SeekStart

	if len(args) >= 3 {
		w := int(args[2].ToInt())
		switch w {
		case 1:
			whence = io.SeekCurrent
		case 2:
			whence = io.SeekEnd
		}
	}

	if file, ok := res.Handle.(*os.File); ok {
		_, err := file.Seek(offset, whence)
		if err != nil {
			return runtime.NewInt(-1)
		}
		return runtime.NewInt(0)
	}

	return runtime.NewInt(-1)
}

func builtinFtell(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	res, ok := args[0].(*runtime.Resource)
	if !ok {
		return runtime.FALSE
	}

	if file, ok := res.Handle.(*os.File); ok {
		pos, err := file.Seek(0, io.SeekCurrent)
		if err != nil {
			return runtime.FALSE
		}
		return runtime.NewInt(pos)
	}

	return runtime.FALSE
}

func builtinRewind(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	res, ok := args[0].(*runtime.Resource)
	if !ok {
		return runtime.FALSE
	}

	if file, ok := res.Handle.(*os.File); ok {
		_, err := file.Seek(0, io.SeekStart)
		if err != nil {
			return runtime.FALSE
		}
		return runtime.TRUE
	}

	return runtime.FALSE
}

func builtinReadfile(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	filename, ok := args[0].(*runtime.String)
	if !ok {
		return runtime.FALSE
	}

	content, err := os.ReadFile(filename.Value)
	if err != nil {
		return runtime.FALSE
	}

	fmt.Print(string(content))
	return runtime.NewInt(int64(len(content)))
}

func builtinFgetcsv(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	res, ok := args[0].(*runtime.Resource)
	if !ok {
		return runtime.FALSE
	}

	file, ok := res.Handle.(*os.File)
	if !ok {
		return runtime.FALSE
	}

	// Default parameters
	delimiter := ','

	if len(args) >= 3 {
		if delim, ok := args[2].(*runtime.String); ok && len(delim.Value) > 0 {
			delimiter = rune(delim.Value[0])
		}
	}

	reader := csv.NewReader(file)
	reader.Comma = delimiter
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = false

	record, err := reader.Read()
	if err == io.EOF {
		return runtime.FALSE
	}
	if err != nil {
		return runtime.FALSE
	}

	// Convert string slice to PHP array
	arr := runtime.NewArray()
	for _, field := range record {
		arr.Set(nil, runtime.NewString(field))
	}

	return arr
}

func builtinFputcsv(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}

	res, ok := args[0].(*runtime.Resource)
	if !ok {
		return runtime.FALSE
	}

	file, ok := res.Handle.(*os.File)
	if !ok {
		return runtime.FALSE
	}

	fields, ok := args[1].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}

	// Default parameters
	delimiter := ','

	if len(args) >= 3 {
		if delim, ok := args[2].(*runtime.String); ok && len(delim.Value) > 0 {
			delimiter = rune(delim.Value[0])
		}
	}

	writer := csv.NewWriter(file)
	writer.Comma = delimiter

	// Convert PHP array to string slice
	var record []string
	for _, key := range fields.Keys {
		val := fields.Elements[key]
		switch v := val.(type) {
		case *runtime.String:
			record = append(record, v.Value)
		case *runtime.Int:
			record = append(record, fmt.Sprintf("%d", v.Value))
		case *runtime.Float:
			record = append(record, fmt.Sprintf("%g", v.Value))
		case *runtime.Bool:
			if v.Value {
				record = append(record, "1")
			} else {
				record = append(record, "")
			}
		default:
			record = append(record, "")
		}
	}

	err := writer.Write(record)
	if err != nil {
		return runtime.FALSE
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return runtime.FALSE
	}

	return runtime.TRUE
}

func builtinUnlink(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	filename := args[0].ToString()
	err := os.Remove(filename)
	if err != nil {
		return runtime.FALSE
	}
	return runtime.TRUE
}

func builtinCopy(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}

	source := args[0].ToString()
	dest := args[1].ToString()

	// Read source file
	input, err := os.ReadFile(source)
	if err != nil {
		return runtime.FALSE
	}

	// Write to destination
	err = os.WriteFile(dest, input, 0644)
	if err != nil {
		return runtime.FALSE
	}

	return runtime.TRUE
}

func builtinRename(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}

	oldpath := args[0].ToString()
	newpath := args[1].ToString()

	err := os.Rename(oldpath, newpath)
	if err != nil {
		return runtime.FALSE
	}

	return runtime.TRUE
}

func builtinChmod(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}

	filename := args[0].ToString()
	mode := args[1].ToInt()

	err := os.Chmod(filename, os.FileMode(mode))
	if err != nil {
		return runtime.FALSE
	}

	return runtime.TRUE
}

func builtinTouch(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	filename := args[0].ToString()

	// Check if file exists
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		// Create empty file
		file, err := os.Create(filename)
		if err != nil {
			return runtime.FALSE
		}
		file.Close()
	} else {
		// Update modification time
		now := time.Now()
		err = os.Chtimes(filename, now, now)
		if err != nil {
			return runtime.FALSE
		}
	}

	return runtime.TRUE
}

// ----------------------------------------------------------------------------
// Directory functions

func builtinMkdir(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	pathname := args[0].ToString()
	mode := int64(0777)

	if len(args) >= 2 {
		mode = args[1].ToInt()
	}

	recursive := false
	if len(args) >= 3 {
		recursive = args[2].ToBool()
	}

	var err error
	if recursive {
		err = os.MkdirAll(pathname, os.FileMode(mode))
	} else {
		err = os.Mkdir(pathname, os.FileMode(mode))
	}

	if err != nil {
		return runtime.FALSE
	}
	return runtime.TRUE
}

func builtinRmdir(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	dirname := args[0].ToString()
	err := os.Remove(dirname)
	if err != nil {
		return runtime.FALSE
	}
	return runtime.TRUE
}

func builtinScandir(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	dir := args[0].ToString()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return runtime.FALSE
	}

	result := runtime.NewArray()
	for _, entry := range entries {
		result.Set(nil, runtime.NewString(entry.Name()))
	}

	return result
}

func (i *Interpreter) builtinChdir(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	dir := args[0].ToString()
	err := os.Chdir(dir)
	if err != nil {
		return runtime.FALSE
	}

	// Update interpreter's current directory
	newCwd, _ := os.Getwd()
	i.currentDir = newCwd

	return runtime.TRUE
}

func (i *Interpreter) builtinGetcwd(args ...runtime.Value) runtime.Value {
	cwd, err := os.Getwd()
	if err != nil {
		return runtime.FALSE
	}
	return runtime.NewString(cwd)
}

// ----------------------------------------------------------------------------
// Variable handling functions

func (i *Interpreter) builtinCompact(args ...runtime.Value) runtime.Value {
	result := runtime.NewArray()

	for _, arg := range args {
		varName := arg.ToString()
		// Try to get variable from environment
		if val, ok := i.env.Get(varName); ok {
			result.Set(runtime.NewString(varName), val)
		}
	}

	return result
}

func (i *Interpreter) builtinExtract(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewInt(0)
	}

	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewInt(0)
	}

	extractType := int64(0) // EXTR_OVERWRITE by default
	if len(args) >= 2 {
		extractType = args[1].ToInt()
	}

	count := int64(0)
	for _, key := range arr.Keys {
		varName := key.ToString()
		value := arr.Elements[key]

		// Check if variable exists
		_, exists := i.env.Get(varName)

		switch extractType {
		case 0: // EXTR_OVERWRITE - overwrite existing variables (default)
			i.env.Set(varName, value)
			count++
		case 1: // EXTR_SKIP - skip existing variables
			if !exists {
				i.env.Set(varName, value)
				count++
			}
		default:
			i.env.Set(varName, value)
			count++
		}
	}

	return runtime.NewInt(count)
}

func builtinArrayPad(args ...runtime.Value) runtime.Value {
	if len(args) < 3 {
		return runtime.NewArray()
	}

	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewArray()
	}

	size := int(args[1].ToInt())
	value := args[2]

	result := runtime.NewArray()

	// Copy existing elements
	for _, key := range arr.Keys {
		result.Set(key, arr.Elements[key])
	}

	currentLen := len(arr.Keys)
	absSize := size
	if absSize < 0 {
		absSize = -absSize
	}

	if currentLen >= absSize {
		return result
	}

	// Pad at the end (positive size) or beginning (negative size)
	padCount := absSize - currentLen

	if size > 0 {
		// Pad at the end
		for i := 0; i < padCount; i++ {
			result.Set(nil, value)
		}
	} else {
		// Pad at the beginning
		newKeys := make([]runtime.Value, 0, absSize)
		newElements := make(map[runtime.Value]runtime.Value)

		// Add padding values first
		for i := 0; i < padCount; i++ {
			key := runtime.NewInt(int64(i))
			newKeys = append(newKeys, key)
			newElements[key] = value
		}

		// Add original elements
		idx := int64(padCount)
		for _, k := range result.Keys {
			newKey := runtime.NewInt(idx)
			newKeys = append(newKeys, newKey)
			newElements[newKey] = result.Elements[k]
			idx++
		}

		result.Keys = newKeys
		result.Elements = newElements
		result.NextIndex = idx
	}

	return result
}
