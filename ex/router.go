package ex

import (
	"net/http"
	"strings"
)

// URL 패턴별로 실행할 핸들러들을 http 메서드별로 정의하였다.
type router struct {
	// 키: http 메서드
	// 값: URL 패턴별로 실행할 HandlerFunc
	handlers map[string]map[string]HandlerFunc
}

// 라우터에 핸들러를 등록하기 위한 메서드인 HandleFunc
// 매개변수로 전달된 http 메서드, URL 패턴, 핸들러 함수를 2차원 맵인 handlers 필드에 등록
func (r *router) HandleFunc(method, pattern string, h HandlerFunc) {
	// http 메서드로 등록된 맵이 있는지 확인
	m, ok := r.handlers[method]
	if !ok {
		// 등록된 맵이 없으면 새 맵을 생성
		m = make(map[string]HandlerFunc)
		r.handlers[method] = m
	}
	// http 메서드로 등록된 맵에 URL패턴과 핸들러 함수 등록
	m[pattern] = h
}

// 라우터에 등록된 동적 URL 패턴과 실제 URL 경로가 일치하는지 확인하는 match 함수
func match(pattern, path string) (bool, map[string]string) {
	// 패턴과 패스가 정확히 일치하면 바로 true 반환
	if pattern == path {
		return true, nil
	}

	// 패턴과 패스를 "/" 단위로 구분
	patterns := strings.Split(pattern, "/")
	paths := strings.Split(path, "/")

	// 패턴과 패스를 "/" 로 구분한 후 부분 문자열 집합의 개수가 다르면 false 반환
	if len(patterns) != len(paths) {
		return false, nil
	}

	// 패턴에 일치하는 URL 매개변수를 담기 위한 params 맵 생성
	params := make(map[string]string)

	// "/" 로 구분된 패턴/패스의 각 문자열을 하나씩 비교
	for i := 0; i < len(patterns); i++ {
		switch {
		case patterns[i] == paths[i]:
			// 패턴과 패스의 부분 문자열이 일치하면 바로 다음 루프 수행
		case len(patterns[i]) > 0 && patterns[i][0] == ':':
			// 패턴이 ':' 문자로 시작하면 params에 URL params를 담은 후 다음 루프 수행
			params[patterns[i][1:]] = paths[i]
		default:
			// 일치하는 경우가 없으면 false 반환
			return false, nil
		}
	}
	// true와 params를 반환
	return true, params
}

func (r *router) handler() HandlerFunc {
	return func(c *Context) {
		// http 메서드에 맞는 모든 handlers를 반복하여 요청 URL에 해당하는 handler를 찾음
		for pattern, handler := range r.handlers[c.Request.Method] {
			if ok, params := match(pattern, c.Request.URL.Path); ok {
				for k, v := range params {
					c.Params[k] = v
				}
				// 요청 URL에 해당하는 handler 수행
				handler(c)
				return
			}
		}
		// 요청 URL에 해당하는 handler를 찾지 못하면 NotFound 에러 처리
		http.NotFound(c.ResponseWriter, c.Request)
		return
	}
}
