# 커스텀 테마 추가하기

이 문서는 `termtype` 타이핑 연습 프로그램에 자신만의 테마를 추가하는 방법을 안내합니다.

## 개요

프로그램은 실행 시 동적으로 테마를 로드합니다. 새로운 테마를 추가하기 위해 기존 코드를 수정할 필요가 거의 없습니다. `Theme` 인터페이스를 구현하고 `init()` 함수를 통해 자신을 등록하기만 하면 됩니다.

## 1단계: Go 파일 생성

먼저, 프로젝트 루트에 새로운 `.go` 파일을 생성합니다. 파일명은 테마의 이름과 맞추는 것이 좋습니다. (예: `my_theme.go`)

## 2단계: `Theme` 인터페이스 구현

생성한 파일에 새로운 구조체를 정의하고, 아래의 `Theme` 인터페이스를 구현해야 합니다.

```go
// Theme 인터페이스는 모든 테마가 구현해야 할 메서드를 정의합니다.
type Theme interface {
	// ResetState는 새 라운드를 위해 게임 상태를 초기화합니다.
	ResetState(*GameState)
	// UpdateScreen은 현재 게임 상태를 화면에 그립니다.
	UpdateScreen(tcell.Screen, *GameState)
	// OnTick은 실시간 업데이트가 필요할 때 호출됩니다 (예: 애니메이션).
	OnTick(*GameState)
}
```

- `ResetState`: 새로운 문제가 시작될 때 호출됩니다. 여기서 타이핑할 문장을 설정하고 테마에 필요한 상태를 초기화합니다.
- `UpdateScreen`: 화면을 다시 그려야 할 때마다(키 입력, 시간 경과 등) 호출됩니다. 화면의 모든 요소를 그리는 로직을 담당합니다.
- `OnTick`: 약 1초마다 호출됩니다. 애니메이션 효과처럼 시간에 따라 변하는 요소를 업데이트하는 데 사용됩니다.

## 3단계: `init()` 함수로 테마 등록

Go의 `init()` 함수를 사용하여, 프로그램 시작 시 당신의 테마를 전역 `Themes` 맵에 자동으로 등록해야 합니다. `init()` 함수 안에 `Themes["테마이름"] = &MyTheme{}` 코드를 추가하세요. "테마이름"은 `-theme` 플래그로 사용될 이름입니다.

## 예시: `RainbowTheme`

아래는 매 글자마다 색이 바뀌는 간단한 `RainbowTheme`의 전체 코드 예시입니다. 이 코드를 `rainbow_theme.go` 파일로 프로젝트에 추가하기만 하면 바로 `-theme=rainbow` 플래그로 사용할 수 있습니다.

```go
package main

import (
	"math/rand"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// 프로그램 시작 시 자동으로 테마를 등록합니다.
func init() {
	Themes["rainbow"] = &RainbowTheme{}
}

// RainbowTheme 구조체
type RainbowTheme struct{}

// ResetState 구현
func (t *RainbowTheme) ResetState(gs *GameState) {
	gs.resetCommon()
	gs.targetSentence = gs.sentences[rand.Intn(len(gs.sentences))]
}

// UpdateScreen 구현
func (t *RainbowTheme) UpdateScreen(s tcell.Screen, gs *GameState) {
	s.Clear()

	// 무지개 색상 배열
	colors := []tcell.Color{
		tcell.ColorRed,
		tcell.ColorOrange,
		tcell.ColorYellow,
		tcell.ColorGreen,
		tcell.ColorBlue,
		tcell.ColorIndigo,
		tcell.ColorViolet,
	}

	if !gs.isFinished {
		targetRunes := []rune(gs.targetSentence)
		for i, r := range targetRunes {
			style := tcell.StyleDefault.Foreground(colors[i%len(colors)])
			s.SetContent(i+1, 1, r, nil, style)
		}

		// 사용자 입력은 단색으로 표시
		drawText(s, 1, 3, tcell.StyleDefault, gs.userInput)

		cursorX := 1 + runewidth.StringWidth(gs.userInput)
		s.ShowCursor(cursorX, 3)
	} else {
		s.HideCursor()
		resultText := fmt.Sprintf("WPM: %.2f | Accuracy: %.2f%%", gs.wpm, gs.accuracy)
		drawText(s, 1, 1, tcell.StyleDefault, resultText)
	}

	s.Show()
}

// OnTick 구현 (이 테마는 애니메이션이 없으므로 비워둡니다)
func (t *RainbowTheme) OnTick(gs *GameState) {}

```
