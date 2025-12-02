package themes

import "termtype/internal/domain"

// Themes는 프로그램에 등록된 모든 테마를 저장하는 맵입니다.
var Themes = make(map[string]domain.Theme)
