package heybox

import (
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// HumanBehavior 人类行为模拟
type HumanBehavior struct {
	page *rod.Page
}

// NewHumanBehavior 创建行为模拟实例
func NewHumanBehavior(page *rod.Page) *HumanBehavior {
	return &HumanBehavior{page: page}
}

// RandomDelay 随机延迟（模拟人类思考时间）
func (h *HumanBehavior) RandomDelay(minMs, maxMs int) {
	delay := rand.Intn(maxMs-minMs) + minMs
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

// ShortDelay 短延迟（100-300ms）
func (h *HumanBehavior) ShortDelay() {
	h.RandomDelay(100, 300)
}

// MediumDelay 中等延迟（300-800ms）
func (h *HumanBehavior) MediumDelay() {
	h.RandomDelay(300, 800)
}

// LongDelay 长延迟（800-2000ms）
func (h *HumanBehavior) LongDelay() {
	h.RandomDelay(800, 2000)
}

// HumanMouseMove 模拟人类鼠标移动（带轨迹）
func (h *HumanBehavior) HumanMouseMove(targetX, targetY float64) error {
	// 使用 MoveLinear 模拟轨迹
	steps := 10 + rand.Intn(10)
	err := h.page.Mouse.MoveLinear(proto.Point{X: targetX, Y: targetY}, steps)
	if err != nil {
		return err
	}

	// 随机微小延迟
	h.RandomDelay(10, 30)
	return nil
}

// HumanClick 模拟人类点击
func (h *HumanBehavior) HumanClick(selector string) error {
	// 先找到元素
	el, err := h.page.Element(selector)
	if err != nil {
		return err
	}

	// 获取元素形状
	shape, err := el.Shape()
	if err != nil {
		return err
	}

	// 计算点击位置（元素中心 + 随机偏移）
	targetX := shape.Box().X + shape.Box().Width/2 + float64(rand.Intn(10)-5)
	targetY := shape.Box().Y + shape.Box().Height/2 + float64(rand.Intn(10)-5)

	// 模拟鼠标移动
	if err := h.HumanMouseMove(targetX, targetY); err != nil {
		return err
	}

	// 短暂停顿（人类点击前的准备）
	h.ShortDelay()

	// 按下鼠标
	if err := h.page.Mouse.Down(proto.InputMouseButtonLeft, 1); err != nil {
		return err
	}

	// 随机按住时间（50-150ms）
	h.RandomDelay(50, 150)

	// 释放鼠标
	if err := h.page.Mouse.Up(proto.InputMouseButtonLeft, 1); err != nil {
		return err
	}

	// 点击后延迟
	h.MediumDelay()

	return nil
}

// HumanScroll 模拟人类滚动
func (h *HumanBehavior) HumanScroll(distance int) error {
	// 分多次滚动，每次随机距离
	totalScrolled := 0
	direction := 1
	if distance < 0 {
		direction = -1
		distance = -distance
	}

	for totalScrolled < distance {
		// 随机滚动距离（50-150像素）
		scrollAmount := 50 + rand.Intn(100)
		if totalScrolled+scrollAmount > distance {
			scrollAmount = distance - totalScrolled
		}

		// 滚动
		if err := h.page.Mouse.Scroll(0, float64(scrollAmount*direction), 1); err != nil {
			return err
		}

		totalScrolled += scrollAmount

		// 随机延迟
		h.RandomDelay(100, 300)
	}

	return nil
}

// HumanType 模拟人类输入（简化版，直接输入整个字符串）
func (h *HumanBehavior) HumanType(text string) error {
	// 随机输入延迟
	h.RandomDelay(100, 300)

	// 使用页面的 InsertText 方法
	_, err := h.page.Eval(`() => document.execCommand('insertText', false, ` + "`" + text + "`" + `)`)
	return err
}

// SimulateReading 模拟阅读行为（滚动 + 停顿）
func (h *HumanBehavior) SimulateReading() {
	// 随机滚动一些
	h.HumanScroll(100 + rand.Intn(200))

	// 模拟阅读时间
	h.RandomDelay(1000, 3000)

	// 再滚动一些
	h.HumanScroll(50 + rand.Intn(150))
}

// SimulatePageInteraction 模拟页面交互（进入页面后的自然行为）
func (h *HumanBehavior) SimulatePageInteraction() {
	// 初始等待（页面加载后的自然停顿）
	h.LongDelay()

	// 随机移动鼠标到页面某处
	randomX := float64(rand.Intn(800) + 100)
	randomY := float64(rand.Intn(400) + 100)
	h.HumanMouseMove(randomX, randomY)

	// 模拟浏览
	h.SimulateReading()

	// 随机再移动一次
	h.HumanMouseMove(float64(rand.Intn(800)+100), float64(rand.Intn(400)+100))
	h.MediumDelay()
}
