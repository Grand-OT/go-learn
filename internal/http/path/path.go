package path

import (
	"errors"
	"net/url"
	"strings"
)

var (
	ErrInvalidPattern = errors.New("invalid pattern")
	ErrInvalidPath    = errors.New("invalid path")
	ErrBadEncoding    = errors.New("invalid URL encoding")
	ErrEncodedSlash   = errors.New("encoded slash not allowed")
	ErrDuplicateParam = errors.New("duplicate param name")
)

// MatchPath("/todos/:id", "/todos/123") => ok=true, {"id":"123"}
func Match(pattern, path string) (ok bool, params map[string]string, err error) {

	if len(pattern) == 0 {
		return false, nil, ErrInvalidPattern
	}

	if pattern[0] != '/' {
		return false, nil, ErrInvalidPattern
	}

	if pattern[len(pattern)-1] == '/' && len(pattern) != 1 {
		return false, nil, ErrInvalidPattern
	}

	if len(path) == 0 {
		return false, nil, ErrInvalidPath
	}

	if path[0] != '/' {
		return false, nil, ErrInvalidPath
	}

	if path[len(path)-1] == '/' && len(path) != 1 {
		return false, nil, ErrInvalidPath
	}

	partsPat := strings.Split(pattern, "/")
	partsPath := strings.Split(path, "/")
	if len(partsPat) != len(partsPath) {
		return false, nil, nil
	}

	params = make(map[string]string)

	for i := 0; i < len(partsPath); i++ {

		decoded, err := url.PathUnescape(partsPath[i])
		if err != nil {
			return false, nil, ErrBadEncoding
		}

		// check for a part to be a parameter
		if partPat := partsPat[i]; len(partPat) != 0 && partPat[0] == ':' {
			key := partPat[1:]
			if len(key) == 0 {
				return false, nil, ErrInvalidPattern
			}
			_, ok = params[key]
			if ok {
				return false, nil, ErrDuplicateParam
			}
			if len(decoded) == 0 {
				return false, nil, nil
			}
			if strings.Contains(decoded, "/") {
				return false, nil, ErrEncodedSlash
			}
			params[key] = decoded
		} else {
			if partsPat[i] != decoded {
				return false, nil, nil
			}
		}
	}

	return true, params, nil
}
