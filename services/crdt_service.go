package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Step: ProseMirror 변경 하나
type Step struct {
	StepType string      `json:"stepType"` // "replace"
	From     int         `json:"from"`
	To       int         `json:"to"`
	Slice    interface{} `json:"slice"` // 새로 삽입할 node/문자열 JSON
}

// Change: 여러 Steps + 메타데이터
type Change struct {
	Type      string    `json:"type"`    // "note"
	Version   int       `json:"version"` // 클라이언트 버전
	Steps     []Step    `json:"steps"`
	ClientID  string    `json:"clientID"`
	Timestamp time.Time `json:"timestamp"`
}

// LWW CRDT
type CRDT struct {
	mu       sync.RWMutex
	Document map[string]interface{} // 문서(트리) 구조 (ProseMirror JSON)
	Version  int
}

// NewCRDT: CRDT 초기화
func NewCRDT() *CRDT {
	return &CRDT{
		Document: make(map[string]interface{}),
		Version:  0,
	}
}

// ApplyChange: Change.Version > c.Version 일 때만 Step 적용 (LWW)
func (c *CRDT) ApplyChange(change Change) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if change.Version <= c.Version {
		// 이미 적용된 버전 또는 낮은 버전이면 무시
		return nil
	}

	// 1) 문서를 문자열로 펼침
	docStr, err := docToText(c.Document)
	if err != nil {
		return fmt.Errorf("flatten doc failed: %v", err)
	}

	// 2) 각 Step에 대해 replace 적용
	for _, step := range change.Steps {
		switch step.StepType {
		case "replace":
			docStr, err = applyReplace(docStr, step)
			if err != nil {
				return fmt.Errorf("applyReplace error: %v", err)
			}
		default:
			// 그 외 StepType은 미구현
			fmt.Printf("[Warning] Unknown stepType: %s\n", step.StepType)
		}
	}

	// 3) 최종 문자열을 다시 JSON으로 역변환
	newDoc, err := textToDoc(docStr)
	if err != nil {
		return fmt.Errorf("failed to parse new doc: %v", err)
	}

	// 4) CRDT 상태 갱신
	c.Document = newDoc
	c.Version = change.Version
	return nil
}

// docToText: ProseMirror JSON → 단순 텍스트로 flatten
func docToText(doc map[string]interface{}) (string, error) {
	// 매우 단순화:
	// - doc["type"] == "doc", doc["content"] == [{...}, {...}]
	// - 노드 트리를 순회하며 text 노드를 연결
	sb := &strings.Builder{}
	err := traverseNode(doc, sb)
	return sb.String(), err
}

// traverseNode: 노드를 재귀적으로 돌며 textContent를 sb에 추가
func traverseNode(node interface{}, sb *strings.Builder) error {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case map[string]interface{}:
		// 예: { "type":"doc", "content":[...] }
		nodeType, _ := n["type"].(string)
		if nodeType == "text" {
			// text 노드 => attrs: { text }
			if txt, ok := n["text"].(string); ok {
				sb.WriteString(txt)
			}
		}
		// content가 있으면 순회
		if content, ok := n["content"].([]interface{}); ok {
			for _, c := range content {
				traverseNode(c, sb)
			}
		}
	}
	return nil
}

// textToDoc: 문자열을 다시 (아주 단순한) ProseMirror doc 형태로 변환
func textToDoc(text string) (map[string]interface{}, error) {
	// 여기선 모든 텍스트를 하나의 paragraph 안의 text 노드로 넣는 예시
	newDoc := map[string]interface{}{
		"type": "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": text,
					},
				},
			},
		},
	}
	return newDoc, nil
}

// applyReplace: docStr[from:to]를 sliceStr로 치환
func applyReplace(docStr string, step Step) (string, error) {
	from := step.From
	to := step.To
	if from < 0 || to < 0 || from > len(docStr) || to > len(docStr) || from > to {
		return "", fmt.Errorf("invalid from/to range: %d-%d", from, to)
	}

	// step.Slice를 flatten 문자열로 변환
	sliceText, err := sliceToText(step.Slice)
	if err != nil {
		return "", err
	}

	// 문자열 치환
	return docStr[:from] + sliceText + docStr[to:], nil
}

// sliceToText: step.Slice 내부의 ProseMirror JSON을 flatten
func sliceToText(slice interface{}) (string, error) {
	// slice 예: { "content":[{...}, {...}] }, "type":"paragraph" 등
	bytes, _ := json.Marshal(slice)
	var node map[string]interface{}
	if err := json.Unmarshal(bytes, &node); err != nil {
		return "", fmt.Errorf("sliceToText unmarshal fail: %v", err)
	}

	sb := &strings.Builder{}
	traverseNode(node, sb)
	return sb.String(), nil
}
