package labeler

import (
	"strings"

	"ethClassify/internal/domain"
)

type StaticLabeler struct {
	labels map[string]string
}

func NewStaticLabeler(labels map[string]string) *StaticLabeler {
	normalized := make(map[string]string, len(labels))
	for addr, label := range labels {
		normalized[strings.ToLower(addr)] = label
	}
	return &StaticLabeler{labels: normalized}
}

func (l *StaticLabeler) Label(addr string) string {
	if l == nil || len(l.labels) == 0 {
		return ""
	}
	label, ok := l.labels[strings.ToLower(addr)]
	if !ok {
		return ""
	}
	return label
}

var _ domain.AddressLabeler = (*StaticLabeler)(nil)
