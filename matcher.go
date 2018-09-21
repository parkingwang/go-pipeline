package gopl

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"net/url"
	"strings"
)

// 匹配器
type Matcher interface {
	// 返回消息对象是否匹配
	Match(pack *DataFrame) bool
}

////

type AnyMatcher struct {
	Matcher
}

func (m *AnyMatcher) Match(pack *DataFrame) bool {
	return true
}

func (AnyMatcher) String() string {
	return "AnyMatcher"
}

//

type NoneMatcher struct {
	Matcher
}

func (*NoneMatcher) Match(pack *DataFrame) bool {
	return true
}

func (NoneMatcher) String() string {
	return "NoneMatcher"
}

////

type DefaultURLMatcher struct {
	Matcher

	spec  string
	path  string
	query url.Values
}

func (slf DefaultURLMatcher) String() string {
	return "DefaultURLMatcher[:" + slf.spec + "]"
}

func (slf *DefaultURLMatcher) Match(pack *DataFrame) bool {
	traceEnable := Debugs().RoutingTrace
	// Topic的Path部分，是匹配两个Topic是否匹配的第一标准
	if slf.path == pack.Topic() {
		// 然后以Matcher为标准，比较Header参数是否匹配
		for baseKey, baseVal := range slf.query {
			if traceEnable {
				log.Debug().Msgf("Matcher topic OK, testing header: %s == %s", baseKey, baseVal)
			}
			if headerVal, hit := pack.headers[baseKey]; !hit {
				if traceEnable {
					log.Debug().Msgf("Matcher.header NOT-MATCH, require: %s", baseKey)
				}
				return false
			} else {
				match := baseVal[0] == headerVal
				if !match && traceEnable {
					log.Debug().Msgf("Matcher.header value NOT-MATCH, accept: %s, was: %s", baseVal[0], headerVal)
				}
				return match
			}
		}
		return true
	} else {
		if traceEnable {
			log.Debug().Msgf("Matcher.topic NOT-MATCH, accept: %s, was: %s", slf.path, pack.Topic())
		}
		return false
	}
}

func NewDefaultURLMatcher(spec string) (Matcher, error) {
	if parsed, err := url.Parse(spec); nil != err {
		return nil, errors.WithMessage(err, "Default Matcher only accept http URL spec")
	} else {
		path := parsed.String()
		//  2:  /a?
		if i := strings.Index(path, "?"); i >= 2 {
			path = path[:i]
		}
		return &DefaultURLMatcher{
			spec:  spec,
			query: parsed.Query(),
			path:  path,
		}, nil
	}
}
