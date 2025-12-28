package interpreter

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"math"
	"mime/quotedprintable"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	goruntime "runtime"
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
	case "vprintf":
		return i.builtinVprintf
	case "str_repeat":
		return builtinStrRepeat
	case "substr_replace":
		return builtinSubstrReplace
	case "count_chars":
		return builtinCountChars
	case "sscanf":
		return builtinSscanf
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
	case "str_word_count":
		return builtinStrWordCount
	case "str_shuffle":
		return builtinStrShuffle
	case "str_getcsv":
		return builtinStrGetcsv
	case "str_rot13":
		return builtinStrRot13
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
	case "array_key_first":
		return builtinArrayKeyFirst
	case "array_key_last":
		return builtinArrayKeyLast
	case "array_is_list":
		return builtinArrayIsList
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
	case "natsort":
		return builtinNatsort
	case "natcasesort":
		return builtinNatcasesort

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
	case "fdiv":
		return builtinFdiv
	case "intdiv":
		return builtinIntdiv
	case "fmod":
		return builtinFmod
	case "is_finite":
		return builtinIsFinite
	case "is_nan":
		return builtinIsNan
	case "is_infinite":
		return builtinIsInfinite
	case "rand":
		return builtinRand
	case "mt_rand":
		return builtinMtRand
	case "lcg_value":
		return builtinLcgValue

	// Type functions
	case "gettype":
		return builtinGettype
	case "settype":
		return builtinSettype
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
	case "is_callable":
		return i.builtinIsCallable
	case "filter_var":
		return builtinFilterVar
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
	case "var_export":
		return i.builtinVarExport

	// Error handling
	case "trigger_error":
		return i.builtinTriggerError
	case "error_reporting":
		return i.builtinErrorReporting

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
	case "define":
		return i.builtinDefine
	case "defined":
		return i.builtinDefined
	case "constant":
		return i.builtinConstant
	case "ini_get":
		return i.builtinIniGet
	case "ini_set":
		return i.builtinIniSet
	case "version_compare":
		return builtinVersionCompare
	case "phpversion":
		return builtinPhpversion
	case "extension_loaded":
		return builtinExtensionLoaded
	case "memory_get_usage":
		return builtinMemoryGetUsage
	case "memory_get_peak_usage":
		return builtinMemoryGetPeakUsage
	case "function_exists":
		return i.builtinFunctionExists
	case "class_exists":
		return i.builtinClassExists
	case "class_alias":
		return i.builtinClassAlias
	case "spl_autoload_register":
		return i.builtinSplAutoloadRegister
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
	case "getenv":
		return builtinGetenv
	case "putenv":
		return builtinPutenv
	case "parse_ini_file":
		return builtinParseIniFile
	case "parse_ini_string":
		return builtinParseIniString

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
	case "strftime":
		return builtinStrftime
	case "gmstrftime":
		return builtinGmstrftime
	case "gmdate":
		return builtinGmdate
	case "gmmktime":
		return builtinGmmktime
	case "getdate":
		return builtinGetdate
	case "checkdate":
		return builtinCheckdate
	case "idate":
		return builtinIdate

	// Hash functions
	case "md5":
		return builtinMd5
	case "sha1":
		return builtinSha1
	case "hash":
		return builtinHash
	case "crc32":
		return builtinCrc32
	case "hash_hmac":
		return builtinHashHmac
	case "hash_equals":
		return builtinHashEquals
	case "password_hash":
		return builtinPasswordHash
	case "password_verify":
		return builtinPasswordVerify
	case "base64_encode":
		return builtinBase64Encode
	case "base64_decode":
		return builtinBase64Decode
	case "bin2hex":
		return builtinBin2hex
	case "hex2bin":
		return builtinHex2bin
	case "quoted_printable_encode":
		return builtinQuotedPrintableEncode
	case "quoted_printable_decode":
		return builtinQuotedPrintableDecode
	case "convert_uuencode":
		return builtinConvertUuencode
	case "convert_uudecode":
		return builtinConvertUudecode

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
	case "htmlspecialchars_decode":
		return builtinHtmlspecialcharsDecode
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
	case "array_fill_keys":
		return builtinArrayFillKeys
	case "array_replace":
		return builtinArrayReplace
	case "array_chunk":
		return builtinArrayChunk
	case "array_column":
		return builtinArrayColumn
	case "array_count_values":
		return builtinArrayCountValues
	case "array_diff":
		return builtinArrayDiff
	case "array_diff_key":
		return builtinArrayDiffKey
	case "array_diff_assoc":
		return builtinArrayDiffAssoc
	case "array_intersect":
		return builtinArrayIntersect
	case "array_intersect_key":
		return builtinArrayIntersectKey
	case "array_intersect_assoc":
		return builtinArrayIntersectAssoc
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
	case "deg2rad":
		return builtinDeg2rad
	case "rad2deg":
		return builtinRad2deg
	case "hypot":
		return builtinHypot
	case "log10":
		return builtinLog10
	case "log1p":
		return builtinLog1p
	case "expm1":
		return builtinExpm1

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
	case "spl_object_hash":
		return builtinSplObjectHash
	case "spl_object_id":
		return builtinSplObjectId
	case "get_object_vars":
		return i.builtinGetObjectVars
	case "get_class_vars":
		return i.builtinGetClassVars

	// Additional string functions
	case "strstr", "strchr":
		return builtinStrstr
	case "strrchr":
		return builtinStrrchr
	case "mb_strlen":
		return builtinMbStrlen
	case "mb_substr":
		return builtinMbSubstr
	case "mb_strpos":
		return builtinMbStrpos
	case "mb_strtoupper":
		return builtinMbStrtoupper
	case "mb_strtolower":
		return builtinMbStrtolower
	case "substr_count":
		return builtinSubstrCount
	case "substr_compare":
		return builtinSubstrCompare
	case "strtr":
		return builtinStrtr
	case "str_ireplace":
		return builtinStrIreplace
	case "strpbrk":
		return builtinStrpbrk
	case "similar_text":
		return builtinSimilarText
	case "soundex":
		return builtinSoundex
	case "levenshtein":
		return builtinLevenshtein

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
	case "array_multisort":
		return builtinArrayMultisort
	case "array_change_key_case":
		return builtinArrayChangeKeyCase

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
	case "sys_get_temp_dir":
		return builtinSysGetTempDir
	case "tempnam":
		return builtinTempnam

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
	case "opendir":
		return i.builtinOpendir
	case "readdir":
		return builtinReaddir
	case "closedir":
		return builtinClosedir
	case "disk_free_space":
		return builtinDiskFreeSpace
	case "disk_total_space":
		return builtinDiskTotalSpace

	// Variable handling
	case "compact":
		return i.builtinCompact
	case "extract":
		return i.builtinExtract
	case "get_defined_vars":
		return i.builtinGetDefinedVars
	case "get_defined_constants":
		return i.builtinGetDefinedConstants
	case "array_pad":
		return builtinArrayPad

	// Network functions
	case "ip2long":
		return builtinIp2long
	case "long2ip":
		return builtinLong2ip
	case "gethostbyname":
		return builtinGethostbyname
	case "gethostbyaddr":
		return builtinGethostbyaddr
	case "inet_pton":
		return builtinInetPton
	case "inet_ntop":
		return builtinInetNtop

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

func (i *Interpreter) builtinVprintf(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewInt(0)
	}
	format := args[0].ToString()
	argsArray, ok := args[1].(*runtime.Array)
	if !ok {
		return runtime.NewInt(0)
	}

	// Convert array to interface slice
	fmtArgs := make([]interface{}, 0, len(argsArray.Keys))
	for _, key := range argsArray.Keys {
		val := argsArray.Elements[key]
		switch v := val.(type) {
		case *runtime.Int:
			fmtArgs = append(fmtArgs, v.Value)
		case *runtime.Float:
			fmtArgs = append(fmtArgs, v.Value)
		case *runtime.String:
			fmtArgs = append(fmtArgs, v.Value)
		default:
			fmtArgs = append(fmtArgs, val.ToString())
		}
	}

	output := fmt.Sprintf(format, fmtArgs...)
	i.writeOutput(output)
	return runtime.NewInt(int64(len(output)))
}

func builtinSubstrReplace(args ...runtime.Value) runtime.Value {
	if len(args) < 3 {
		return runtime.NewString("")
	}
	str := args[0].ToString()
	replacement := args[1].ToString()
	start := int(args[2].ToInt())

	length := len(str) - start
	if len(args) >= 4 {
		length = int(args[3].ToInt())
	}

	// Handle negative start
	if start < 0 {
		start = len(str) + start
		if start < 0 {
			start = 0
		}
	}

	// Handle out of bounds start
	if start > len(str) {
		return runtime.NewString(str)
	}

	// Calculate end position
	end := start + length
	if length < 0 {
		end = len(str) + length
	}
	if end > len(str) {
		end = len(str)
	}
	if end < start {
		end = start
	}

	result := str[:start] + replacement + str[end:]
	return runtime.NewString(result)
}

func builtinCountChars(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewArray()
	}
	str := args[0].ToString()
	mode := int64(0)
	if len(args) >= 2 {
		mode = args[1].ToInt()
	}

	counts := make(map[byte]int)
	for i := 0; i < len(str); i++ {
		counts[str[i]]++
	}

	result := runtime.NewArray()
	switch mode {
	case 0: // All bytes with count
		for i := 0; i < 256; i++ {
			result.Set(runtime.NewInt(int64(i)), runtime.NewInt(int64(counts[byte(i)])))
		}
	case 1: // Only bytes with count > 0
		for i := 0; i < 256; i++ {
			if counts[byte(i)] > 0 {
				result.Set(runtime.NewInt(int64(i)), runtime.NewInt(int64(counts[byte(i)])))
			}
		}
	case 2: // Only bytes with count == 0
		for i := 0; i < 256; i++ {
			if counts[byte(i)] == 0 {
				result.Set(runtime.NewInt(int64(i)), runtime.NewInt(int64(counts[byte(i)])))
			}
		}
	case 3: // All unique characters as string
		var chars []byte
		for i := 0; i < 256; i++ {
			if counts[byte(i)] > 0 {
				chars = append(chars, byte(i))
			}
		}
		return runtime.NewString(string(chars))
	case 4: // All unused characters as string
		var chars []byte
		for i := 0; i < 256; i++ {
			if counts[byte(i)] == 0 {
				chars = append(chars, byte(i))
			}
		}
		return runtime.NewString(string(chars))
	}

	return result
}

func builtinSscanf(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	str := args[0].ToString()
	format := args[1].ToString()

	// Simple implementation - parse basic format specifiers
	// %d = integer, %s = string, %f = float
	result := runtime.NewArray()

	// Split format into parts
	parts := strings.Split(format, "%")
	strIdx := 0

	for i := 1; i < len(parts); i++ {
		if len(parts[i]) == 0 {
			continue
		}

		spec := parts[i][0]
		// Skip to next non-whitespace in string
		for strIdx < len(str) && (str[strIdx] == ' ' || str[strIdx] == '\t') {
			strIdx++
		}

		switch spec {
		case 'd': // Integer
			numStr := ""
			for strIdx < len(str) && str[strIdx] >= '0' && str[strIdx] <= '9' {
				numStr += string(str[strIdx])
				strIdx++
			}
			if numStr != "" {
				val, _ := strconv.ParseInt(numStr, 10, 64)
				result.Set(nil, runtime.NewInt(val))
			}
		case 's': // String (until whitespace)
			token := ""
			for strIdx < len(str) && str[strIdx] != ' ' && str[strIdx] != '\t' {
				token += string(str[strIdx])
				strIdx++
			}
			if token != "" {
				result.Set(nil, runtime.NewString(token))
			}
		case 'f': // Float
			numStr := ""
			for strIdx < len(str) && (str[strIdx] >= '0' && str[strIdx] <= '9' || str[strIdx] == '.') {
				numStr += string(str[strIdx])
				strIdx++
			}
			if numStr != "" {
				val, _ := strconv.ParseFloat(numStr, 64)
				result.Set(nil, runtime.NewFloat(val))
			}
		}
	}

	return result
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

func builtinStrWordCount(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewInt(0)
	}

	str := args[0].ToString()
	format := int64(0) // Default format: return word count

	if len(args) >= 2 {
		format = args[1].ToInt()
	}

	// Split by whitespace and punctuation
	words := strings.FieldsFunc(str, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '\'' || r == '-')
	})

	// Filter empty strings
	var validWords []string
	for _, word := range words {
		if len(word) > 0 {
			validWords = append(validWords, word)
		}
	}

	switch format {
	case 0: // Return word count
		return runtime.NewInt(int64(len(validWords)))
	case 1: // Return array of words
		result := runtime.NewArray()
		for _, word := range validWords {
			result.Set(nil, runtime.NewString(word))
		}
		return result
	case 2: // Return associative array (position => word)
		result := runtime.NewArray()
		pos := 0
		currentPos := 0
		for _, word := range validWords {
			// Find position in original string
			idx := strings.Index(str[currentPos:], word)
			if idx >= 0 {
				pos = currentPos + idx
				result.Set(runtime.NewInt(int64(pos)), runtime.NewString(word))
				currentPos = pos + len(word)
			}
		}
		return result
	default:
		return runtime.NewInt(int64(len(validWords)))
	}
}

func builtinStrShuffle(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	str := args[0].ToString()
	runes := []rune(str)

	// Fisher-Yates shuffle
	for i := len(runes) - 1; i > 0; i-- {
		j := int(time.Now().UnixNano()%(int64(i)+1)) % (i + 1)
		runes[i], runes[j] = runes[j], runes[i]
	}

	return runtime.NewString(string(runes))
}

func builtinStrGetcsv(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	input := args[0].ToString()
	delimiter := ","
	enclosure := "\""
	escape := "\\"

	if len(args) >= 2 {
		delimiter = args[1].ToString()
		if len(delimiter) == 0 {
			delimiter = ","
		} else {
			delimiter = string(delimiter[0])
		}
	}

	if len(args) >= 3 {
		enclosure = args[2].ToString()
		if len(enclosure) == 0 {
			enclosure = "\""
		} else {
			enclosure = string(enclosure[0])
		}
	}

	if len(args) >= 4 {
		escape = args[3].ToString()
		if len(escape) == 0 {
			escape = "\\"
		} else {
			escape = string(escape[0])
		}
	}

	// Use csv.Reader
	reader := csv.NewReader(strings.NewReader(input))
	reader.Comma = rune(delimiter[0])
	reader.LazyQuotes = true

	if enclosure[0] == delimiter[0] {
		// If enclosure and delimiter are the same, disable quoting
		reader.LazyQuotes = true
	}

	record, err := reader.Read()
	if err != nil {
		return runtime.FALSE
	}

	result := runtime.NewArray()
	for _, field := range record {
		result.Set(nil, runtime.NewString(field))
	}

	return result
}

func builtinStrRot13(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	str := args[0].ToString()
	result := make([]rune, len(str))

	for i, ch := range str {
		switch {
		case ch >= 'a' && ch <= 'z':
			result[i] = 'a' + (ch-'a'+13)%26
		case ch >= 'A' && ch <= 'Z':
			result[i] = 'A' + (ch-'A'+13)%26
		default:
			result[i] = ch
		}
	}

	return runtime.NewString(string(result))
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

func builtinArrayKeyFirst(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NULL
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NULL
	}
	if len(arr.Keys) == 0 {
		return runtime.NULL
	}
	return arr.Keys[0]
}

func builtinArrayKeyLast(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NULL
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NULL
	}
	if len(arr.Keys) == 0 {
		return runtime.NULL
	}
	return arr.Keys[len(arr.Keys)-1]
}

func builtinArrayIsList(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}

	// An array is a list if it has sequential integer keys starting from 0
	for i, key := range arr.Keys {
		intKey, ok := key.(*runtime.Int)
		if !ok {
			return runtime.FALSE
		}
		if intKey.Value != int64(i) {
			return runtime.FALSE
		}
	}
	return runtime.TRUE
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

func builtinNatsort(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}

	// Create slice of key-value pairs
	type kv struct {
		key runtime.Value
		val runtime.Value
	}
	pairs := make([]kv, 0, len(arr.Keys))
	for _, key := range arr.Keys {
		pairs = append(pairs, kv{key, arr.Elements[key]})
	}

	// Natural sort by value
	sort.Slice(pairs, func(i, j int) bool {
		return naturalCompare(pairs[i].val.ToString(), pairs[j].val.ToString(), false) < 0
	})

	// Rebuild array maintaining original keys
	arr.Keys = make([]runtime.Value, len(pairs))
	arr.Elements = make(map[runtime.Value]runtime.Value)
	for i, p := range pairs {
		arr.Keys[i] = p.key
		arr.Elements[p.key] = p.val
	}

	return runtime.TRUE
}

func builtinNatcasesort(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}

	// Create slice of key-value pairs
	type kv struct {
		key runtime.Value
		val runtime.Value
	}
	pairs := make([]kv, 0, len(arr.Keys))
	for _, key := range arr.Keys {
		pairs = append(pairs, kv{key, arr.Elements[key]})
	}

	// Natural sort by value (case-insensitive)
	sort.Slice(pairs, func(i, j int) bool {
		return naturalCompare(pairs[i].val.ToString(), pairs[j].val.ToString(), true) < 0
	})

	// Rebuild array maintaining original keys
	arr.Keys = make([]runtime.Value, len(pairs))
	arr.Elements = make(map[runtime.Value]runtime.Value)
	for i, p := range pairs {
		arr.Keys[i] = p.key
		arr.Elements[p.key] = p.val
	}

	return runtime.TRUE
}

func naturalCompare(a, b string, caseInsensitive bool) int {
	if caseInsensitive {
		a = strings.ToLower(a)
		b = strings.ToLower(b)
	}

	ia, ib := 0, 0
	for ia < len(a) && ib < len(b) {
		// Check if both are at digit positions
		if a[ia] >= '0' && a[ia] <= '9' && b[ib] >= '0' && b[ib] <= '9' {
			// Extract numbers
			numA, numB := 0, 0
			for ia < len(a) && a[ia] >= '0' && a[ia] <= '9' {
				numA = numA*10 + int(a[ia]-'0')
				ia++
			}
			for ib < len(b) && b[ib] >= '0' && b[ib] <= '9' {
				numB = numB*10 + int(b[ib]-'0')
				ib++
			}
			if numA != numB {
				return numA - numB
			}
		} else {
			// Regular character comparison
			if a[ia] != b[ib] {
				return int(a[ia]) - int(b[ib])
			}
			ia++
			ib++
		}
	}

	return len(a) - len(b)
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

func builtinFdiv(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewFloat(0)
	}
	dividend := args[0].ToFloat()
	divisor := args[1].ToFloat()
	// fdiv performs floating-point division, handling division by zero as INF
	if divisor == 0 {
		if dividend == 0 {
			return runtime.NewFloat(math.NaN())
		}
		if dividend > 0 {
			return runtime.NewFloat(math.Inf(1))
		}
		return runtime.NewFloat(math.Inf(-1))
	}
	return runtime.NewFloat(dividend / divisor)
}

func builtinIntdiv(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewInt(0)
	}
	dividend := args[0].ToInt()
	divisor := args[1].ToInt()
	if divisor == 0 {
		// In PHP, intdiv throws DivisionByZeroError, here we return 0
		return runtime.NewInt(0)
	}
	return runtime.NewInt(dividend / divisor)
}

func builtinFmod(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewFloat(0)
	}
	x := args[0].ToFloat()
	y := args[1].ToFloat()
	return runtime.NewFloat(math.Mod(x, y))
}

func builtinIsFinite(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	f := args[0].ToFloat()
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return runtime.FALSE
	}
	return runtime.TRUE
}

func builtinIsNan(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	f := args[0].ToFloat()
	if math.IsNaN(f) {
		return runtime.TRUE
	}
	return runtime.FALSE
}

func builtinIsInfinite(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	f := args[0].ToFloat()
	if math.IsInf(f, 0) {
		return runtime.TRUE
	}
	return runtime.FALSE
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

func builtinLcgValue(args ...runtime.Value) runtime.Value {
	// Linear Congruential Generator - returns a pseudo random number between 0 and 1
	// Using time-based seed for randomness
	seed := uint64(time.Now().UnixNano())
	// LCG parameters (from Numerical Recipes)
	const a uint64 = 1664525
	const c uint64 = 1013904223
	const m uint64 = 1 << 32

	value := float64((a*seed+c)%m) / float64(m)
	return runtime.NewFloat(value)
}

// ----------------------------------------------------------------------------
// Type functions

func builtinGettype(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("NULL")
	}
	return runtime.NewString(args[0].Type())
}

func builtinSettype(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}

	// Note: In PHP, settype modifies the variable by reference
	// This implementation returns the converted value
	// For proper reference support, this would need deeper integration

	targetType := args[1].ToString()
	value := args[0]

	switch targetType {
	case "boolean", "bool":
		return runtime.NewBool(value.ToBool())
	case "integer", "int":
		return runtime.NewInt(value.ToInt())
	case "float", "double":
		return runtime.NewFloat(value.ToFloat())
	case "string":
		return runtime.NewString(value.ToString())
	case "array":
		arr := runtime.NewArray()
		arr.Set(nil, value)
		return arr
	case "null":
		return runtime.NULL
	default:
		return runtime.FALSE
	}
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

func (i *Interpreter) builtinIsCallable(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	value := args[0]

	// Check if it's a function
	if _, ok := value.(*runtime.Function); ok {
		return runtime.TRUE
	}

	// Check if it's a string referring to a function name
	if str, ok := value.(*runtime.String); ok {
		if _, exists := i.env.GetFunction(str.Value); exists {
			return runtime.TRUE
		}
	}

	// Could also check for callable arrays [object, method] or [class, method]
	// For now, keep it simple

	return runtime.FALSE
}

func builtinFilterVar(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NULL
	}

	value := args[0].ToString()
	filterType := int64(516) // FILTER_DEFAULT

	if len(args) >= 2 {
		filterType = args[1].ToInt()
	}

	switch filterType {
	case 257: // FILTER_VALIDATE_INT
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return runtime.FALSE
		}
		return runtime.NewInt(val)

	case 259: // FILTER_VALIDATE_FLOAT
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return runtime.FALSE
		}
		return runtime.NewFloat(val)

	case 273: // FILTER_VALIDATE_EMAIL
		// Simple email validation
		if strings.Contains(value, "@") && strings.Contains(value, ".") {
			return runtime.NewString(value)
		}
		return runtime.FALSE

	case 277: // FILTER_VALIDATE_URL
		// Simple URL validation
		if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
			return runtime.NewString(value)
		}
		return runtime.FALSE

	case 275: // FILTER_VALIDATE_IP
		// Simple IP validation
		parts := strings.Split(value, ".")
		if len(parts) == 4 {
			valid := true
			for _, part := range parts {
				num, err := strconv.Atoi(part)
				if err != nil || num < 0 || num > 255 {
					valid = false
					break
				}
			}
			if valid {
				return runtime.NewString(value)
			}
		}
		return runtime.FALSE

	case 272: // FILTER_VALIDATE_BOOLEAN
		lower := strings.ToLower(value)
		if lower == "1" || lower == "true" || lower == "on" || lower == "yes" {
			return runtime.TRUE
		}
		if lower == "0" || lower == "false" || lower == "off" || lower == "no" || lower == "" {
			return runtime.FALSE
		}
		return runtime.NULL

	case 513: // FILTER_SANITIZE_STRING
		// Remove HTML tags
		result := regexp.MustCompile(`<[^>]*>`).ReplaceAllString(value, "")
		return runtime.NewString(result)

	case 515: // FILTER_SANITIZE_EMAIL
		// Keep only valid email characters
		result := regexp.MustCompile(`[^a-zA-Z0-9@._+-]`).ReplaceAllString(value, "")
		return runtime.NewString(result)

	case 518: // FILTER_SANITIZE_NUMBER_INT
		// Keep only digits and signs
		result := regexp.MustCompile(`[^0-9+-]`).ReplaceAllString(value, "")
		return runtime.NewString(result)

	case 516: // FILTER_DEFAULT
		fallthrough
	default:
		return runtime.NewString(value)
	}
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

func (i *Interpreter) builtinVarExport(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NULL
	}

	returnOutput := false
	if len(args) >= 2 {
		returnOutput = args[1].ToBool()
	}

	output := i.exportValue(args[0], 0)
	if returnOutput {
		return runtime.NewString(output)
	}
	i.writeOutput(output)
	return runtime.NULL
}

func (i *Interpreter) exportValue(v runtime.Value, indent int) string {
	switch val := v.(type) {
	case *runtime.String:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(strings.ReplaceAll(val.Value, "\\", "\\\\"), "'", "\\'"))
	case *runtime.Int:
		return fmt.Sprintf("%d", val.Value)
	case *runtime.Float:
		return fmt.Sprintf("%g", val.Value)
	case *runtime.Bool:
		if val.Value {
			return "true"
		}
		return "false"
	case *runtime.Null:
		return "NULL"
	case *runtime.Array:
		if len(val.Keys) == 0 {
			return "array ()"
		}
		var sb strings.Builder
		sb.WriteString("array (\n")
		indentStr := strings.Repeat("  ", indent+1)
		for _, key := range val.Keys {
			sb.WriteString(indentStr)
			sb.WriteString(i.exportValue(key, indent+1))
			sb.WriteString(" => ")
			sb.WriteString(i.exportValue(val.Elements[key], indent+1))
			sb.WriteString(",\n")
		}
		sb.WriteString(strings.Repeat("  ", indent))
		sb.WriteString(")")
		return sb.String()
	default:
		return "NULL"
	}
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
// Error handling functions

func (i *Interpreter) builtinTriggerError(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	message := args[0].ToString()
	errorType := int64(1024) // E_USER_NOTICE by default

	if len(args) >= 2 {
		errorType = args[1].ToInt()
	}

	// Map error types to names
	errorName := "Notice"
	switch errorType {
	case 256: // E_USER_ERROR
		errorName = "Fatal error"
	case 512: // E_USER_WARNING
		errorName = "Warning"
	case 1024: // E_USER_NOTICE
		errorName = "Notice"
	case 2048: // E_USER_DEPRECATED
		errorName = "Deprecated"
	}

	// Output the error message
	output := fmt.Sprintf("PHP %s: %s\n", errorName, message)
	i.writeOutput(output)

	return runtime.TRUE
}

func (i *Interpreter) builtinErrorReporting(args ...runtime.Value) runtime.Value {
	// Get current error reporting level from ini settings
	currentLevel := i.iniSettings["error_reporting"]
	currentInt, _ := strconv.ParseInt(currentLevel, 10, 64)

	// If no argument, return current level
	if len(args) == 0 {
		return runtime.NewInt(currentInt)
	}

	// Set new error reporting level
	newLevel := args[0].ToInt()
	i.iniSettings["error_reporting"] = strconv.FormatInt(newLevel, 10)

	// Return old level
	return runtime.NewInt(currentInt)
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

func (i *Interpreter) builtinDefine(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}

	name := args[0].ToString()
	value := args[1]

	// Check if constant already exists
	if _, ok := i.env.GetConstant(name); ok {
		return runtime.FALSE
	}

	// Define the constant
	i.env.DefineConstant(name, value)
	return runtime.TRUE
}

func (i *Interpreter) builtinDefined(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	name := args[0].ToString()
	_, ok := i.env.GetConstant(name)
	return runtime.NewBool(ok)
}

func (i *Interpreter) builtinConstant(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NULL
	}

	name := args[0].ToString()
	value, ok := i.env.GetConstant(name)
	if !ok {
		return runtime.NULL
	}

	return value
}

func (i *Interpreter) builtinIniGet(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	name := args[0].ToString()
	if value, ok := i.iniSettings[name]; ok {
		return runtime.NewString(value)
	}

	return runtime.FALSE
}

func (i *Interpreter) builtinIniSet(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}

	name := args[0].ToString()
	newValue := args[1].ToString()

	// Get old value
	oldValue := ""
	if val, ok := i.iniSettings[name]; ok {
		oldValue = val
	}

	// Set new value
	i.iniSettings[name] = newValue

	// Return old value or false if it didn't exist
	if oldValue != "" {
		return runtime.NewString(oldValue)
	}
	return runtime.FALSE
}

func builtinVersionCompare(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NULL
	}

	version1 := args[0].ToString()
	version2 := args[1].ToString()

	// Parse versions
	parts1 := parseVersion(version1)
	parts2 := parseVersion(version2)

	// Compare versions
	result := compareVersionParts(parts1, parts2)

	// If operator is provided, return bool based on comparison
	if len(args) >= 3 {
		operator := args[2].ToString()
		switch operator {
		case "<", "lt":
			return runtime.NewBool(result < 0)
		case "<=", "le":
			return runtime.NewBool(result <= 0)
		case ">", "gt":
			return runtime.NewBool(result > 0)
		case ">=", "ge":
			return runtime.NewBool(result >= 0)
		case "==", "=", "eq":
			return runtime.NewBool(result == 0)
		case "!=", "<>", "ne":
			return runtime.NewBool(result != 0)
		}
	}

	// Return numeric comparison result
	return runtime.NewInt(int64(result))
}

func parseVersion(version string) []int {
	// Remove common version prefixes/suffixes
	version = strings.ToLower(version)
	version = strings.ReplaceAll(version, "v", "")
	version = strings.ReplaceAll(version, "-", ".")
	version = strings.ReplaceAll(version, "_", ".")

	// Split by dots
	parts := strings.Split(version, ".")
	result := make([]int, 0)

	for _, part := range parts {
		// Try to parse as integer
		if num, err := strconv.Atoi(part); err == nil {
			result = append(result, num)
		} else {
			// Handle alpha, beta, rc, etc.
			// For simplicity, treat non-numeric as 0
			result = append(result, 0)
		}
	}

	return result
}

func compareVersionParts(parts1, parts2 []int) int {
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		v1 := 0
		v2 := 0

		if i < len(parts1) {
			v1 = parts1[i]
		}
		if i < len(parts2) {
			v2 = parts2[i]
		}

		if v1 < v2 {
			return -1
		}
		if v1 > v2 {
			return 1
		}
	}

	return 0
}

func builtinPhpversion(args ...runtime.Value) runtime.Value {
	// Return a version that indicates PHP 8.0 compatibility
	// This is the version phpgo emulates
	return runtime.NewString("8.0.0")
}

func builtinExtensionLoaded(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	extension := strings.ToLower(args[0].ToString())

	// List of built-in extensions that phpgo supports
	supportedExtensions := map[string]bool{
		"json":       true,
		"pcre":       true,
		"hash":       true,
		"reflection": true,
		"spl":        true,
		"standard":   true,
		"core":       true,
		"date":       true,
		"filter":     true,
	}

	if supportedExtensions[extension] {
		return runtime.TRUE
	}

	return runtime.FALSE
}

func builtinMemoryGetUsage(args ...runtime.Value) runtime.Value {
	var m goruntime.MemStats
	goruntime.ReadMemStats(&m)
	// Return allocated memory in bytes
	return runtime.NewInt(int64(m.Alloc))
}

func builtinMemoryGetPeakUsage(args ...runtime.Value) runtime.Value {
	var m goruntime.MemStats
	goruntime.ReadMemStats(&m)
	// Return peak memory usage in bytes
	return runtime.NewInt(int64(m.TotalAlloc))
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

func (i *Interpreter) builtinClassAlias(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}

	originalName := args[0].ToString()
	aliasName := args[1].ToString()

	// Get the original class
	class, ok := i.env.GetClass(originalName)
	if !ok {
		return runtime.FALSE
	}

	// Create the alias by defining the same class with the new name
	i.env.DefineClass(aliasName, class)
	return runtime.TRUE
}

func (i *Interpreter) builtinSplAutoloadRegister(args ...runtime.Value) runtime.Value {
	// If no callback provided, use default autoload
	if len(args) == 0 {
		return runtime.TRUE
	}

	callback := args[0]

	// Verify the callback is callable
	if _, ok := callback.(*runtime.Function); !ok {
		// Could also be a string referring to a function name
		if str, ok := callback.(*runtime.String); ok {
			_, exists := i.env.GetFunction(str.Value)
			if !exists {
				return runtime.FALSE
			}
		} else {
			return runtime.FALSE
		}
	}

	// Register the autoload function
	i.autoloadFuncs = append(i.autoloadFuncs, callback)
	return runtime.TRUE
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

func builtinGetenv(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	varName := args[0].ToString()
	value := os.Getenv(varName)
	if value == "" {
		// Check if variable exists but is empty vs doesn't exist
		if _, exists := os.LookupEnv(varName); !exists {
			return runtime.FALSE
		}
	}
	return runtime.NewString(value)
}

func builtinPutenv(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	setting := args[0].ToString()

	// Parse "NAME=value" format
	parts := strings.SplitN(setting, "=", 2)
	if len(parts) != 2 {
		return runtime.FALSE
	}

	err := os.Setenv(parts[0], parts[1])
	if err != nil {
		return runtime.FALSE
	}
	return runtime.TRUE
}

func builtinParseIniFile(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	filename := args[0].ToString()

	// Read file contents
	content, err := os.ReadFile(filename)
	if err != nil {
		return runtime.FALSE
	}

	return parseIniString(string(content))
}

func builtinParseIniString(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	content := args[0].ToString()
	return parseIniString(content)
}

func parseIniString(content string) runtime.Value {
	result := runtime.NewArray()
	lines := strings.Split(content, "\n")
	var currentSection runtime.Value
	currentSection = nil

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for section header [section]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sectionName := strings.TrimSpace(line[1 : len(line)-1])
			currentSection = runtime.NewString(sectionName)
			result.Set(currentSection, runtime.NewArray())
			continue
		}

		// Parse key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		if currentSection != nil {
			// Add to section
			section, ok := result.Elements[currentSection].(*runtime.Array)
			if ok {
				section.Set(runtime.NewString(key), runtime.NewString(value))
			}
		} else {
			// Add to root level
			result.Set(runtime.NewString(key), runtime.NewString(value))
		}
	}

	return result
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

func builtinStrftime(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}

	format := args[0].ToString()
	var t time.Time

	if len(args) >= 2 {
		timestamp := args[1].ToInt()
		t = time.Unix(timestamp, 0)
	} else {
		t = time.Now()
	}

	return runtime.NewString(formatStrftime(format, t))
}

func builtinGmstrftime(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}

	format := args[0].ToString()
	var t time.Time

	if len(args) >= 2 {
		timestamp := args[1].ToInt()
		t = time.Unix(timestamp, 0).UTC()
	} else {
		t = time.Now().UTC()
	}

	return runtime.NewString(formatStrftime(format, t))
}

func formatStrftime(format string, t time.Time) string {
	var result strings.Builder
	i := 0
	for i < len(format) {
		if format[i] == '%' && i+1 < len(format) {
			switch format[i+1] {
			case 'a': // Abbreviated weekday name
				result.WriteString(t.Weekday().String()[:3])
			case 'A': // Full weekday name
				result.WriteString(t.Weekday().String())
			case 'b', 'h': // Abbreviated month name
				result.WriteString(t.Month().String()[:3])
			case 'B': // Full month name
				result.WriteString(t.Month().String())
			case 'C': // Century (year/100)
				result.WriteString(fmt.Sprintf("%02d", t.Year()/100))
			case 'd': // Day of month (01-31)
				result.WriteString(fmt.Sprintf("%02d", t.Day()))
			case 'e': // Day of month ( 1-31)
				result.WriteString(fmt.Sprintf("%2d", t.Day()))
			case 'H': // Hour 24-hour (00-23)
				result.WriteString(fmt.Sprintf("%02d", t.Hour()))
			case 'I': // Hour 12-hour (01-12)
				h := t.Hour() % 12
				if h == 0 {
					h = 12
				}
				result.WriteString(fmt.Sprintf("%02d", h))
			case 'j': // Day of year (001-366)
				result.WriteString(fmt.Sprintf("%03d", t.YearDay()))
			case 'm': // Month (01-12)
				result.WriteString(fmt.Sprintf("%02d", t.Month()))
			case 'M': // Minute (00-59)
				result.WriteString(fmt.Sprintf("%02d", t.Minute()))
			case 'p': // AM or PM
				if t.Hour() < 12 {
					result.WriteString("AM")
				} else {
					result.WriteString("PM")
				}
			case 'S': // Second (00-59)
				result.WriteString(fmt.Sprintf("%02d", t.Second()))
			case 'w': // Weekday (0-6, Sunday = 0)
				result.WriteString(fmt.Sprintf("%d", t.Weekday()))
			case 'y': // Year without century (00-99)
				result.WriteString(fmt.Sprintf("%02d", t.Year()%100))
			case 'Y': // Year with century
				result.WriteString(fmt.Sprintf("%04d", t.Year()))
			case 'Z': // Timezone name
				result.WriteString(t.Location().String())
			case 'z': // Timezone offset +hhmm
				_, offset := t.Zone()
				sign := "+"
				if offset < 0 {
					sign = "-"
					offset = -offset
				}
				hours := offset / 3600
				minutes := (offset % 3600) / 60
				result.WriteString(fmt.Sprintf("%s%02d%02d", sign, hours, minutes))
			case '%': // Literal %
				result.WriteByte('%')
			default:
				// Unknown specifier, output as-is
				result.WriteByte('%')
				result.WriteByte(format[i+1])
			}
			i += 2
		} else {
			result.WriteByte(format[i])
			i++
		}
	}
	return result.String()
}

func builtinGmdate(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}

	format := args[0].ToString()
	var timestamp int64

	if len(args) >= 2 {
		timestamp = args[1].ToInt()
	} else {
		timestamp = time.Now().Unix()
	}

	t := time.Unix(timestamp, 0).UTC()
	return runtime.NewString(convertPHPDateFormat(format, t))
}

func builtinGmmktime(args ...runtime.Value) runtime.Value {
	// Get the arguments (hour, minute, second, month, day, year)
	hour, minute, second := 0, 0, 0
	month, day, year := 1, 1, 1970

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

	t := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
	return runtime.NewInt(t.Unix())
}

func builtinGetdate(args ...runtime.Value) runtime.Value {
	var timestamp int64
	if len(args) >= 1 {
		timestamp = args[0].ToInt()
	} else {
		timestamp = time.Now().Unix()
	}

	t := time.Unix(timestamp, 0)
	result := runtime.NewArray()

	result.Set(runtime.NewString("seconds"), runtime.NewInt(int64(t.Second())))
	result.Set(runtime.NewString("minutes"), runtime.NewInt(int64(t.Minute())))
	result.Set(runtime.NewString("hours"), runtime.NewInt(int64(t.Hour())))
	result.Set(runtime.NewString("mday"), runtime.NewInt(int64(t.Day())))
	result.Set(runtime.NewString("wday"), runtime.NewInt(int64(t.Weekday())))
	result.Set(runtime.NewString("mon"), runtime.NewInt(int64(t.Month())))
	result.Set(runtime.NewString("year"), runtime.NewInt(int64(t.Year())))
	result.Set(runtime.NewString("yday"), runtime.NewInt(int64(t.YearDay()-1)))
	result.Set(runtime.NewString("weekday"), runtime.NewString(t.Weekday().String()))
	result.Set(runtime.NewString("month"), runtime.NewString(t.Month().String()))
	result.Set(runtime.NewInt(0), runtime.NewInt(timestamp))

	return result
}

func builtinCheckdate(args ...runtime.Value) runtime.Value {
	if len(args) < 3 {
		return runtime.FALSE
	}

	month := int(args[0].ToInt())
	day := int(args[1].ToInt())
	year := int(args[2].ToInt())

	// Check if month is valid (1-12)
	if month < 1 || month > 12 {
		return runtime.FALSE
	}

	// Check if year is valid (1-32767)
	if year < 1 || year > 32767 {
		return runtime.FALSE
	}

	// Check if day is valid for the given month/year
	daysInMonth := []int{0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

	// Check for leap year
	isLeap := (year%4 == 0 && year%100 != 0) || (year%400 == 0)
	if isLeap && month == 2 {
		daysInMonth[2] = 29
	}

	if day < 1 || day > daysInMonth[month] {
		return runtime.FALSE
	}

	return runtime.TRUE
}

func builtinIdate(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewInt(0)
	}

	format := args[0].ToString()
	var timestamp int64

	if len(args) >= 2 {
		timestamp = args[1].ToInt()
	} else {
		timestamp = time.Now().Unix()
	}

	t := time.Unix(timestamp, 0)

	if len(format) == 0 {
		return runtime.NewInt(0)
	}

	switch format[0] {
	case 'B': // Swatch Internet time
		return runtime.NewInt(0) // Not implemented
	case 'd': // Day of the month
		return runtime.NewInt(int64(t.Day()))
	case 'h': // Hour (12-hour format)
		h := t.Hour() % 12
		if h == 0 {
			h = 12
		}
		return runtime.NewInt(int64(h))
	case 'H': // Hour (24-hour format)
		return runtime.NewInt(int64(t.Hour()))
	case 'i': // Minutes
		return runtime.NewInt(int64(t.Minute()))
	case 'I': // 1 if DST, 0 otherwise
		return runtime.NewInt(0) // Not fully implemented
	case 'L': // 1 if leap year, 0 otherwise
		year := t.Year()
		isLeap := (year%4 == 0 && year%100 != 0) || (year%400 == 0)
		if isLeap {
			return runtime.NewInt(1)
		}
		return runtime.NewInt(0)
	case 'm': // Month number
		return runtime.NewInt(int64(t.Month()))
	case 's': // Seconds
		return runtime.NewInt(int64(t.Second()))
	case 't': // Days in current month
		year, month := t.Year(), t.Month()
		daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
		return runtime.NewInt(int64(daysInMonth))
	case 'U': // Unix timestamp
		return runtime.NewInt(timestamp)
	case 'w': // Day of the week (0=Sunday)
		return runtime.NewInt(int64(t.Weekday()))
	case 'W': // ISO-8601 week number
		_, week := t.ISOWeek()
		return runtime.NewInt(int64(week))
	case 'y': // Year (2 digits)
		return runtime.NewInt(int64(t.Year() % 100))
	case 'Y': // Year (4 digits)
		return runtime.NewInt(int64(t.Year()))
	case 'z': // Day of the year
		return runtime.NewInt(int64(t.YearDay() - 1))
	case 'Z': // Timezone offset in seconds
		_, offset := t.Zone()
		return runtime.NewInt(int64(offset))
	default:
		return runtime.NewInt(0)
	}
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

func builtinCrc32(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewInt(0)
	}
	data := []byte(args[0].ToString())
	checksum := crc32.ChecksumIEEE(data)
	return runtime.NewInt(int64(checksum))
}

func builtinHashHmac(args ...runtime.Value) runtime.Value {
	if len(args) < 3 {
		return runtime.FALSE
	}
	algo := strings.ToLower(args[0].ToString())
	data := []byte(args[1].ToString())
	key := []byte(args[2].ToString())

	var h []byte
	switch algo {
	case "md5":
		mac := hmac.New(md5.New, key)
		mac.Write(data)
		h = mac.Sum(nil)
	case "sha1":
		mac := hmac.New(sha1.New, key)
		mac.Write(data)
		h = mac.Sum(nil)
	case "sha256":
		mac := hmac.New(sha256.New, key)
		mac.Write(data)
		h = mac.Sum(nil)
	default:
		return runtime.FALSE
	}

	return runtime.NewString(hex.EncodeToString(h))
}

func builtinHashEquals(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	known := args[0].ToString()
	user := args[1].ToString()

	// Timing-safe comparison
	if len(known) != len(user) {
		return runtime.FALSE
	}

	result := 0
	for i := 0; i < len(known); i++ {
		result |= int(known[i]) ^ int(user[i])
	}

	return runtime.NewBool(result == 0)
}

func builtinPasswordHash(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	password := args[0].ToString()

	// Simple bcrypt-like hash using SHA-256 (not actual bcrypt for simplicity)
	// In production, use golang.org/x/crypto/bcrypt
	hash := sha256.Sum256([]byte(password))
	// Add a simple salt prefix (this is NOT secure, just for demo)
	result := "$2y$10$" + hex.EncodeToString(hash[:])

	return runtime.NewString(result)
}

func builtinPasswordVerify(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	password := args[0].ToString()
	hash := args[1].ToString()

	// Simple verification - hash the password and compare
	computed := sha256.Sum256([]byte(password))
	expected := "$2y$10$" + hex.EncodeToString(computed[:])

	return runtime.NewBool(expected == hash)
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

func builtinBin2hex(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	data := []byte(args[0].ToString())
	return runtime.NewString(hex.EncodeToString(data))
}

func builtinHex2bin(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}
	hexStr := args[0].ToString()
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return runtime.FALSE
	}
	return runtime.NewString(string(data))
}

func builtinQuotedPrintableEncode(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	str := args[0].ToString()
	var buf bytes.Buffer
	writer := quotedprintable.NewWriter(&buf)
	_, err := writer.Write([]byte(str))
	if err != nil {
		return runtime.FALSE
	}
	writer.Close()
	return runtime.NewString(buf.String())
}

func builtinQuotedPrintableDecode(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	str := args[0].ToString()
	reader := quotedprintable.NewReader(strings.NewReader(str))
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return runtime.FALSE
	}
	return runtime.NewString(string(decoded))
}

func builtinConvertUuencode(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	data := []byte(args[0].ToString())
	var result bytes.Buffer

	// Process data in chunks of 45 bytes
	for i := 0; i < len(data); i += 45 {
		end := i + 45
		if end > len(data) {
			end = len(data)
		}
		chunk := data[i:end]

		// Write length character
		result.WriteByte(byte(len(chunk) + 32))

		// Encode chunk
		for j := 0; j < len(chunk); j += 3 {
			var b1, b2, b3 byte
			b1 = chunk[j]
			if j+1 < len(chunk) {
				b2 = chunk[j+1]
			}
			if j+2 < len(chunk) {
				b3 = chunk[j+2]
			}

			result.WriteByte(byte(((b1 >> 2) & 0x3F) + 32))
			result.WriteByte(byte((((b1 << 4) | (b2 >> 4)) & 0x3F) + 32))
			result.WriteByte(byte((((b2 << 2) | (b3 >> 6)) & 0x3F) + 32))
			result.WriteByte(byte((b3 & 0x3F) + 32))
		}
		result.WriteByte('\n')
	}

	// Add terminator
	result.WriteByte('`')
	result.WriteByte('\n')

	return runtime.NewString(result.String())
}

func builtinConvertUudecode(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	data := args[0].ToString()
	lines := strings.Split(data, "\n")
	var result bytes.Buffer

	for _, line := range lines {
		if len(line) == 0 || line[0] == '`' {
			continue
		}

		// Get length
		length := int(line[0]) - 32
		if length <= 0 || length > 45 {
			continue
		}

		// Decode line
		encoded := line[1:]
		var decoded []byte

		for i := 0; i < len(encoded); i += 4 {
			if i+3 >= len(encoded) {
				break
			}

			c1 := byte(encoded[i]) - 32
			c2 := byte(encoded[i+1]) - 32
			c3 := byte(encoded[i+2]) - 32
			c4 := byte(encoded[i+3]) - 32

			b1 := (c1 << 2) | (c2 >> 4)
			b2 := (c2 << 4) | (c3 >> 2)
			b3 := (c3 << 6) | c4

			decoded = append(decoded, b1)
			if len(decoded) < length {
				decoded = append(decoded, b2)
			}
			if len(decoded) < length {
				decoded = append(decoded, b3)
			}

			if len(decoded) >= length {
				break
			}
		}

		result.Write(decoded[:length])
	}

	return runtime.NewString(result.String())
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

func builtinHtmlspecialcharsDecode(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	s := args[0].ToString()
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#039;", "'")
	s = strings.ReplaceAll(s, "&#39;", "'")
	return runtime.NewString(s)
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

func builtinArrayFillKeys(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	keys, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}
	value := args[1]

	result := runtime.NewArray()
	for _, key := range keys.Keys {
		keyVal := keys.Elements[key]
		result.Set(keyVal, value)
	}
	return result
}

func builtinArrayReplace(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewArray()
	}

	// Start with first array
	result := runtime.NewArray()
	if arr, ok := args[0].(*runtime.Array); ok {
		for _, key := range arr.Keys {
			result.Set(key, arr.Elements[key])
		}
	}

	// Replace values from subsequent arrays
	for i := 1; i < len(args); i++ {
		if arr, ok := args[i].(*runtime.Array); ok {
			for _, key := range arr.Keys {
				result.Set(key, arr.Elements[key])
			}
		}
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

func builtinArrayDiffKey(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewArray()
	}
	arr1, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewArray()
	}

	// Collect keys from all other arrays
	excludeKeys := make(map[string]bool)
	for i := 1; i < len(args); i++ {
		if arr, ok := args[i].(*runtime.Array); ok {
			for _, key := range arr.Keys {
				excludeKeys[key.ToString()] = true
			}
		}
	}

	result := runtime.NewArray()
	for _, key := range arr1.Keys {
		if !excludeKeys[key.ToString()] {
			result.Set(key, arr1.Elements[key])
		}
	}
	return result
}

func builtinArrayIntersectKey(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewArray()
	}
	arr1, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewArray()
	}

	// Collect keys that exist in ALL arrays
	keyCounts := make(map[string]int)
	numArrays := len(args)

	for i := 0; i < numArrays; i++ {
		if arr, ok := args[i].(*runtime.Array); ok {
			seen := make(map[string]bool)
			for _, key := range arr.Keys {
				keyStr := key.ToString()
				if !seen[keyStr] {
					seen[keyStr] = true
					keyCounts[keyStr]++
				}
			}
		}
	}

	result := runtime.NewArray()
	for _, key := range arr1.Keys {
		if keyCounts[key.ToString()] == numArrays {
			result.Set(key, arr1.Elements[key])
		}
	}
	return result
}

func builtinArrayDiffAssoc(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewArray()
	}
	arr1, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewArray()
	}

	// Collect key-value pairs from all other arrays
	excludePairs := make(map[string]string)
	for i := 1; i < len(args); i++ {
		if arr, ok := args[i].(*runtime.Array); ok {
			for _, key := range arr.Keys {
				keyStr := key.ToString()
				valStr := arr.Elements[key].ToString()
				excludePairs[keyStr] = valStr
			}
		}
	}

	result := runtime.NewArray()
	for _, key := range arr1.Keys {
		keyStr := key.ToString()
		valStr := arr1.Elements[key].ToString()
		// Include if key doesn't exist in other arrays OR if value is different
		if excludeVal, exists := excludePairs[keyStr]; !exists || excludeVal != valStr {
			result.Set(key, arr1.Elements[key])
		}
	}
	return result
}

func builtinArrayIntersectAssoc(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewArray()
	}
	arr1, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewArray()
	}

	// Collect key-value pairs that exist in ALL arrays
	pairCounts := make(map[string]int)
	numArrays := len(args)

	for i := 0; i < numArrays; i++ {
		if arr, ok := args[i].(*runtime.Array); ok {
			seen := make(map[string]bool)
			for _, key := range arr.Keys {
				keyStr := key.ToString()
				valStr := arr.Elements[key].ToString()
				pairKey := keyStr + "\x00" + valStr // Use null byte as separator
				if !seen[pairKey] {
					seen[pairKey] = true
					pairCounts[pairKey]++
				}
			}
		}
	}

	result := runtime.NewArray()
	for _, key := range arr1.Keys {
		keyStr := key.ToString()
		valStr := arr1.Elements[key].ToString()
		pairKey := keyStr + "\x00" + valStr
		if pairCounts[pairKey] == numArrays {
			result.Set(key, arr1.Elements[key])
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

func builtinDeg2rad(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewFloat(0)
	}
	degrees := args[0].ToFloat()
	return runtime.NewFloat(degrees * math.Pi / 180.0)
}

func builtinRad2deg(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewFloat(0)
	}
	radians := args[0].ToFloat()
	return runtime.NewFloat(radians * 180.0 / math.Pi)
}

func builtinHypot(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewFloat(0)
	}
	x := args[0].ToFloat()
	y := args[1].ToFloat()
	return runtime.NewFloat(math.Hypot(x, y))
}

func builtinLog10(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewFloat(0)
	}
	return runtime.NewFloat(math.Log10(args[0].ToFloat()))
}

func builtinLog1p(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewFloat(0)
	}
	return runtime.NewFloat(math.Log1p(args[0].ToFloat()))
}

func builtinExpm1(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewFloat(0)
	}
	return runtime.NewFloat(math.Expm1(args[0].ToFloat()))
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

func builtinSplObjectHash(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	obj, ok := args[0].(*runtime.Object)
	if !ok {
		return runtime.FALSE
	}

	// Generate a unique hash for the object using its memory address
	// In Go, we can use fmt.Sprintf with %p to get the pointer address
	hash := fmt.Sprintf("%p", obj)
	// Remove the "0x" prefix to get a clean hex string
	hash = strings.TrimPrefix(hash, "0x")
	// Pad to 32 characters to match PHP's format
	hash = fmt.Sprintf("%032s", hash)

	return runtime.NewString(hash)
}

func builtinSplObjectId(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	obj, ok := args[0].(*runtime.Object)
	if !ok {
		return runtime.FALSE
	}

	// Generate a unique integer ID for the object
	// We can use the pointer address as the ID
	id := fmt.Sprintf("%p", obj)
	// Remove 0x prefix and convert hex to int
	id = strings.TrimPrefix(id, "0x")
	var intId int64
	fmt.Sscanf(id, "%x", &intId)

	return runtime.NewInt(intId)
}

func (i *Interpreter) builtinGetObjectVars(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NULL
	}

	obj, ok := args[0].(*runtime.Object)
	if !ok {
		return runtime.NULL
	}

	result := runtime.NewArray()

	// Add all instance properties that are accessible
	for name, value := range obj.Properties {
		result.Set(runtime.NewString(name), value)
	}

	return result
}

func (i *Interpreter) builtinGetClassVars(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NULL
	}

	className := args[0].ToString()

	// Look up the class in the environment
	class, ok := i.env.GetClass(className)
	if !ok {
		return runtime.NULL
	}

	result := runtime.NewArray()

	// Return the default values of class properties
	for name, prop := range class.Properties {
		// Only include properties with default values
		if prop.Default != nil {
			result.Set(runtime.NewString(name), prop.Default)
		} else {
			result.Set(runtime.NewString(name), runtime.NULL)
		}
	}

	return result
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

func builtinMbStrlen(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewInt(0)
	}
	str := args[0].ToString()
	// Count runes (Unicode characters) instead of bytes
	return runtime.NewInt(int64(len([]rune(str))))
}

func builtinMbSubstr(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewString("")
	}
	str := args[0].ToString()
	start := args[1].ToInt()

	runes := []rune(str)
	length := int64(len(runes))

	// Handle negative start
	if start < 0 {
		start = length + start
		if start < 0 {
			start = 0
		}
	}

	if start >= length {
		return runtime.NewString("")
	}

	// Handle length parameter
	if len(args) >= 3 {
		subLen := args[2].ToInt()
		if subLen < 0 {
			// Negative length: stop at that position from end
			end := length + subLen
			if end <= start {
				return runtime.NewString("")
			}
			return runtime.NewString(string(runes[start:end]))
		}
		end := start + subLen
		if end > length {
			end = length
		}
		return runtime.NewString(string(runes[start:end]))
	}

	return runtime.NewString(string(runes[start:]))
}

func builtinMbStrpos(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}
	haystack := args[0].ToString()
	needle := args[1].ToString()
	offset := int64(0)

	if len(args) >= 3 {
		offset = args[2].ToInt()
	}

	haystackRunes := []rune(haystack)
	needleRunes := []rune(needle)

	if offset < 0 {
		offset = 0
	}

	if offset >= int64(len(haystackRunes)) {
		return runtime.FALSE
	}

	// Search for needle in haystack starting from offset
	for i := int(offset); i <= len(haystackRunes)-len(needleRunes); i++ {
		found := true
		for j := 0; j < len(needleRunes); j++ {
			if haystackRunes[i+j] != needleRunes[j] {
				found = false
				break
			}
		}
		if found {
			return runtime.NewInt(int64(i))
		}
	}

	return runtime.FALSE
}

func builtinMbStrtoupper(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	str := args[0].ToString()
	return runtime.NewString(strings.ToUpper(str))
}

func builtinMbStrtolower(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewString("")
	}
	str := args[0].ToString()
	return runtime.NewString(strings.ToLower(str))
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

func builtinStrpbrk(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.FALSE
	}

	haystack := args[0].ToString()
	charList := args[1].ToString()

	// Find first occurrence of any character from char_list
	for i, ch := range haystack {
		if strings.ContainsRune(charList, ch) {
			return runtime.NewString(haystack[i:])
		}
	}

	return runtime.FALSE
}

func builtinSimilarText(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewInt(0)
	}

	str1 := args[0].ToString()
	str2 := args[1].ToString()

	// Calculate similarity using longest common subsequence algorithm
	similarity := calculateSimilarity(str1, str2)

	return runtime.NewInt(int64(similarity))
}

func calculateSimilarity(str1, str2 string) int {
	len1, len2 := len(str1), len(str2)
	if len1 == 0 || len2 == 0 {
		return 0
	}

	// Simple similarity: count matching characters
	var sum int
	maxLen := len1
	if len2 > maxLen {
		maxLen = len2
	}

	for i := 0; i < maxLen && i < len1 && i < len2; i++ {
		if str1[i] == str2[i] {
			sum++
		}
	}

	return sum
}

func builtinSoundex(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	str := args[0].ToString()
	if len(str) == 0 {
		return runtime.FALSE
	}

	// Convert to uppercase and get first letter
	str = strings.ToUpper(str)
	var result strings.Builder

	// Find first letter
	var firstLetter rune
	for _, ch := range str {
		if ch >= 'A' && ch <= 'Z' {
			firstLetter = ch
			result.WriteRune(ch)
			break
		}
	}

	if firstLetter == 0 {
		return runtime.FALSE
	}

	// Soundex mapping
	soundexMap := map[rune]rune{
		'B': '1', 'F': '1', 'P': '1', 'V': '1',
		'C': '2', 'G': '2', 'J': '2', 'K': '2', 'Q': '2', 'S': '2', 'X': '2', 'Z': '2',
		'D': '3', 'T': '3',
		'L': '4',
		'M': '5', 'N': '5',
		'R': '6',
	}

	prevCode := soundexMap[firstLetter]

	for _, ch := range str {
		if ch == firstLetter {
			continue
		}
		if code, ok := soundexMap[ch]; ok {
			if code != prevCode {
				result.WriteRune(code)
				prevCode = code
				if result.Len() >= 4 {
					break
				}
			}
		} else {
			// A, E, I, O, U, H, W, Y reset the previous code
			if ch == 'A' || ch == 'E' || ch == 'I' || ch == 'O' || ch == 'U' || ch == 'H' || ch == 'W' || ch == 'Y' {
				prevCode = 0
			}
		}
	}

	// Pad with zeros
	for result.Len() < 4 {
		result.WriteRune('0')
	}

	return runtime.NewString(result.String()[:4])
}

func builtinLevenshtein(args ...runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewInt(-1)
	}

	str1 := args[0].ToString()
	str2 := args[1].ToString()

	// Levenshtein distance algorithm
	len1, len2 := len(str1), len(str2)

	// Create matrix
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}

	// Initialize first row and column
	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if str1[i-1] != str2[j-1] {
				cost = 1
			}

			delete := matrix[i-1][j] + 1
			insert := matrix[i][j-1] + 1
			substitute := matrix[i-1][j-1] + cost

			min := delete
			if insert < min {
				min = insert
			}
			if substitute < min {
				min = substitute
			}

			matrix[i][j] = min
		}
	}

	return runtime.NewInt(int64(matrix[len1][len2]))
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

func builtinArrayMultisort(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.FALSE
	}

	// For simplicity, implement basic single-array sorting
	// Full implementation would handle multiple arrays and sort order flags

	// Sort by value (ascending by default)
	type kvPair struct {
		key runtime.Value
		val runtime.Value
	}

	pairs := make([]kvPair, 0, len(arr.Keys))
	for _, key := range arr.Keys {
		pairs = append(pairs, kvPair{key, arr.Elements[key]})
	}

	// Sort pairs by value
	sort.SliceStable(pairs, func(i, j int) bool {
		vi := pairs[i].val
		vj := pairs[j].val

		// Compare based on type
		switch v1 := vi.(type) {
		case *runtime.Int:
			if v2, ok := vj.(*runtime.Int); ok {
				return v1.Value < v2.Value
			}
		case *runtime.Float:
			if v2, ok := vj.(*runtime.Float); ok {
				return v1.Value < v2.Value
			}
		case *runtime.String:
			if v2, ok := vj.(*runtime.String); ok {
				return v1.Value < v2.Value
			}
		}
		return false
	})

	// Rebuild array with sorted values (reindex)
	newKeys := make([]runtime.Value, 0, len(pairs))
	newElements := make(map[runtime.Value]runtime.Value)

	for i, pair := range pairs {
		newKey := runtime.NewInt(int64(i))
		newKeys = append(newKeys, newKey)
		newElements[newKey] = pair.val
	}

	arr.Keys = newKeys
	arr.Elements = newElements
	arr.NextIndex = int64(len(pairs))

	return runtime.TRUE
}

func builtinArrayChangeKeyCase(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewArray()
	}

	arr, ok := args[0].(*runtime.Array)
	if !ok {
		return runtime.NewArray()
	}

	// Default to CASE_LOWER (0)
	caseType := int64(0)
	if len(args) >= 2 {
		caseType = args[1].ToInt()
	}

	result := runtime.NewArray()

	for _, key := range arr.Keys {
		value := arr.Elements[key]
		keyStr := key.ToString()

		// Change case based on type
		var newKeyStr string
		if caseType == 1 { // CASE_UPPER
			newKeyStr = strings.ToUpper(keyStr)
		} else { // CASE_LOWER (0)
			newKeyStr = strings.ToLower(keyStr)
		}

		result.Set(runtime.NewString(newKeyStr), value)
	}

	return result
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

func builtinSysGetTempDir(args ...runtime.Value) runtime.Value {
	return runtime.NewString(os.TempDir())
}

func builtinTempnam(args ...runtime.Value) runtime.Value {
	dir := os.TempDir()
	prefix := "php"

	if len(args) >= 1 {
		dir = args[0].ToString()
	}
	if len(args) >= 2 {
		prefix = args[1].ToString()
	}

	// Create a temporary file
	file, err := os.CreateTemp(dir, prefix)
	if err != nil {
		return runtime.FALSE
	}

	filename := file.Name()
	file.Close()

	return runtime.NewString(filename)
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

func (i *Interpreter) builtinOpendir(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	path := args[0].ToString()
	file, err := os.Open(path)
	if err != nil {
		return runtime.FALSE
	}

	// Check if it's a directory
	info, err := file.Stat()
	if err != nil || !info.IsDir() {
		file.Close()
		return runtime.FALSE
	}

	// Create resource
	res := &runtime.Resource{
		ResType: "dir",
		Handle:  file,
		ID:      i.nextResourceID,
	}
	i.nextResourceID++
	i.resources[res.ID] = res

	return res
}

func builtinReaddir(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	res, ok := args[0].(*runtime.Resource)
	if !ok || res.ResType != "dir" {
		return runtime.FALSE
	}

	file, ok := res.Handle.(*os.File)
	if !ok {
		return runtime.FALSE
	}

	// Read one entry
	entries, err := file.Readdir(1)
	if err != nil || len(entries) == 0 {
		return runtime.FALSE
	}

	return runtime.NewString(entries[0].Name())
}

func builtinClosedir(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	res, ok := args[0].(*runtime.Resource)
	if !ok || res.ResType != "dir" {
		return runtime.FALSE
	}

	if file, ok := res.Handle.(*os.File); ok {
		file.Close()
		return runtime.TRUE
	}

	return runtime.FALSE
}

func builtinDiskFreeSpace(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	// This is platform-specific and complex to implement properly
	// For now, return a placeholder value
	// In a real implementation, we'd use syscall to get actual disk stats
	return runtime.NewInt(1000000000) // 1GB placeholder
}

func builtinDiskTotalSpace(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	// This is platform-specific and complex to implement properly
	// For now, return a placeholder value
	return runtime.NewInt(10000000000) // 10GB placeholder
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

func (i *Interpreter) builtinGetDefinedVars(args ...runtime.Value) runtime.Value {
	result := runtime.NewArray()

	// Get all variables from the current environment
	vars := i.env.GetAllVariables()
	for name, value := range vars {
		result.Set(runtime.NewString(name), value)
	}

	return result
}

func (i *Interpreter) builtinGetDefinedConstants(args ...runtime.Value) runtime.Value {
	result := runtime.NewArray()

	// Get all constants from the environment
	constants := i.env.GetAllConstants()
	for name, value := range constants {
		result.Set(runtime.NewString(name), value)
	}

	return result
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

// ----------------------------------------------------------------------------
// Network functions

func builtinIp2long(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	ipStr := args[0].ToString()

	// Parse IP address
	parts := strings.Split(ipStr, ".")
	if len(parts) != 4 {
		return runtime.FALSE
	}

	var result int64
	for i, part := range parts {
		val, err := strconv.Atoi(part)
		if err != nil || val < 0 || val > 255 {
			return runtime.FALSE
		}
		result += int64(val) << uint(8*(3-i))
	}

	return runtime.NewInt(result)
}

func builtinLong2ip(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	num := args[0].ToInt()

	// Convert long to IP address
	octet1 := (num >> 24) & 0xFF
	octet2 := (num >> 16) & 0xFF
	octet3 := (num >> 8) & 0xFF
	octet4 := num & 0xFF

	ipStr := fmt.Sprintf("%d.%d.%d.%d", octet1, octet2, octet3, octet4)
	return runtime.NewString(ipStr)
}

func builtinGethostbyname(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	hostname := args[0].ToString()

	// Look up IP addresses for the hostname
	addrs, err := net.LookupHost(hostname)
	if err != nil || len(addrs) == 0 {
		// Return the hostname itself if lookup fails (PHP behavior)
		return runtime.NewString(hostname)
	}

	// Return the first IP address
	return runtime.NewString(addrs[0])
}

func builtinGethostbyaddr(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	ipAddr := args[0].ToString()

	// Look up hostnames for the IP address
	names, err := net.LookupAddr(ipAddr)
	if err != nil || len(names) == 0 {
		// Return the IP address itself if lookup fails (PHP behavior)
		return runtime.NewString(ipAddr)
	}

	// Return the first hostname (remove trailing dot if present)
	hostname := names[0]
	if strings.HasSuffix(hostname, ".") {
		hostname = hostname[:len(hostname)-1]
	}
	return runtime.NewString(hostname)
}

func builtinInetPton(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	address := args[0].ToString()
	ip := net.ParseIP(address)
	if ip == nil {
		return runtime.FALSE
	}

	// Convert to binary representation
	// For IPv4, use the 4-byte representation
	if ipv4 := ip.To4(); ipv4 != nil {
		return runtime.NewString(string(ipv4))
	}

	// For IPv6, use the 16-byte representation
	return runtime.NewString(string(ip.To16()))
}

func builtinInetNtop(args ...runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.FALSE
	}

	in := args[0].ToString()
	inBytes := []byte(in)

	// Check length to determine if IPv4 or IPv6
	if len(inBytes) == 4 {
		// IPv4
		ip := net.IP(inBytes)
		return runtime.NewString(ip.String())
	} else if len(inBytes) == 16 {
		// IPv6
		ip := net.IP(inBytes)
		return runtime.NewString(ip.String())
	}

	return runtime.FALSE
}
