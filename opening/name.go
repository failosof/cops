package opening

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

type Name [2]string

func ParseName(s string) (n Name) {
	parts := strings.Split(s, ":")
	n[0] = parts[0]
	if len(parts) > 1 {
		variation := strings.TrimSpace(parts[1])
		n[1] = strings.Split(variation, ",")[0]
	}
	return
}

func (n Name) Empty() bool {
	return len(n[0]) == 0 && len(n[1]) == 0
}

func (n Name) Family() string {
	return n[0]
}

func (n Name) Variation() string {
	return n[1]
}

func (n Name) String() string {
	var s strings.Builder
	s.WriteString(n[0])
	if len(n[1]) > 0 {
		s.WriteString(": ")
		s.WriteString(n[1])
	}
	return s.String()
}

func (n Name) FamilyTag() string {
	return sanitize(n[0])
}

func (n Name) VariationTag() string {
	return sanitize(n[1])
}

func (n Name) Tag() string {
	var tag strings.Builder
	tag.WriteString(n.FamilyTag())
	if len(n[1]) > 0 {
		tag.WriteRune('_')
		tag.WriteString(n.VariationTag())
	}
	return tag.String()
}

var nameSanitizer = strings.NewReplacer("'", "", " ", "_")

func sanitize(s string) string {
	s = removeDiacritics(s)
	s = nameSanitizer.Replace(s)
	return s
}

func removeDiacritics(s string) string {
	t := norm.NFD.String(s)
	var sb strings.Builder
	for _, r := range t {
		if !unicode.IsMark(r) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
