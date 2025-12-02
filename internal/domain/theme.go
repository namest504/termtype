package domain

import "github.com/gdamore/tcell/v2"

// Renderer 인터페이스는 화면 그리기 기능을 추상화합니다.
type Renderer interface {
	DrawText(x, y int, style tcell.Style, text string)
	DrawRune(x, y int, runeVal rune, style tcell.Style) int
	Clear()
	Show()
	Size() (int, int)
	SetContent(x, y int, runeVal rune, style tcell.Style)
	HideCursor()
	ShowCursor(x, y int)
}

// Theme 인터페이스는 모든 테마가 구현해야 할 메서드를 정의합니다.
type Theme interface {
	// ResetState는 새 라운드를 위해 게임 상태를 초기화합니다.
	ResetState(*GameState)
	// UpdateScreen은 현재 게임 상태를 화면에 그립니다.
	UpdateScreen(Renderer, *GameState)
	// OnTick은 실시간 업데이트가 필요할 때 호출됩니다 (예: 애니메이션).
	OnTick(*GameState)
}
