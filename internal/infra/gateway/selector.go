package gateway

import (
	"errors"
	"net/http"
	"sort"
	"strings"
)

const (
	selectorHeaderServer = "X-Mcp-Server"
	selectorHeaderTags   = "X-Mcp-Tags"
)

var (
	ErrSelectorRequired = errors.New("selector required: use /server/{name} or /tags/{tag1,tag2}")
)

type Selector struct {
	Server string
	Tags   []string
}

func (s Selector) empty() bool {
	return strings.TrimSpace(s.Server) == "" && len(s.Tags) == 0
}

func (s Selector) normalized() Selector {
	s.Server = strings.TrimSpace(s.Server)
	s.Tags = normalizeTags(s.Tags)
	return s
}

func (s Selector) equal(other Selector) bool {
	if s.Server != other.Server {
		return false
	}
	if len(s.Tags) != len(other.Tags) {
		return false
	}
	for i := range s.Tags {
		if s.Tags[i] != other.Tags[i] {
			return false
		}
	}
	return true
}

func NormalizeTags(tags []string) []string {
	return normalizeTags(tags)
}

func SelectorKey(sel Selector) string {
	sel = sel.normalized()
	if sel.Server != "" {
		return "server:" + sel.Server
	}
	if len(sel.Tags) > 0 {
		return "tags:" + strings.Join(sel.Tags, ",")
	}
	return ""
}

func ParseSelector(r *http.Request, basePath string) (Selector, error) {
	pathSel, pathOK, err := parsePathSelector(r, basePath)
	if err != nil {
		return Selector{}, err
	}
	headerSel, headerOK, err := parseHeaderSelector(r)
	if err != nil {
		return Selector{}, err
	}
	if pathOK {
		pathSel = pathSel.normalized()
		if headerOK {
			headerSel = headerSel.normalized()
			if !pathSel.equal(headerSel) {
				return Selector{}, errors.New("selector conflict: url and header selectors differ")
			}
		}
		return pathSel, nil
	}
	if headerOK {
		return headerSel.normalized(), nil
	}
	return Selector{}, ErrSelectorRequired
}

func parsePathSelector(r *http.Request, basePath string) (Selector, bool, error) {
	relative, ok := stripBasePath(r.URL.Path, basePath)
	if !ok {
		return Selector{}, false, nil
	}
	relative = strings.Trim(relative, "/")
	if relative == "" {
		return Selector{}, false, nil
	}
	parts := strings.Split(relative, "/")
	if len(parts) != 2 {
		return Selector{}, false, errors.New("invalid selector path: expected /server/{name} or /tags/{tag1,tag2}")
	}
	kind := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	if kind == "server" {
		if value == "" {
			return Selector{}, false, errors.New("invalid selector path: server name is required")
		}
		return Selector{Server: value}, true, nil
	}
	if kind == "tags" {
		if value == "" {
			return Selector{}, false, errors.New("invalid selector path: tag list is required")
		}
		tags := normalizeTags(strings.Split(value, ","))
		if len(tags) == 0 {
			return Selector{}, false, errors.New("invalid selector path: tag list is required")
		}
		return Selector{Tags: tags}, true, nil
	}
	return Selector{}, false, errors.New("invalid selector path: expected /server/{name} or /tags/{tag1,tag2}")
}

func parseHeaderSelector(r *http.Request) (Selector, bool, error) {
	server := strings.TrimSpace(r.Header.Get(selectorHeaderServer))
	tagsHeader := strings.TrimSpace(r.Header.Get(selectorHeaderTags))
	if server != "" && tagsHeader != "" {
		return Selector{}, false, errors.New("invalid selector headers: X-Mcp-Server and X-Mcp-Tags are mutually exclusive")
	}
	if server != "" {
		return Selector{Server: server}, true, nil
	}
	if tagsHeader != "" {
		tags := normalizeTags(strings.Split(tagsHeader, ","))
		if len(tags) == 0 {
			return Selector{}, false, errors.New("invalid selector headers: tag list is required")
		}
		return Selector{Tags: tags}, true, nil
	}
	return Selector{}, false, nil
}

func stripBasePath(path, basePath string) (string, bool) {
	cleanBase := strings.TrimSpace(basePath)
	if cleanBase == "" {
		cleanBase = "/"
	}
	if cleanBase != "/" {
		cleanBase = strings.TrimRight(cleanBase, "/")
	}
	if cleanBase == "/" {
		if strings.HasPrefix(path, "/") {
			return path[1:], true
		}
		return path, true
	}
	if path == cleanBase {
		return "", true
	}
	prefix := cleanBase + "/"
	if strings.HasPrefix(path, prefix) {
		return path[len(prefix):], true
	}
	return "", false
}

func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	unique := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalized == "" {
			continue
		}
		unique[normalized] = struct{}{}
	}
	if len(unique) == 0 {
		return nil
	}
	result := make([]string, 0, len(unique))
	for tag := range unique {
		result = append(result, tag)
	}
	sort.Strings(result)
	return result
}
