package keepass

import "github.com/tobischo/gokeepasslib/v3"

type Group struct {
	*gokeepasslib.Group
	parent *Group
}

func NewGroup() *Group {
	inner := gokeepasslib.NewGroup()
	g := &Group{
		&inner, nil,
	}

	return g
}

func (g *Group) AddGroup(group *Group) {
	for _, nextGroup := range g.Groups {
		if group.UUID.Compare(nextGroup.UUID) {
			return
		}
	}

	g.Groups = append(g.Groups, *group.Group)
	group.parent = g
}

func (g *Group) RmGroup(group *Group) {

	for i, nextGroup := range g.Groups {
		if group.UUID.Compare(nextGroup.UUID) {
			g.Groups = append(g.Groups[:i], g.Groups[i+1:]...)
			group.parent = nil
			break
		}
	}
}

func (g *Group) AddEntry(entry *Entry) {
	for _, nextEntry := range g.Entries {
		if entry.UUID.Compare(nextEntry.UUID) {
			return
		}
	}
	g.Entries = append(g.Entries, *entry.Entry)
	entry.parent = g
}

func (g *Group) RmEntry(entry *Entry) {
	for i, nextEntry := range g.Entries {
		if entry.UUID.Compare(nextEntry.UUID) {
			g.Entries = append(g.Entries[:i], g.Entries[i+1:]...)
			if entry.parent != nil {
				entry.parent = nil
			}

			break
		}
	}
}
