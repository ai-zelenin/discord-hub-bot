package hub

import (
	"github.com/tidwall/gjson"
	"strings"
)

type Scope struct {
	Raw string
	Sub *Subscription
	R   interface{}
	Cfg *SourceConfig
}

func (s *Scope) JsonPath(p string) string {
	return gjson.Get(s.Raw, p).String()
}

func (s *Scope) MatchTokens(found, val string) bool {
	pathParts := strings.Split(found, " ")
	valParts := strings.Split(val, " ")
	for _, valPart := range valParts {
		var findValPart bool
		for _, pathPart := range pathParts {
			if strings.ToLower(valPart) == strings.ToLower(pathPart) {
				findValPart = true
				break
			}
		}
		if !findValPart {
			return false
		}
	}
	return true
}
