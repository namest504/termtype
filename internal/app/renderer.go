package app

import "github.com/gdamore/tcell/v2"

// Renderer는 화면 그리기를 담당하는 헬퍼 구조체입니다.
type Renderer struct {
	screen tcell.Screen
}

// NewRenderer는 새로운 Renderer 인스턴스를 생성합니다.
func NewRenderer(s tcell.Screen) *Renderer {
	return &Renderer{screen: s}
}

// DrawText는 화면에 텍스트를 그립니다.
func (r *Renderer) DrawText(x, y int, style tcell.Style, text string) {
	for i, r := range []rune(text) {
		r.screen.SetContent(x+i, y, r, nil, style)
	}
}

// Clear는 화면을 지웁니다.
func (r *Renderer) Clear() {
	r.screen.Clear()
}

// ShowCursor는 커서를 표시합니다.
func (r *Renderer) ShowCursor(x, y int) {
	r.screen.ShowCursor(x, y)
}

// HideCursor는 커서를 숨깁니다.
func (r *Renderer) HideCursor() {
	r.screen.HideCursor()
}

// Show는 화면을 업데이트합니다.
func (r *Renderer) Show() {
	r.screen.Show()
}

// Size는 화면 크기를 반환합니다.
func (r *Renderer) Size() (int, int) {
	return r.screen.Size()
}

// SetContent는 화면의 특정 위치에 문자를 설정합니다.
func (r *Renderer) SetContent(x, y int, r rune, style tcell.Style) {
	r.screen.SetContent(x, y, r, nil, style)
}
