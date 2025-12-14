package keepass

import (
	"strings"
	"time"

	"github.com/tobischo/gokeepasslib/v3"
	"github.com/tobischo/gokeepasslib/v3/wrappers"
)

const (
	KP_TITLE    = "Title"
	KP_USERNAME = "Username"
	KP_PASSWORD = "Password"
	KP_NOTES    = "Notes"
	KP_URL      = "URL"
	KP_PATH     = "BEARZ_PATH"
)

type Entry struct {
	*gokeepasslib.Entry
	parent *Group
}

func NewEntry() *Entry {
	inner := gokeepasslib.NewEntry()
	e := &Entry{
		&inner, nil,
	}

	e.SetPassword("")
	e.SetUsername("")
	e.SetTitle("")
	e.SetUrl("")
	e.SetNotes("")

	return e
}

func (e *Entry) SetUrl(url string) *Entry {
	return e.SetValue(KP_URL, url)
}

func (e *Entry) SetTitle(title string) *Entry {
	return e.SetValue(KP_TITLE, title)
}

func (e *Entry) SetUsername(username string) *Entry {
	return e.SetValue(KP_USERNAME, username)
}

func (e *Entry) SetPassword(password string) *Entry {
	return e.SetProtectedValue(KP_PASSWORD, password)
}

func (e *Entry) SetNotes(notes string) *Entry {
	return e.SetValue(KP_NOTES, notes)
}

func (e *Entry) SetPath(path string) *Entry {
	return e.SetValue(KP_PATH, path)
}

func (e *Entry) GetUrl() string {
	return e.GetContent(KP_URL)
}

func (e *Entry) GetUsername() string {
	return e.GetContent(KP_USERNAME)
}

func (e *Entry) GetPath() string {
	return e.GetContent(KP_PATH)
}

func (e *Entry) SetValue(key string, value string) *Entry {
	i := e.Entry.GetIndex(key)
	if i == -1 {
		e.Entry.Values = append(e.Entry.Values, gokeepasslib.ValueData{
			Key:   key,
			Value: gokeepasslib.V{Content: value},
		})
	} else {
		e.Entry.Values[i].Value = gokeepasslib.V{Content: value}
	}

	return e
}

func (e *Entry) SetProtectedValue(key string, value string) *Entry {
	i := e.Entry.GetIndex(key)
	if i == -1 {
		e.Entry.Values = append(e.Entry.Values, gokeepasslib.ValueData{
			Key:   key,
			Value: gokeepasslib.V{Content: value, Protected: wrappers.NewBoolWrapper(true)},
		})
	} else {
		e.Entry.Values[i].Value = gokeepasslib.V{Content: value, Protected: wrappers.NewBoolWrapper(true)}
	}

	return e
}

func (e *Entry) AddTag(name string) {
	tags := e.Tags()
	if _, ok := tags[name]; ok {
		return
	}

	e.Entry.Tags = e.Entry.Tags + "," + name
}

func (e *Entry) RemoveTag(name string) {
	tags := e.Tags()
	if _, ok := tags[name]; !ok {
		return
	}

	sb := make([]string, 0)
	for tag := range tags {
		if strings.EqualFold(tag, name) {
			continue
		}

		sb = append(sb, tag)
	}

	e.Entry.Tags = strings.Join(sb, ",")
}

func (e *Entry) Tags() map[string]*string {
	tags := make(map[string]*string)

	t := e.Entry.Tags
	tagNames := strings.Split(t, ",")
	for _, tagName := range tagNames {
		tagName = strings.TrimSpace(tagName)
		tags[tagName] = nil
	}

	return tags
}

func (e *Entry) ExpiresAt() *time.Time {
	if e.Times.ExpiryTime != nil {
		return &e.Times.ExpiryTime.Time
	}

	return nil
}

func (e *Entry) CreatedAt() *time.Time {
	if e.Times.CreationTime != nil {
		return &e.Times.CreationTime.Time
	}

	return nil
}

// key returns the full path of the entry, transversing
// the group tree to get the full path, exluding the root group
func (e *Entry) Key() string {
	name := e.GetPath()
	if name == "" {
		name = e.GetTitle()
	}

	groups := []*Group{}
	group := e.parent
	for group != nil {
		groups = append(groups, group)
		group = group.parent
	}

	// pop the root group
	groups = groups[:len(groups)-1]
	path := ""
	for _, group := range groups {
		if path != "" {
			path += "/"
		}

		path += group.Name
	}

	return path + "/" + name
}

// Gets the base64 value of the UUID
func (e *Entry) Version() *string {
	base64, err := e.UUID.MarshalText()
	if err != nil {
		return nil
	}

	uid := string(base64)
	return &uid
}
