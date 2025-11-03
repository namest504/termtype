package app

import "github.com/gdamore/tcell/v2"

// Theme 인터페이스는 모든 테마가 구현해야 할 메서드를 정의합니다.
type Theme interface {
	// ResetState는 새 라운드를 위해 게임 상태를 초기화합니다.
	ResetState(*GameState)
	// UpdateScreen은 현재 게임 상태를 화면에 그립니다.
	UpdateScreen(*Renderer, *GameState)
	// OnTick은 실시간 업데이트가 필요할 때 호출됩니다 (예: 애니메이션).
	OnTick(*GameState)
}

// Themes는 프로그램에 등록된 모든 테마를 저장하는 맵입니다.
var Themes = make(map[string]Theme)
